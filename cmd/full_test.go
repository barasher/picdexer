package cmd

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func runNominal(context.Context, Config, []string) error {
	return nil
}

func runError(context.Context, Config, []string) error {
	return fmt.Errorf("run error")
}

func TestDoFull_Nominal(t *testing.T) {
	assert.Nil(t, doFull("../testdata/conf/picdexer_nominal.json", "", []string{}, runNominal))
}

func TestDoFull_FailOnWrongLoggingLevel(t *testing.T) {
	assert.NotNil(t, doFull("../testdata/conf/picdexer_wrongLoggingLevel.json", "", []string{}, runNominal))
}

func TestDoFull_FailOnConfLoad(t *testing.T) {
	assert.NotNil(t, doFull("nonExistingFile", "", []string{}, runNominal))
}

func TestDoFull_FailOnRun(t *testing.T) {
	assert.NotNil(t, doFull("../testdata/conf/picdexer_nominal.json", "", []string{}, runError))
}
