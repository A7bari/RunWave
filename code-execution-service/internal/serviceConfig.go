package internal

import (
	"strings"
	"time"
)

type ServiceConfig struct {
	Namespace     string
	Commands      map[string]string
	GetPodRetries int
	Standbylabel  string
	InUseLabel    string
	Timeout       time.Duration
}

func (e *ServiceConfig) AddCmd(lang string, cmd string) {
	e.Commands[lang] = cmd
}

func (e *ServiceConfig) GetCmd(lang string) ([]string, error) {
	cmd, exists := e.Commands[lang]
	if !exists || cmd == "" {
		return nil, NewErrorf(ErrorCodeUnsupportLanguage, "unsupported language: %s", lang)
	}
	return strings.Fields(cmd), nil
}
