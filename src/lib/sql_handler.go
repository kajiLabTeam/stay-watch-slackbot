package lib

import (
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/kajiLabTeam/stay-watch-slackbot/conf"
)

func SqlConnect() (database *gorm.DB) {
	var db *gorm.DB
	var err error
	c := conf.GetMysqlConfig()
	dsn := c.GetString("mysql.user") + ":" + c.GetString("mysql.password") + "@" + c.GetString("mysql.protocol") + "/" + c.GetString("mysql.dbname") + "?charset=utf8&parseTime=true&loc=Asia%2FTokyo"
	dialector := mysql.Open(dsn)

	if db, err = gorm.Open(dialector); err != nil {
		db = connect(dialector, 10)
	}
	fmt.Println("db connected!!")

	return db
}

func connect(dialector gorm.Dialector, count uint) *gorm.DB {
	var err error
	var db *gorm.DB
	if db, err = gorm.Open(dialector); err != nil {
		if count > 1 {
			time.Sleep(time.Second * 2)
			count--
			fmt.Printf("retry... count:%v\n", count)
			connect(dialector, count)
		}
		panic(err.Error())
	}
	return (db)
}
