package langs

import "fmt"

// GetLangHelper returns a LangHelper for the passed in language
func GetLangHelper(lang string) (LangHelper, error) {
	switch lang {
	case "dotnet":
		return &DotNetLangHelper{}, nil
	case "go":
		return &GoLangHelper{}, nil
	case "node":
		return &NodeLangHelper{}, nil
	case "php":
		return &PHPHelper{}, nil
	case "python":
		return &PythonHelper{}, nil
	case "ruby":
		return &RubyLangHelper{}, nil
	case "rust":
		return &RustLangHelper{}, nil
	}
	return nil, fmt.Errorf("No language helper found for %v", lang)
}

type LangHelper interface {
	Entrypoint() string
	HasPreBuild() bool
	PreBuild() error
	AfterBuild() error
}
