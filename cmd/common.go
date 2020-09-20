package cmd

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func setLoggingLevel(lvl string) error {
	if lvl != "" {
		lvl, err := zerolog.ParseLevel(lvl)
		if err != nil {
			return fmt.Errorf("error while setting logging level (%v): %w", lvl, err)
		}
		zerolog.SetGlobalLevel(lvl)
		log.Debug().Msgf("Logging level: %v", lvl)
	}
	return nil
}
