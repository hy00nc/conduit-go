package models

type LoginValidator struct {
	User struct {
		Email    string `json:"email" validate:"required"`
		Password string `json:"password" validate:"required"`
	} `json:"user"`
}

type RegisterValidator struct {
	User struct {
		Username string `json:"username" validate:"required"`
		Email    string `json:"email" validate:"required"`
		Password string `json:"password" validate:"required"`
	} `json:"user"`
}

type ArticleValidator struct {
	Article struct {
		Title       string   `json:"title" validate:"required"`
		Description string   `json:"description" validate:"required"`
		Body        string   `json:"body" validate:"required"`
		TagList     []string `json:"tagList"`
	} `json:"article"`
}

type CommentValidator struct {
	Comment struct {
		Body string `json:"body" validate:"required"`
	} `json:"comment"`
}
