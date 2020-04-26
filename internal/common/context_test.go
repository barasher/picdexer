package common

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBuildContextWithProvidedID(t *testing.T) {
	ctx := NewContext("anID")
	v := ctx.Value(importIdCtxKey)
	assert.NotNil(t, v)
	assert.Equal(t, "anID", v.(string))
}

func TestBuildContextWithGeneratedID(t *testing.T) {
	ctx := NewContext("")
	v := ctx.Value(importIdCtxKey)
	assert.NotNil(t, v)
	assert.NotZero(t, v.(string))
}

func TestGetImportIDWithID(t *testing.T) {
	ctx := NewContext("anID")
	assert.Equal(t, "anID", GetImportID(ctx))
}

func TestGetImportIDWithoutID(t *testing.T) {
	ctx := context.Background()
	assert.Equal(t, "", GetImportID(ctx))
}
