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
	AuthorID    int
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
	ProfileID int
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
	ArticleID int
	Author    Profile
	AuthorID  int
}

type Tag struct {
	gorm.Model
	Name string `gorm:"unique"`
	Slug string
}
