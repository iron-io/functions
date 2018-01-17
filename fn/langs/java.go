package langs

import (
	"fmt"
	"os"
	"os/exec"
)

type JavaLangHelper struct {
	BaseHelper
}

func (lh *JavaLangHelper) Entrypoint() string {
	return "java -cp lib/*:. Func"
}

func (lh *JavaLangHelper) HasPreBuild() bool {
	return true
}

// PreBuild for Go builds the binary so the final image can be as small as possible
func (lh *JavaLangHelper) PreBuild() error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	cmd := exec.Command(
		"docker", "run",
		"--rm", "-v",
		wd+":/java", "-w", "/java", "iron/java:dev",
		"/bin/sh", "-c", "javac -cp lib/*:. *.java",
	)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error running docker build: %v", err)
	}
	return nil
}

func (lh *JavaLangHelper) AfterBuild() error {
	return nil
}
