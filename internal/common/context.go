package common

import (
	"context"
	"strconv"
	"time"
)

const importIdCtxKey    = "impID"

func NewContext(i string) context.Context {
	impID := i
	if impID == "" {
		impID = strconv.FormatInt(time.Now().UnixNano(), 10)
	}
	return context.WithValue(context.Background(), importIdCtxKey, impID)
}


func GetImportID(ctx context.Context) string {
	if v := ctx.Value(importIdCtxKey); v != nil {
		return v.(string)
	}
	return ""
}