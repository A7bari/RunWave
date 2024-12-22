package coderunner

import (
	"fmt"

	"github.com/A7bari/RunWave/internal/taskqueue"
)

func FormatCommand(task taskqueue.Task) ([]string, error) {
	switch task.GetLanguage() {
	case "python":
		return []string{"python", "-c", task.GetCode()}, nil
	case "java":
		return []string{"java", "-c", task.GetCode()}, nil
	case "c":
		return []string{"gcc", "-o", "temp", "-x", "c", "-", task.GetCode()}, nil
	default:
		return nil, fmt.Errorf("unsupported language")
	}
}
