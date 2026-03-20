package main

import (
	"log"
	"virturalDevice/pkg/vds"
	"virturalDevice/pkg/vds/infrastructure/connection"

	"github.com/spf13/viper"
	"go.uber.org/fx"
)

func init() {
	viper.SetConfigName("vds")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("../config")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("读取配置文件失败, 将回退到默认参数: %+v", err)
		return
	}
}

func main() {
	app := fx.New(
		connection.Module(),
		vds.Module(),
	)
	app.Run()
}
