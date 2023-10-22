package utils

import (
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const signingKey = "somethingVeryStrong"

func GetToken(id uint) (string, error) {
	token := jwt.NewWithClaims(
		jwt.GetSigningMethod("HS256"),
		jwt.MapClaims{
			"id":  id,
			"exp": time.Now().Add(time.Hour * 24).Unix(),
		},
	)
	signed, err := token.SignedString([]byte(signingKey))
	if err != nil {
		log.Println("Error while signing JWT:", err.Error())
		return "", err
	}
	return signed, err
}

func CheckToken(token string) (jwt.Claims, error) {
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(signingKey), nil
	})
	if err != nil {
		return nil, err
	}
	return parsedToken.Claims, err
}

type contextKey string

const ContextKeyUserData = contextKey("userData")

func CreateInvalidResponse(key string) map[string]interface{} {
	return map[string]interface{}{key: "is invalid"}
}
