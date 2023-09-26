package models

import (
	"gorm.io/gorm"
)

type ArticleSerializer struct {
	Article
}

type ArticlesSerializer struct {
	Articles []Article
}

type UserSerializer struct {
	User
}

type TagSerializer struct {
	Tag
}

type TagsSerializer struct {
	Tags []Tag
}

type ArticleResponse struct {
	Title          string       `json:"title"`
	Slug           string       `json:"slug"`
	Description    string       `json:"description"`
	Body           string       `json:"body"`
	CreatedAt      string       `json:"createdAt"`
	UpdatedAt      string       `json:"updatedAt"`
	Author         UserResponse `json:"author"`
	Tags           []string     `json:"tagList"`
	Favorite       bool         `json:"favorited"`
	FavoritesCount uint         `json:"favoritesCount"`
}

type UserResponse struct {
	Username  string `json:"username"`
	Bio       string `json:"bio"`
	Image     string `json:"image"`
	Following bool   `json:"following"`
}

func (s *ArticleSerializer) Response(db *gorm.DB) ArticleResponse {
	user := User{Name: s.Author.Name}
	db.Where("name = ?", s.Author.Name).First(&user)
	authorSerializer := UserSerializer{user}
	response := ArticleResponse{
		Slug:           s.Slug, // FIXME: Generate slug instead
		Title:          s.Title,
		Description:    s.Description,
		Body:           s.Body,
		CreatedAt:      s.CreatedAt.UTC().Format("2006-01-02T15:04:05.999Z"),
		UpdatedAt:      s.UpdatedAt.UTC().Format("2006-01-02T15:04:05.999Z"),
		Author:         authorSerializer.Response(db),
		Favorite:       false, // FIXME
		FavoritesCount: 0,     // FIXME
	}
	response.Tags = make([]string, 0)
	// TODO: Add tags
	return response
}

func (s *ArticlesSerializer) Response(db *gorm.DB) []ArticleResponse {
	response := []ArticleResponse{}
	for _, article := range s.Articles {
		serializer := ArticleSerializer{article}
		response = append(response, serializer.Response(db))
	}
	return response
}

func (s *UserSerializer) Response(db *gorm.DB) UserResponse {
	userInfo := UserResponse{
		Username:  s.Name,
		Bio:       s.Bio,
		Image:     s.Image,
		Following: false, // FIXME
	}
	return userInfo
}

func (s *TagSerializer) Response() string {
	return s.Name
}

func (s *TagsSerializer) Response() []string {
	response := []string{}
	for _, tag := range s.Tags {
		serializer := TagSerializer{tag}
		response = append(response, serializer.Response())
	}
	return response
}
