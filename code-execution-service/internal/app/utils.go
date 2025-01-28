package app

import "github.com/A7bari/RunWave/internal/config"

func IsSupportedLanguage(lang string) bool {
	langs := config.GetConfig().Languages

	for _, l := range langs {
		if l == lang {
			return true
		}
	}

	return false
}
