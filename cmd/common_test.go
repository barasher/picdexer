package cmd

import (
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"testing"
)



func TestSetLoggingLevel(t *testing.T) {
	var tcs = []struct {
		tcID   string
		preLvl  zerolog.Level
		inLvl string
		expSuccess bool
		expLvl zerolog.Level
	}{
		{"debug", zerolog.InfoLevel, "debug", true, zerolog.DebugLevel},
		{"info", zerolog.DebugLevel, "info", true, zerolog.InfoLevel},
		{"warn", zerolog.DebugLevel, "warn", true, zerolog.WarnLevel},
		{"undefined", zerolog.DebugLevel, "undefined", false, zerolog.WarnLevel},
		{"empty", zerolog.DebugLevel, "", true, zerolog.DebugLevel},

	}

	preTestLvl := zerolog.GlobalLevel()
	defer zerolog.SetGlobalLevel(preTestLvl)

	for _, tc := range tcs {
		t.Run(tc.tcID, func(t *testing.T) {
			zerolog.SetGlobalLevel(tc.preLvl)
			err := setLoggingLevel(tc.inLvl)
			if tc.expSuccess {
				assert.Nil(t, err)
				assert.Equal(t, tc.expLvl, zerolog.GlobalLevel())
			} else {
				assert.NotNil(t, err)
			}
			assert.Equal(t, tc.expSuccess, err == nil)
		})
	}
}
