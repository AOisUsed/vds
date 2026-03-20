package vds

import (
	"context"
	"log"
	"virturalDevice/pkg/vds/domain"
	"virturalDevice/pkg/vds/domain/codec"
	"virturalDevice/pkg/vds/domain/connection"
	"virturalDevice/pkg/vds/domain/repository"
	"virturalDevice/pkg/vds/domain/sender"

	"go.uber.org/fx"
)

func newVDS(
	conn connection.Connection,
	repo repository.VDRepository,
	sender sender.Sender,
	codec codec.Codec,
) *domain.VDS {
	return domain.NewVDS(conn, repo, sender, codec)
}

func Module() fx.Option {
	return fx.Options(
		fx.Provide(newVDS),
		fx.Invoke(func(vds *domain.VDS, lc fx.Lifecycle) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					log.Println("正在启动VDS服务...")
					vds.Start()
					log.Println("VDS服务已启动.")
					return nil
				},
				OnStop: func(ctx context.Context) error {
					log.Println("正在停止 VDS 服务...")
					vds.Stop()
					log.Println("VDS 服务已停止")
					return nil
				},
			})
		}),
	)
}
