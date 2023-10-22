package models

import (
	"net/http"

	"github.com/hy00nc/conduit-go/internal/utils"
	"gorm.io/gorm"
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

type UserRequest struct {
	User struct {
		Email    string `json:"email"`
		Bio      string `json:"bio"`
		Image    string `json:"image"`
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"user"`
}

type ArticleRequest struct {
	Article struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Body        string `json:"body"`
	} `json:"article"`
}

func (s *ArticleSerializer) Response(db *gorm.DB, r *http.Request) ArticleResponse {
	var userProfile Profile
	db.Where("id = ?", s.AuthorID).First(&userProfile)
	authorSerializer := ProfileSerializer{userProfile}
	favorited := false

	userData := r.Context().Value(utils.ContextKeyUserData)
	if userData != nil {
		userData := userData.(User)
		var favorite Favorite
		db.Where(&Favorite{ArticleID: s.Article.ID, FavoritedByID: userData.ProfileID}).Find(&favorite)
		favorited = favorite.ID != 0
	}

	var favorites []Favorite
	db.Find(&favorites, "article_id = ?", s.Article.ID)
	favoritesCount := len(favorites)

	response := ArticleResponse{
		Slug:           s.Slug,
		Title:          s.Title,
		Description:    s.Description,
		Body:           s.Body,
		CreatedAt:      s.CreatedAt.UTC().Format("2006-01-02T15:04:05.999Z"),
		UpdatedAt:      s.UpdatedAt.UTC().Format("2006-01-02T15:04:05.999Z"),
		Author:         authorSerializer.Response(db, r),
		Favorite:       favorited,
		FavoritesCount: uint(favoritesCount),
	}
	tagList := s.Article.Tags
	tagLen := len(tagList)
	tags := make([]string, tagLen)
	if tagList != nil {
		for i := 0; i < len(tagList); i++ {
			tags[i] = tagList[i].Name
		}
	}
	response.Tags = tags
	return response
}

func (s *ArticlesSerializer) Response(db *gorm.DB, r *http.Request) []ArticleResponse {
	response := []ArticleResponse{}
	for _, article := range s.Articles {
		serializer := ArticleSerializer{article}
		response = append(response, serializer.Response(db, r))
	}
	return response
}

func (s *ProfileSerializer) Response(db *gorm.DB, r *http.Request) ProfileResponse {
	userData := r.Context().Value(utils.ContextKeyUserData)
	following := false
	if userData != nil {
		userData := userData.(User)
		var follow Follow
		db.Where(&Follow{UserID: userData.ProfileID, FollowingID: s.Profile.ID}).Find(&follow)
		following = (follow.ID != 0)
	}
	userProfile := ProfileResponse{
		Username:  s.Name,
		Bio:       s.Bio,
		Image:     s.Image,
		Following: following,
	}
	return userProfile
}

func (s *UserSerializer) Response() UserResponse {
	token, _ := utils.GetToken(s.ID)
	userResp := UserResponse{
		Email:    s.Email,
		Token:    token,
		Username: s.Profile.Name,
		Bio:      s.Profile.Bio,
		Image:    s.Profile.Image,
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

func (s *CommentSerializer) Response(db *gorm.DB, r *http.Request) CommentResponse {
	var userProfile Profile
	db.Where("id = ?", s.AuthorID).First(&userProfile)
	authorSerializer := ProfileSerializer{userProfile}
	response := CommentResponse{
		ID:        s.ID,
		CreatedAt: s.CreatedAt.UTC().Format("2006-01-02T15:04:05.999Z"),
		UpdatedAt: s.UpdatedAt.UTC().Format("2006-01-02T15:04:05.999Z"),
		Body:      s.Body,
		Author:    authorSerializer.Response(db, r),
	}
	return response
}

func (s *CommentsSerializer) Response(db *gorm.DB, r *http.Request) []CommentResponse {
	response := []CommentResponse{}
	for _, comment := range s.Comments {
		serializer := CommentSerializer{comment}
		response = append(response, serializer.Response(db, r))
	}
	return response
}
