package netconn

import (
	"context"

	"go.uber.org/fx"
)

func NewConnStore(lc fx.Lifecycle) *Store {
	store := NewStore()
	lc.Append(fx.Hook{
		OnStop: func(context.Context) error {
			return store.CloseAll()
		},
	})
	return store
}
