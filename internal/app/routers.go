package app

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/hy00nc/conduit-go/internal/database"
	"github.com/hy00nc/conduit-go/internal/models"
	"github.com/hy00nc/conduit-go/internal/utils"
	"gorm.io/gorm/clause"
)

func loggingMiddleware(next http.Handler) http.Handler {
	return handlers.LoggingHandler(os.Stdout, next)
}

func ignoreOptionsMiddleware(next http.Handler) http.Handler {
	// Handling OPTIONS.... why is it not supported by IgnoreOptions in CORS? why should it return nil??
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			return
		}
		next.ServeHTTP(w, r)
	})
}

func matchAuthOptionalRoutes(url string) (matched bool, err error) {
	regexExp := [...]string{"/api/articles", "/api/profiles/([a-zA-z]+$)"}
	for i := 0; i < len(regexExp); i++ {
		if matched, err = regexp.MatchString(regexExp[i], url); matched {
			return
		}
	}
	return false, nil
}

func jwtMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		// empty token
		if len(tokenString) == 0 {
			// exclude optional authentication APIs
			if matched, _ := matchAuthOptionalRoutes(r.URL.Path); matched {
				next.ServeHTTP(w, r)
				return
			}
			writeResponse(w, map[string]interface{}{"errors": utils.CreateInvalidResponse("Authorization Header")}, http.StatusUnauthorized)
			return
		}
		tokenString = strings.Replace(tokenString, "Token ", "", 1)
		claims, err := utils.CheckToken(tokenString)
		if err != nil {
			writeResponse(w, map[string]interface{}{"errors": utils.CreateInvalidResponse("JWT Token")}, http.StatusUnauthorized)
			return
		}
		userId := claims.(jwt.MapClaims)["id"]

		var userData models.User
		db := database.GetDB()
		db.Model(&userData).Preload(clause.Associations).First(&userData, userId)

		if userData.ID == 0 {
			writeResponse(w, map[string]interface{}{"errors": utils.CreateInvalidResponse("User data")}, http.StatusUnauthorized)
			return
		}

		// Update context
		ctx := context.WithValue(r.Context(), utils.ContextKeyUserData, userData)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func writeResponse(w http.ResponseWriter, data map[string]interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func RegisterArticles(router *mux.Router) {
	router.HandleFunc("/{slug}", GetArticle).Methods("GET")
	router.HandleFunc("/{slug}/comments", GetComments).Methods("GET")
}

func RegisterArticlesAuthenticated(router *mux.Router) {
	// with authentication
	router.Use(jwtMiddleware)
	router.HandleFunc("", GetArticles).Methods("GET")
	router.HandleFunc("", CreateArticle).Methods("POST")
	router.HandleFunc("/feed", GetFeed).Methods("GET")
	router.HandleFunc("/{slug}", ArticleSlugEndpointAuthenticated).Methods("PUT", "DELETE")
	router.HandleFunc("/{slug}/comments", AddComments).Methods("POST")
	router.HandleFunc("/{slug}/comments/{id}", DeleteComment).Methods("DELETE")
	router.HandleFunc("/{slug}/favorite", FavoriteArticleEndpoint).Methods("POST", "DELETE")
}

func RegisterTags(router *mux.Router) {
	router.HandleFunc("", GetTags).Methods("GET")
}

func RegisterUsers(router *mux.Router) {
	// without authentication
	router.HandleFunc("", CreateUser).Methods("POST")
	router.HandleFunc("/login", LoginUser).Methods("POST")
}

func RegisterUser(router *mux.Router) {
	router.Use(jwtMiddleware)
	router.HandleFunc("", GetUser).Methods("GET")
	router.HandleFunc("", UpdateUser).Methods("PUT")
}

func RegisterProfiles(router *mux.Router) {
	router.Use(jwtMiddleware)
	router.HandleFunc("/{username}", GetProfile).Methods("GET")
	router.HandleFunc("/{username}/follow", FollowUserEndpoint).Methods("POST", "DELETE")
}

func MakeWebHandler(log bool) http.Handler {
	// Create new router
	router := mux.NewRouter().PathPrefix("/api").Subrouter()

	RegisterArticlesAuthenticated(router.PathPrefix("/articles").Subrouter())
	RegisterArticles(router.PathPrefix("/articles").Subrouter())
	RegisterTags(router.PathPrefix("/tags").Subrouter())
	RegisterUsers(router.PathPrefix("/users").Subrouter())
	RegisterUser(router.PathPrefix("/user").Subrouter())
	RegisterProfiles(router.PathPrefix("/profiles").Subrouter())

	// TODO: Add Swagger?
	// router.PathPrefix("/swagger").Hander(httpSwagger.WrapHandler)

	// Add middleware
	if log {
		router.Use(loggingMiddleware)
	}
	router.Use(ignoreOptionsMiddleware)

	return router
}
