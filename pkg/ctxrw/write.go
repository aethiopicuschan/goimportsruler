package ctxrw

import (
	"context"

	"github.com/aethiopicuschan/goimportsruler/pkg/config"
)

type IContextWriter interface {
	SetConfig(ctx context.Context, cfg config.Config) context.Context
}

type ContextWriter struct{}

func NewContextWriter() *ContextWriter {
	return &ContextWriter{}
}

func (r *ContextWriter) SetConfig(ctx context.Context, cfg config.Config) context.Context {
	return context.WithValue(ctx, ContextKeyConfig, cfg)
}
