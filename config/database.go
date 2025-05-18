package config

import (
	"gorm.io/gorm"
)

var DB *gorm.DB

func PostgresSQL() *gorm.DB {
	// dsn := "host=localhost user=gorm password=gorm dbname=gorm port=9920 sslmode=disable TimeZone=Asia/Shanghai"
	// db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	// db, err := gorm.Open(postgres.New(postgres.Config{
	// 	DSN:                  "user=gorm password=gorm dbname=gorm port=9920 sslmode=disable TimeZone=Asia/Shanghai",
	// 	PreferSimpleProtocol: true, // disables implicit prepared statement usage
	// }), &gorm.Config{})
	return &gorm.DB{}
}
