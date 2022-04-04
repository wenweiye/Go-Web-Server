package main

import (
	"Go-web-server/dao/mysql"
	"Go-web-server/dao/redis"
	"Go-web-server/logger"
	"Go-web-server/route"
	"Go-web-server/setting"
	"context"
	"flag"
	"fmt"
	"github.com/apex/log"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"net/http"
	"syscall"
	"time"

	"os"
	"os/signal"
)

func main(){
	filePath := ""
	flag.StringVar(&filePath, "file", "", "配置文件地址")
	flag.Parse()
	// 加载配置文件
	if err := setting.Init(filePath);err != nil{
		fmt.Printf("init settings failed:%s \n", err)
		return
	}
	// 初始化日志
	if err := logger.Init(setting.Config.LogConfig, setting.Config.Mode); err != nil{
		fmt.Printf("init settings failed:%s \n", err)
		return
	}
	zap.L().Debug("logger init success")
	defer zap.L().Sync()

	// 初始化mysql
	if err := mysql.Init(setting.Config.MysqlConfig); err != nil{
		fmt.Printf("init mysql failed:%s \n", err)
		return
	}
	zap.L().Debug("mysql init success")
	// 初始化redis
	if err := redis.Init(setting.Config.RedisConfig); err != nil{
		fmt.Printf("init redis failed:%s \n", err)
		return
	}

	defer mysql.Db.Close()
	defer redis.Rdb.Close()
	zap.L().Debug("redis init success")
	//注册路由
	r := route.Setup()

	srv := &http.Server{
		Addr: fmt.Sprintf(":%d", viper.GetInt("app.port")),
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed{
			zap.L().Error("listen: %s\n", zap.Error(err))
		}
	}()

	// 等待终端信号来优雅地关闭服务器，为关闭服务器操作设置一个5秒的超时
	quit := make(chan os.Signal, 1) // 创建一个接收信号的通道
	// kill 默认会发送 syscall.SIGTERM 信号
	// kill -2 发送 syscall.SIGINT 信号，我们常用的Ctrl+C就是触发系统SIGINT信号
	// kill -9 发送 syscall.SIGKILL 信号， 但是不能被捕获，所以不需要添加它
	// signal.Notify把收到的 syscall.SIGINT或syscall.SIGTERM 信号转发给quit
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM) //此处不会阻塞
	<- quit // 阻塞在此，当接收到上述信号时才会往下执行
	zap.L().Info("Shutdown server")
	// 创建一个5秒超时的context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// 5秒内优雅关闭服务（将未处理完的请求处理完再关闭服务），超过5秒就超时退出
	if err := srv.Shutdown(ctx);err != nil{
		log.Fatalf("Server Shutdown: ", err)
		zap.L().Error("Server Shutdown: ", zap.Error(err))
	}
	zap.L().Info("Server exiting")
}
