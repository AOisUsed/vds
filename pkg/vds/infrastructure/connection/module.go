package connection

import (
	"context"
	"log"
	"virturalDevice/pkg/vds/domain/connection"

	"go.uber.org/fx"
)

func newConnection(kind string, lc fx.Lifecycle) connection.Connection {

	if kind == "mock" {
		conn := NewConn()
		lc.Append(fx.Hook{
			OnStart: func(context.Context) error {
				log.Println("正在启动 mock connection")
				err := conn.Serve()
				log.Println("mock connection 启动成功")
				return err
			},
			OnStop: func(context.Context) error {
				log.Println("正在关闭 mock connection")
				_ = conn.Close()
				log.Println("mock connection 关闭成功")
				return nil
			},
		})
		return conn
	} else if kind == "udp" {
		conn := NewUDPConn()
	}

}

func Module() fx.Option {
	return fx.Options(
		fx.Provide(newConnection),
	)
}
