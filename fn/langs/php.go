package langs

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type PHPHelper struct {
}

func (lh *PHPHelper) Entrypoint() string {
	return "php func.php"
}

func (lh *PHPHelper) HasPreBuild() bool {
	return true
}

// PreBuild for Go builds the binary so the final image can be as small as possible
func (lh *PHPHelper) PreBuild() error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	pbcmd := fmt.Sprintf("docker run --rm -v %s:/worker -w /worker iron/php:dev composer install", wd)
	fmt.Println("Running prebuild command:", pbcmd)
	parts := strings.Fields(pbcmd)
	head := parts[0]
	parts = parts[1:len(parts)]
	cmd := exec.Command(head, parts...)
	// cmd.Dir = dir
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error running docker build: %v", err)
	}
	return nil
}

func (lh *PHPHelper) AfterBuild() error {
	return nil
}
