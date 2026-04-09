package main

import (
	"flag"
	"fmt"
	"myreddit/controller"
	"myreddit/dao/postgres"
	"myreddit/dao/redis"
	"myreddit/logger"
	"myreddit/pkg/snowflake"
	"myreddit/routes"
	"myreddit/settings"

	"go.uber.org/zap"
)

// main 初始化配置、日志、存储与路由并启动 HTTP 服务。
func main() {
	var configPath string
	flag.StringVar(&configPath, "c", "conf/config.yaml", "配置文件路径")
	flag.Parse()
	if err := settings.Init(configPath); err != nil {
		fmt.Printf("init settings failed, err:%v\n", err)
		return
	}

	if err := logger.Init(settings.Conf.LogConfig, settings.Conf.Mode); err != nil {
		fmt.Printf("init logger failed, err:%v\n", err)
		return
	}
	defer func(l *zap.Logger) {
		err := l.Sync()
		if err != nil {
			fmt.Printf("sync logger failed, err:%v\n", err)
		}
	}(zap.L())
	zap.L().Debug("init logger success")

	if err := postgres.Init(settings.Conf.PostgresConfig); err != nil {
		fmt.Printf("init postgres failed, err:%v\n", err)
		return
	}
	defer postgres.Close()

	if err := redis.Init(settings.Conf.RedisConfig); err != nil {
		fmt.Printf("init redis failed, err:%v\n", err)
		return
	}
	defer redis.Close()

	if err := snowflake.Init(settings.Conf.StartTime, settings.Conf.MachineID); err != nil {
		fmt.Printf("init snowflake failed, err:%v\n", err)
		return
	}

	if err := controller.InitTrans("zh"); err != nil {
		fmt.Printf("init trans failed, err:%v\n", err)
		return
	}

	r := routes.SetupRouter(settings.Conf.Mode)

	err := r.Run(fmt.Sprintf(":%d", settings.Conf.Port))
	if err != nil {
		zap.L().Error("start server failed", zap.Error(err))
		fmt.Printf("start server failed, err:%v\n", err)
		return
	}

}
