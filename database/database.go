package database

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"time"
)

var db *gorm.DB

type Database struct {
	Host     string `json:"host"`
	Port     uint   `json:"port"`
	Username string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
}

func InitDatabase(dbInfo Database) error {
	var err error

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbInfo.Username,
		dbInfo.Password,
		dbInfo.Host,
		dbInfo.Port,
		dbInfo.Database,
	)

	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Discard,
	})
	if err != nil {
		return err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxIdleTime(time.Minute * 5)

	_ = db.AutoMigrate(&User{}, &Session{}, &Message{})

	return nil
}

func GetDatabase() *gorm.DB {
	return db
}
