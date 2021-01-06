package cmd

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

type setupMock struct {
	doBuildFail    bool
	doSetupEsFail  bool
	doSetupKibFail bool
}

func (s setupMock) Build(string, string, string) (setupInterface, error) {
	if s.doBuildFail {
		return nil, fmt.Errorf("build error mocked")
	}
	return s, nil
}

func (s setupMock) SetupElasticsearch() error {
	if s.doSetupEsFail {
		return fmt.Errorf("es error mocked")
	}
	return nil
}

func (s setupMock) SetupKibana() error {
	if s.doSetupKibFail {
		return fmt.Errorf("kib error mocked")
	}
	return nil
}

func TestDoConfigure_Nominal(t *testing.T) {
	s := setupMock{}
	assert.Nil(t, doConfigure("../testdata/conf/picdexer_nominal.json", s.Build))
}

func TestDoConfigure_FailOnSetupElasticsearch(t *testing.T) {
	s := setupMock{doSetupEsFail: true}
	assert.NotNil(t, doConfigure("../testdata/conf/picdexer_nominal.json", s.Build))
}

func TestDoConfigure_FailOnSetupKibana(t *testing.T) {
	s := setupMock{doSetupKibFail: true}
	assert.NotNil(t, doConfigure("../testdata/conf/picdexer_nominal.json", s.Build))
}

func TestDoConfigure_FailOnConfLoad(t *testing.T) {
	s := setupMock{}
	assert.NotNil(t, doConfigure("nonExistingFile", s.Build))
}

func TestDoConfigure_FailOnBuild(t *testing.T) {
	s := setupMock{doBuildFail: true}
	assert.NotNil(t, doConfigure("../testdata/conf/picdexer_nominal.json", s.Build))
}

func TestDoConfigure_FailOnWrongLoggingLevel(t *testing.T) {
	s := setupMock{}
	assert.NotNil(t, doConfigure("../testdata/conf/picdexer_wrongLoggingLevel.json", s.Build))
}