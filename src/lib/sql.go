package lib

import (
	"fmt"
	"time"

	"github.com/kajiLabTeam/stay-watch-slackbot/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func SQLConnect() (database *gorm.DB) {
	var db *gorm.DB
	var err error

	dsn := fmt.Sprintf("%s:%s@%s/%s?charset=utf8&parseTime=true&loc=Asia%%2FTokyo",
		config.DB.User, config.DB.Password, config.DB.Protocol, config.DB.DBName)
	dialector := mysql.Open(dsn)
	// log.Default().Println(dsn)

	if db, err = gorm.Open(dialector); err != nil {
		db = connect(dialector, config.DB.RetryCount)
	}
	fmt.Println("db connected!!")

	return db
}

func connect(dialector gorm.Dialector, count int) *gorm.DB {
	var err error
	var db *gorm.DB
	if db, err = gorm.Open(dialector); err != nil {
		if count > 1 {
			time.Sleep(config.DB.RetryInterval)
			count--
			fmt.Printf("retry... count:%v\n", count)
			connect(dialector, count)
		}
		panic(err.Error())
	}
	return (db)
}
