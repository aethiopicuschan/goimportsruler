package ctxrw

import (
	"context"
	"fmt"

	"github.com/aethiopicuschan/goimportsruler/pkg/config"
)

type IContextReader interface {
	GetConfig(ctx context.Context) (config.Config, error)
}

type ContextReader struct{}

func NewContextReader() *ContextReader {
	return &ContextReader{}
}

func (r *ContextReader) GetConfig(ctx context.Context) (cfg config.Config, err error) {
	if v := ctx.Value(ContextKeyConfig); v == nil {
		err = fmt.Errorf("config not found in context")
		return
	}
	cfg, ok := ctx.Value(ContextKeyConfig).(config.Config)
	if !ok {
		err = fmt.Errorf("failed to type assert config from context")
		return
	}
	return
}
