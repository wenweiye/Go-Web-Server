package setting

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)


type AppConfig struct{
	Name string `mapstructure:"name"`
	Mode string `mapstructure:"mode"`
	Port int `mapstructure:"port"`
	*LogConfig `mapstructure:"log"`
	*MysqlConfig `mapstructure:"mysql"`
	*RedisConfig `mapstructure:"redis"`
}

type LogConfig struct{
	Level string `mapstructure:"level"`
	FileName string `mapstructure:"filename"`
	MaxSize int `mapstructure:"max_size"`
	MaxAge int `mapstructure:"max_age"`
	MaxBackups int `mapstructure:"max_backups"`
}

type MysqlConfig struct{
	Host              string `mapstructure:"host"`
	Port              int    `mapstructure:"port"`
	User              string `mapstructure:"user"`
	Password          string `mapstructure:"password"`
	DbName            string `mapstructure:"dbname"`
	MaxOpenConnection int    `mapstructure:"max_open_connection"`
	MaxIdleConnection int    `mapstructure:"max_idle_connection"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Password string `mapstructure:"passowrd"`
	Post     int    `mapstructure:"port"`
	Db       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

var Config = new(AppConfig)

// Init 加载配置文件
func Init(filePath string) error{
	if len(filePath) == 0{
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		// 配置可以有多个，并且需要告诉viper配置文件的路径
		viper.AddConfigPath(".")
	}else{
		viper.SetConfigFile(filePath)
	}

	err := viper.ReadInConfig()
	if err != nil{
		fmt.Println("viper init failed:", err)
		return err
	}
	// 序列化对象
	if err := viper.Unmarshal(Config); err != nil{
		fmt.Println("viper Unmarshal err", err)
	}
	viper.WatchConfig()
	viper.OnConfigChange(func(in fsnotify.Event){
		fmt.Println("配置文件已修改")
	})
	return err
}