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
	db.AutoMigrate(&models.User{})
	db.AutoMigrate(&models.Tag{})
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

// FIXME: Remove
func CreateTestArticles(db *gorm.DB) {
	articles := []models.Article{
		{
			Slug:        "aaa",
			Title:       "test",
			Description: "test description",
			Body:        "test body",
			Author:      models.User{Name: "haeyoon", Bio: "This is haeyoon", Image: "https://static.productionready.io/images/smiley-cyrus.jpg"},
			Tags:        []models.Tag{{Name: "tag1"}, {Name: "tag2"}},
		},
		{
			Slug:        "eee",
			Title:       "another",
			Description: "another description",
			Body:        "another body",
			Author:      models.User{Name: "yoona", Bio: "This is yoona", Image: "https://static.productionready.io/images/smiley-cyrus.jpg"},
			Tags:        []models.Tag{{Name: "tag2"}, {Name: "tag3"}},
		},
	}

	db.Create(&articles)
	db.Save(&articles)
}
