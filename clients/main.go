package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-openapi/loads/fmts"
	"github.com/go-openapi/spec"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	swaggerURL = "https://raw.githubusercontent.com/iron-io/functions/master/docs/swagger.yml"
	rootTmpDir = "tmp"
)

func main() {
	os.RemoveAll(rootTmpDir)
	cwd, _ := os.Getwd()

	// Download swagger yaml and convert to JSON
	d, err := fmts.YAMLDoc(swaggerURL)
	if err != nil {
		log.Fatalf("Failed to convert swagger yaml to json: %v", err)
	}

	var sw spec.Swagger
	if err := json.Unmarshal(d, &sw); err != nil {
		log.Fatalf("Failed to convert swagger yaml to json: %v", err)
	}

	version := sw.Info.Version
	fmt.Printf("VERSION: %s\n", version)

	var only string
	if len(os.Args) > 1 && os.Args[1] != "" {
		only = os.Args[1]
	}

	var languages []string
	if only != "" {
		languages = append(languages, only)
	} else {
		// Download available languages from swagger generator api
		languages = getLanguages()
	}

	for _, language := range languages {
		var skipFiles []string
		tmpDir := filepath.Join(rootTmpDir, language)
		srcDir := filepath.Join(tmpDir, "src")
		clientDir := filepath.Join(tmpDir, fmt.Sprintf("%s-client", language))

		var options map[string]interface{}
		var deploy [][]string

		// Specfic language configurations
		switch language {
		case "go":
			options["packageName"] = "functions"
			options["packageVersion"] = version
		case "ruby":
			skipFiles = append(skipFiles, "#{gem_name}.gemspec")
			deploy = append(deploy, []string{"gem", "build #{gem_name}.gemspec", "gem push #{gem_name}-#{version}.gem"})
			options["gemName"] = "iron_functions"
			options["moduleName"] = "IronFunctions"
			options["gemVersion"] = version
			options["gemHomepage"] = "https://github.com/iron-io/functions_ruby"
			options["gemSummary"] = "Ruby gem for IronFunctions"
			options["gemDescription"] = "Ruby gem for IronFunctions."
			options["gemAuthorEmail"] = "travis@iron.io"
		case "javascript":
			options["projectName"] = "iron_functions"
			deploy = append(deploy, []string{"npm", "publish"})
		default:
			continue
		}
		log.Printf("Generating `%s` client...\n", language)
		err = os.MkdirAll(tmpDir, 0777)
		if err != nil {
			log.Printf("Failed to create temporary directory for %s client. Skipping...", language)
		}

		// Generate client
		gen, err := generateClient(language, options)
		if err != nil {
			log.Printf("Failed to generated %s client. Skipping...", language)
			continue
		}

		// Download generated client
		log.Printf("Downloading `%s` client...\n", language)
		gf, err := getFile(strings.Replace(gen.Link, "https", "http", 1))
		if err != nil {
			log.Printf("Failed to download generated %s client. Skipping...", language)
		}
		ioutil.WriteFile(filepath.Join(tmpDir, "gen.zip"), gf, 0777)

		// Unzip
		log.Printf("Unzipping `%s` client...\n", language)
		exec.Command("unzip", "-o", filepath.Join(tmpDir, "gen.zip"), "-d", tmpDir).Run()
		os.Remove(filepath.Join(tmpDir, "gen.zip"))

		log.Printf("Cloning previous `%s` source...\n", language)
		exec.Command("git", "clone", fmt.Sprintf("git@github.com:iron-io/functions_%s.git", language), srcDir).Run()

		// Skip language specific files
		for _, skip := range skipFiles {
			os.Remove(filepath.Join(tmpDir, clientDir, skip))
		}

		// Copying new client
		log.Printf("Copying new `%s` client to src directory\n", language)

		// Only solution I found
		filepath.Walk(clientDir, func(path string, info os.FileInfo, err error) error {
			if path == clientDir {
				return nil
			}
			exec.Command("cp", "-r", path, srcDir).Run()
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		})

		os.Chdir(srcDir)
		exec.Command("git", "add", ".").Run()
		exec.Command("git", "commit", "-am", fmt.Sprintf("Updated to api version %s", version)).Run()

		log.Printf("Tagging new `%s` version as `%s`\n", language, version)
		r := exec.Command("git", "tag", "-a", "-m", fmt.Sprintf("Updated to api version %s", version), version).Run()
		if r != nil && r.Error() != "" {
			log.Println("Version already exists, bump swagger the version")
			os.Exit(-1)
		}

		log.Printf("Pushing new `%s` client\n", language)
		r = exec.Command("git", "push", "--follow-tags").Run()
		if r != nil && r.Error() != "" {
			log.Printf("Failed to push new version: %s\n", r.Error())
			os.Exit(-1)
		}

		log.Printf("Releasing new `%s` client\n", language)
		for _, d := range deploy {
			exec.Command(d[0], d[1])
		}

		log.Printf("Updated `%s` client to `%s` \n", language, version)

		os.Chdir(cwd)
		os.RemoveAll(tmpDir)
	}

}

type generatedClient struct {
	Link string `json:"link"`
}

func generateClient(lang string, options map[string]interface{}) (gc generatedClient, err error) {
	payload := map[string]interface{}{
		"swaggerUrl": swaggerURL,
		"options":    options,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return
	}

	resp, err := http.Post(fmt.Sprintf("http://generator.swagger.io/api/gen/clients/%s", lang), "application/json", bytes.NewBuffer(data))
	if err != nil {
		return
	}
	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	err = json.Unmarshal(result, &gc)
	if err != nil {
		return
	}

	return
}

func getFile(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func getLanguages() (langs []string) {
	data, err := getFile("http://generator.swagger.io/api/gen/clients")
	if err != nil {
		log.Fatalf("Failed to load swagger languages: %v", err)
		os.Exit(-1)
	}

	err = json.Unmarshal(data, &langs)
	if err != nil {
		log.Fatalf("Failed to load swagger languages: %v", err)
		os.Exit(-1)
	}

	return
}
