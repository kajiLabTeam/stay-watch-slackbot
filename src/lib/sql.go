package lib

import (
	"fmt"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func SQLConnect() (database *gorm.DB) {
	var db *gorm.DB
	var err error

	user := getEnv("MYSQL_USER", "")
	password := getEnv("MYSQL_PASSWORD", "")
	protocol := getEnv("MYSQL_PROTOCOL", "")
	dbname := getEnv("MYSQL_DBNAME", "")

	dsn := fmt.Sprintf("%s:%s@%s/%s?charset=utf8&parseTime=true&loc=UTC",
		user, password, protocol, dbname)
	dialector := mysql.Open(dsn)
	// log.Default().Println(dsn)

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
