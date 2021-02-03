package cmd

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDoFull_Nominal(t *testing.T) {
	assert.Nil(t, doFull("../testdata/conf/picdexer_nominal.json", "", []string{}, simulateRun(true)))
}

func TestDoFull_FailOnWrongLoggingLevel(t *testing.T) {
	assert.NotNil(t, doFull("../testdata/conf/picdexer_wrongLoggingLevel.json", "", []string{}, simulateRun(true)))
}

func TestDoFull_FailOnConfLoad(t *testing.T) {
	assert.NotNil(t, doFull("nonExistingFile", "", []string{}, simulateRun(true)))
}

func TestDoFull_FailOnRun(t *testing.T) {
	assert.NotNil(t, doFull("../testdata/conf/picdexer_nominal.json", "", []string{}, simulateRun(false)))
}
