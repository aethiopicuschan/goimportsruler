package ctxrw_test

import (
	"context"
	"testing"

	"github.com/aethiopicuschan/goimportsruler/pkg/config"
	"github.com/aethiopicuschan/goimportsruler/pkg/ctxrw"
	"github.com/stretchr/testify/assert"
)

func TestContextWriter_SetConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		ctx  context.Context
		cfg  config.Config
	}{
		{
			name: "stores config in context",
			ctx:  context.Background(),
			cfg:  *config.ExampleConfig(),
		},
		{
			name: "works with derived context",
			ctx:  context.WithValue(context.Background(), "unrelated", "value"),
			cfg:  *config.ExampleConfig(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			w := ctxrw.NewContextWriter()
			gotCtx := w.SetConfig(tc.ctx, tc.cfg)

			v := gotCtx.Value(ctxrw.ContextKeyConfig)
			if !assert.NotNil(t, v) {
				return
			}

			got, ok := v.(config.Config)
			assert.True(t, ok, "expected value type to be config.Config")
			assert.Equal(t, tc.cfg, got, "stored config should match input")
		})
	}
}

func TestContextReader_GetConfig(t *testing.T) {
	t.Parallel()

	example := *config.ExampleConfig()

	tests := []struct {
		name      string
		ctx       context.Context
		want      config.Config
		wantError bool
		errSubstr string
	}{
		{
			name: "returns config when present",
			ctx: func() context.Context {
				w := ctxrw.NewContextWriter()
				return w.SetConfig(context.Background(), example)
			}(),
			want: example,
		},
		{
			name:      "returns error when missing",
			ctx:       context.Background(),
			wantError: true,
			errSubstr: "config not found",
		},
		{
			name: "returns error when wrong type stored",
			ctx: func() context.Context {
				return context.WithValue(context.Background(), ctxrw.ContextKeyConfig, "not-a-config")
			}(),
			wantError: true,
			errSubstr: "type assert",
		},
		{
			name: "returns error when nil stored",
			ctx: func() context.Context {
				return context.WithValue(context.Background(), ctxrw.ContextKeyConfig, nil)
			}(),
			wantError: true,
			errSubstr: "config not found",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			r := ctxrw.NewContextReader()
			got, err := r.GetConfig(tc.ctx)

			if tc.wantError {
				if !assert.Error(t, err) {
					return
				}
				assert.Contains(t, err.Error(), tc.errSubstr)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestContextReader_Interface(t *testing.T) {
	t.Parallel()

	var _ ctxrw.IContextReader = (*ctxrw.ContextReader)(nil)
}

func TestContextWriter_Interface(t *testing.T) {
	t.Parallel()

	var _ ctxrw.IContextWriter = (*ctxrw.ContextWriter)(nil)
}
