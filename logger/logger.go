package logger

import (
	"Go-web-server/setting"
	"github.com/gin-gonic/gin"
	"github.com/natefinch/lumberjack"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strings"
	"time"
)

var zapLog *zap.Logger

// Init InitLogger 初始化Logger
func Init(config *setting.LogConfig, mode string) (err error){
	// 用viper读取配置文件的参数
	lumberJackLogger := &lumberjack.Logger{
		Filename: config.FileName,
		MaxSize: config.MaxSize,
		MaxBackups: config.MaxBackups,
		MaxAge: config.MaxAge,
	}
	writeSyncer := zapcore.AddSync(lumberJackLogger)
	encoder := getEncoder()
	level := new(zapcore.Level)
	err = level.UnmarshalText([]byte(viper.GetString("log.level")))
	if err != nil{
		return
	}
	var core zapcore.Core
	if mode == "dev"{
		// 开发者模式 日志输出到终端
		core = zapcore.NewTee(
			zapcore.NewCore(encoder, writeSyncer, level),
			zapcore.NewCore(zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
				zapcore.Lock(os.Stdout),zapcore.DebugLevel),
				)
	}else{
		core = zapcore.NewCore(encoder, writeSyncer, level)
	}
	zapLog = zap.New(core, zap.AddCaller())
	zap.ReplaceGlobals(zapLog) // 替换zap包的全局logger实例，后续在其他包中只需使用zap.L()调用
	return
}

func getEncoder() zapcore.Encoder{
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.TimeKey = "time"
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.EncodeDuration = zapcore.SecondsDurationEncoder
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	return zapcore.NewJSONEncoder(encoderConfig)
}

// GinLogger 接收gin框架默认的日志
func GinLogger() gin.HandlerFunc{
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		c.Next()
		cost := time.Since(start)
		zapLog.Info(path,
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("path",path),
			zap.String("query",query),
			zap.String("ip",c.ClientIP()),
			zap.String("user-agent",c.Request.UserAgent()),
			zap.Duration("cost",cost),
			)
	}
}

// GinRecovery recover项目可能出现的panic，然后使用zap记录相关日志
func GinRecovery(stack bool) gin.HandlerFunc{
	return func(c *gin.Context){
		defer func() {
			if err := recover(); err != nil{
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok{
					if se, ok := ne.Err.(*os.SyscallError);ok{
						if strings.Contains(strings.ToLower(se.Error()),"broken pipe") ||
							strings.Contains(strings.ToLower(se.Error()), "connection reset by peer"){
							brokenPipe = true
						}
					}
				}
				httpRequest,_ := httputil.DumpRequest(c.Request, false)
				if brokenPipe{
					zapLog.Error(c.Request.URL.Path,
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
					)
					// 如果连接断了，就不写状态
					c.Error(err.(error))
					c.Abort()
					return
				}
				if stack{
					zapLog.Error("[Recovery from panic]",
						zap.Any("error",err),
						zap.String("request", string(httpRequest)),
						zap.String("stack", string(debug.Stack())),
						)
				}else{
					zapLog.Error("[Recovery from panic]",
						zap.Any("error",err),
						zap.String("request", string(httpRequest)),
					)
				}
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}