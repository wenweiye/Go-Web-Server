package mysql

import (
	"Go-web-server/setting"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var Db *sqlx.DB

func Init(config *setting.MysqlConfig) (err error){
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.DbName,
		)
	// 也可以使用MustConnect连接，不成功就panic
	Db, err = sqlx.Connect("mysql",dsn)
	if err != nil{
		zap.L().Error("connect DB failed, err:%v\n", zap.Error(err))
		return
	}
	Db.SetMaxOpenConns(viper.GetInt("mysql.max_open_connection"))
	Db.SetMaxIdleConns(viper.GetInt("mysql.max_idle_connection"))
	return
}

func Close(){
	_ = Db.Close()
}

