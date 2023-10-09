package app

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
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

func writeResponse(w http.ResponseWriter, data map[string]interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func RegisterArticles(router *mux.Router) {
	router.HandleFunc("", GetArticles).Methods("GET")
	router.HandleFunc("/{slug}", GetArticle).Methods("GET")
	router.HandleFunc("/{slug}/comments", GetComments).Methods("GET")
}

func RegisterTags(router *mux.Router) {
	router.HandleFunc("", GetTags).Methods("GET")
}

func RegisterUsers(router *mux.Router) {
	// without authentication
	router.HandleFunc("", CreateUser).Methods("POST")      //, "OPTIONS")
	router.HandleFunc("/login", LoginUser).Methods("POST") //, "OPTIONS")
}

func RegisterProfiles(router *mux.Router) {
	router.HandleFunc("/{username}", GetProfile).Methods("GET")
}

func MakeWebHandler() http.Handler {
	// Create new router
	router := mux.NewRouter().PathPrefix("/api").Subrouter()

	RegisterArticles(router.PathPrefix("/articles").Subrouter())
	RegisterTags(router.PathPrefix("/tags").Subrouter())
	RegisterUsers(router.PathPrefix("/users").Subrouter())
	RegisterProfiles(router.PathPrefix("/profiles").Subrouter())

	// TODO: Add Swagger?
	// router.PathPrefix("/swagger").Hander(httpSwagger.WrapHandler)

	// Add middleware
	router.Use(loggingMiddleware)
	router.Use(ignoreOptionsMiddleware)

	return router
}
