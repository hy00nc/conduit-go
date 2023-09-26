package app

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/hy00nc/conduit-go/internal/database"
	"github.com/hy00nc/conduit-go/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Do stuff here
		log.Println(r.RequestURI)
		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
}

func writeResponse(w http.ResponseWriter, data map[string]interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}

func RetrieveArticles(tag, author, limit, offset, favorited string) ([]models.Article, int64, error) {
	db := database.GetDB()
	var articles []models.Article
	var count int64

	offset_int, err := strconv.Atoi(offset)
	if err != nil {
		offset_int = 0 // set default value
	}

	limit_int, err := strconv.Atoi(limit)
	if err != nil {
		limit_int = 20 // set default default value
	}
	// TODO: Add filtering for tag, author, favorited...
	// Get articles
	db.Model(&articles).Count(&count)
	db.Offset(offset_int).Limit(limit_int).Find(&articles)

	// TODO: Replace with hooks?
	err = db.Model(&articles).Preload(clause.Associations).Find(&articles).Error
	return articles, count, err
}

func RetrieveTags() ([]models.Tag, error) {
	db := database.GetDB()
	var tags []models.Tag

	err := db.Model(&tags).Find(&tags).Error
	return tags, err
}

func MakeWebHandler(db *gorm.DB) http.Handler {
	// Create new router
	router := mux.NewRouter()

	router.HandleFunc("/api/articles", func(w http.ResponseWriter, r *http.Request) {
		// Retrieve optional params
		tag := r.URL.Query().Get("tag")
		author := r.URL.Query().Get("author")
		favorited := r.URL.Query().Get("favorited")
		limit := r.URL.Query().Get("limit")
		offset := r.URL.Query().Get("offset")

		articles, count, err := RetrieveArticles(tag, author, limit, offset, favorited)
		if err != nil {
			log.Println(err)
			writeResponse(w, map[string]interface{}{"error": "Invalid param"}, http.StatusNotFound)
			return
		}
		serializer := models.ArticlesSerializer{articles}
		writeResponse(w, map[string]interface{}{"articles": serializer.Response(db), "articlesCount": count}, http.StatusOK)
	}).Methods("GET")

	router.HandleFunc("/api/tags", func(w http.ResponseWriter, r *http.Request) {
		// Return list of tags
		tags, err := RetrieveTags()
		if err != nil {
			log.Println(err)
			writeResponse(w, map[string]interface{}{"error": "Invalid param"}, http.StatusNotFound)
			return
		}
		serializer := models.TagsSerializer{tags}
		writeResponse(w, map[string]interface{}{"tags": serializer.Response()}, http.StatusOK)
	}).Methods("GET")

	// TODO: Add Swagger?
	// router.PathPrefix("/swagger").Hander(httpSwagger.WrapHandler)

	// Add middleware
	router.Use(loggingMiddleware)

	return router
}
