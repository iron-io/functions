package langs

import (
	"fmt"
	"os"
	"os/exec"
)

type DotNetLangHelper struct {
}

func (lh *DotNetLangHelper) Entrypoint() string {
	return "dotnet func.dll"
}

func (lh *DotNetLangHelper) HasPreBuild() bool {
	return true
}

// PreBuild for Go builds the binary so the final image can be as small as possible
func (lh *DotNetLangHelper) PreBuild() error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	// todo: this won't work if the function is more complex since the import paths won't match up, need to fix
	cmd := exec.Command("docker", "run", "--rm", "-v", wd+":/dotnet", "-w", "/dotnet", "microsoft/dotnet", "dotnet restore")
	// cmd.Dir = dir
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error running docker build: %v", err)
	}
	return nil
}

func (lh *DotNetLangHelper) AfterBuild() error {
	return nil
}
