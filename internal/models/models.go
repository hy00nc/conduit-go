package models

import (
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Article struct {
	gorm.Model
	Slug        string
	Title       string
	Description string
	Body        string
	Author      Profile
	AuthorID    uint
	Tags        []Tag `gorm:"many2many:article_tags;"`
}

type Profile struct {
	gorm.Model
	Name  string `gorm:"unique"`
	Bio   string
	Image string
}

type User struct {
	gorm.Model
	Email     string `gorm:"unique"`
	Profile   Profile
	ProfileID uint
	Hash      string
}

func (u *User) CheckPassword(password string) error {
	bytePassword := []byte(password)
	byteHash := []byte(u.Hash)
	return bcrypt.CompareHashAndPassword(byteHash, bytePassword)
}

type Comment struct {
	gorm.Model
	Body      string
	Article   Article
	ArticleID uint
	Author    Profile
	AuthorID  uint
}

type Tag struct {
	gorm.Model
	Name string `gorm:"unique"`
}

type Follow struct {
	gorm.Model
	User        Profile
	UserID      uint
	Following   Profile
	FollowingID uint
}

type Favorite struct {
	gorm.Model
	Article       Article
	ArticleID     uint
	FavoritedBy   Profile
	FavoritedByID uint
}
