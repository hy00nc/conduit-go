package database

import (
	"github.com/hy00nc/conduit-go/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

// Open database
func InitDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	DB = db
	return DB
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
