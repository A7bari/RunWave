package coderunner

import (
	"fmt"
)

func FormatCommand(lang, code string) ([]string, error) {
	switch lang {
	case "python":
		return []string{"python", "-c", code}, nil
	case "java":
		return []string{"java", "-c", code}, nil
	case "c":
		return []string{"gcc", "-o", "temp", "-x", "c", "-", code}, nil
	default:
		return nil, fmt.Errorf("unsupported language")
	}
}
