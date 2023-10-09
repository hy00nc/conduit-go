package utils

import (
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const secretKey = "somethingVeryStrong"

func GetToken(id uint) string {
	token := jwt.NewWithClaims(
		jwt.GetSigningMethod("HS256"),
		jwt.MapClaims{
			"id":  id,
			"exp": time.Now().Add(time.Hour * 24).Unix(),
		},
	)
	signed, err := token.SignedString([]byte(secretKey))
	if err != nil {
		log.Println("Error while signing JWT:", err.Error())
		return ""
	}
	return signed
}

const ContextKeyIsAuthenticated = "isAuthenticated"

func CreateInvalidResponse(key string) map[string]interface{} {
	return map[string]interface{}{key: "is invalid"}
}
