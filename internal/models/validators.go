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
		Email string `json:"email" validate:"required"`
		Password string `json:"password" validate:"required"`
	} `json:"user"`
}