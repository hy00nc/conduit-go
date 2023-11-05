package database

import (
	"os"

	"github.com/hy00nc/conduit-go/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Open database
func InitDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("app.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	DB = db
	return DB
}

func InitTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		panic("failed to connect database")
	}
	DB = db
	return DB
}

func RemoveDB(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		panic("Error occurred while closing the DB.")
	}
	sqlDB.Close()
	err = os.Remove("test.db")
	return err
}

func MigrateDB(db *gorm.DB) {
	// Migrate the schema
	db.AutoMigrate(&models.Article{})
	db.AutoMigrate(&models.Profile{})
	db.AutoMigrate(&models.User{})
	db.AutoMigrate(&models.Tag{})
	db.AutoMigrate(&models.Comment{})
	db.AutoMigrate(&models.Follow{})
	db.AutoMigrate(&models.Favorite{})
}

func GetDB() *gorm.DB {
	return DB
}

func CloseDB(db *gorm.DB) {
	sqlDB, err := db.DB()
	if err != nil {
		panic("Error occurred while closing the DB.")
	}
	sqlDB.Close()
}
