package models

import (
	"github.com/hy00nc/conduit-go/internal/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ArticleSerializer struct {
	Article
}

type ArticlesSerializer struct {
	Articles []Article
}

type ProfileSerializer struct {
	Profile
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

type CommentSerializer struct {
	Comment
}

type CommentsSerializer struct {
	Comments []Comment
}

type ArticleResponse struct {
	Title          string          `json:"title"`
	Slug           string          `json:"slug"`
	Description    string          `json:"description"`
	Body           string          `json:"body"`
	CreatedAt      string          `json:"createdAt"`
	UpdatedAt      string          `json:"updatedAt"`
	Author         ProfileResponse `json:"author"`
	Tags           []string        `json:"tagList"`
	Favorite       bool            `json:"favorited"`
	FavoritesCount uint            `json:"favoritesCount"`
}

type ProfileResponse struct {
	Username  string `json:"username"`
	Bio       string `json:"bio"`
	Image     string `json:"image"`
	Following bool   `json:"following"`
}

type CommentResponse struct {
	ID        uint            `json:"id"`
	CreatedAt string          `json:"createdAt"`
	UpdatedAt string          `json:"updatedAt"`
	Body      string          `json:"body"`
	Author    ProfileResponse `json:"author"`
}

type UserResponse struct {
	Email    string `json:"email"`
	Token    string `json:"token"`
	Username string `json:"username"`
	Bio      string `json:"bio"`
	Image    string `json:"image"`
}

func (s *ArticleSerializer) Response(db *gorm.DB) ArticleResponse {
	var userProfile Profile
	db.Where("name = ?", s.Author.Name).First(&userProfile)
	authorSerializer := ProfileSerializer{userProfile}
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

func (s *ProfileSerializer) Response(db *gorm.DB) ProfileResponse {
	userProfile := ProfileResponse{
		Username:  s.Name,
		Bio:       s.Bio,
		Image:     s.Image,
		Following: false, // FIXME
	}
	return userProfile
}

func (s *UserSerializer) Response(db *gorm.DB) UserResponse {
	var userData User
	db.Model(&userData).Preload(clause.Associations).Where("id = ?", s.ID).First(&userData)
	userResp := UserResponse{
		Email:    userData.Email,
		Token:    utils.GetToken(s.ID),
		Username: userData.Profile.Name,
		Bio:      userData.Profile.Bio,
		Image:    userData.Profile.Image,
	}

	return userResp
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

func (s *CommentSerializer) Response(db *gorm.DB) CommentResponse {
	var userProfile Profile
	db.Where("id = ?", s.AuthorID).First(&userProfile)
	authorSerializer := ProfileSerializer{userProfile}
	response := CommentResponse{
		ID:        s.ID,
		CreatedAt: s.CreatedAt.UTC().Format("2006-01-02T15:04:05.999Z"),
		UpdatedAt: s.UpdatedAt.UTC().Format("2006-01-02T15:04:05.999Z"),
		Body:      s.Body,
		Author:    authorSerializer.Response(db),
	}
	return response
}

func (s *CommentsSerializer) Response(db *gorm.DB) []CommentResponse {
	response := []CommentResponse{}
	for _, comment := range s.Comments {
		serializer := CommentSerializer{comment}
		response = append(response, serializer.Response(db))
	}
	return response
}
