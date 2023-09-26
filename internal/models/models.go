package models

import (
	"gorm.io/gorm"
)

// FIXME: Temporary
type Article struct {
	gorm.Model
	Slug        string
	Title       string
	Description string
	Body        string
	Author      User
	AuthorID    int
	Tags        []Tag
}

// Use hook https://gorm.io/docs/hooks.html#content-inner
// FIXME: Does not work?
// func (a *Article) AfterCreate(tx *gorm.DB) (err error) {
// 	return tx.Model(a).Preload("Author").Error
// }

// FIXME: Temporary
type User struct {
	gorm.Model
	Name  string
	Bio   string
	Image string
}

type Tag struct {
	gorm.Model
	Name string
	ArticleID int
}
