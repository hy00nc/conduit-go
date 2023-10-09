package app

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/hy00nc/conduit-go/internal/database"
	"github.com/hy00nc/conduit-go/internal/models"
	"github.com/hy00nc/conduit-go/internal/utils"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm/clause"
)

func RetrieveArticle(slug string) (models.Article, error) {
	db := database.GetDB()
	var article models.Article

	err := db.Model(&article).Preload(clause.Associations).Find(&article, "slug = ?", slug).Error
	return article, err
}

func RetrieveArticles(tag_param, author_param, limit_param, offset_param, favorited_param string) ([]models.Article, int64, error) {
	db := database.GetDB()
	var articles []models.Article
	var count int64

	offset_int, err := strconv.Atoi(offset_param)
	if err != nil {
		offset_int = 0 // set default value
	}

	limit_int, err := strconv.Atoi(limit_param)
	if err != nil {
		limit_int = 20 // set default default value
	}
	// TODO: Add filtering author and favorited (after adding authentication)
	if tag_param != "" {
		var tag models.Tag
		db.First(&tag, "name = ?", tag_param)
		result := map[string]interface{}{}
		db.Table("article_tags").Take(&result, "tag_id = ?", tag.ID)
		err = db.Model(&articles).Offset(offset_int).Limit(limit_int).Preload(clause.Associations).Find(&articles, "id = ?", result["article_id"]).Count(&count).Error
		return articles, count, err
	} else {
		db.Model(&articles).Count(&count)
		db.Offset(offset_int).Limit(limit_int).Find(&articles)
		// TODO: Replace preload with hooks?
		err = db.Model(&articles).Preload(clause.Associations).Find(&articles).Error
		return articles, count, err
	}
}

func RetrieveComments(slug string) ([]models.Comment, error) {
	db := database.GetDB()
	var comments []models.Comment
	var article models.Article

	err := db.Model(&article).Preload(clause.Associations).First(&article, "slug = ?", slug).Error
	if err != nil {
		return nil, err
	}
	err = db.Model(&comments).Preload(clause.Associations).Find(&comments, "article_id = ?", article.ID).Error
	return comments, err
}

func RetrieveTags() ([]models.Tag, error) {
	db := database.GetDB()
	var tags []models.Tag

	err := db.Model(&tags).Find(&tags).Error
	return tags, err
}

func RetrieveProfile(username string) (models.Profile, error) {
	db := database.GetDB()
	var profile models.Profile

	err := db.Model(&profile).First(&profile, "name = ?", username).Error
	return profile, err
}

func GetArticles(w http.ResponseWriter, r *http.Request) {
	// Retrieve optional params
	tag := r.URL.Query().Get("tag")
	author := r.URL.Query().Get("author")
	favorited := r.URL.Query().Get("favorited")
	limit := r.URL.Query().Get("limit")
	offset := r.URL.Query().Get("offset")

	articles, count, err := RetrieveArticles(tag, author, limit, offset, favorited)
	if err != nil {
		log.Println(err)
		writeResponse(w, map[string]interface{}{"errors": utils.CreateInvalidResponse("Parameter")}, http.StatusBadRequest)
		return
	}
	serializer := models.ArticlesSerializer{articles}
	writeResponse(w, map[string]interface{}{"articles": serializer.Response(database.GetDB()), "articlesCount": count}, http.StatusOK)
}

func GetArticle(w http.ResponseWriter, r *http.Request) {
	data := r.Header.Values("Authorization")
	log.Println("Data:", data)
	slug := mux.Vars(r)["slug"]
	article, err := RetrieveArticle(slug)
	if err != nil {
		log.Println(err)
		writeResponse(w, map[string]interface{}{"errors": utils.CreateInvalidResponse("Parameter")}, http.StatusBadRequest)
		return
	}
	serializer := models.ArticleSerializer{article}
	writeResponse(w, map[string]interface{}{"article": serializer.Response(database.GetDB())}, http.StatusOK)
}

func GetFeed(w http.ResponseWriter, r *http.Request) {
	data := r.Header.Values("Authorization")
	log.Println("Data:", data)
}

func GetComments(w http.ResponseWriter, r *http.Request) {
	slug := mux.Vars(r)["slug"]
	comments, err := RetrieveComments(slug)
	if err != nil {
		log.Println(err)
		writeResponse(w, map[string]interface{}{"errors": utils.CreateInvalidResponse("Parameter")}, http.StatusBadRequest)
		return
	}
	serializer := models.CommentsSerializer{comments}
	writeResponse(w, map[string]interface{}{"comments": serializer.Response(database.GetDB())}, http.StatusOK)
}

func GetTags(w http.ResponseWriter, r *http.Request) {
	// Return list of tags
	tags, err := RetrieveTags()
	if err != nil {
		log.Println(err)
		writeResponse(w, map[string]interface{}{"errors": utils.CreateInvalidResponse("Parameter")}, http.StatusBadRequest)
		return
	}
	serializer := models.TagsSerializer{tags}
	writeResponse(w, map[string]interface{}{"tags": serializer.Response()}, http.StatusOK)
}

func GetProfile(w http.ResponseWriter, r *http.Request) {
	username := mux.Vars(r)["username"]

	profile, err := RetrieveProfile(username)
	if err != nil {
		log.Println(err)
		writeResponse(w, map[string]interface{}{"errors": utils.CreateInvalidResponse("Parameter")}, http.StatusBadRequest)
	}
	serializer := models.ProfileSerializer{profile}
	writeResponse(w, map[string]interface{}{"profile": serializer.Response(database.GetDB())}, http.StatusOK)
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	// get user data from request
	var registerValidator models.RegisterValidator
	err := json.NewDecoder(r.Body).Decode(&registerValidator)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		writeResponse(w, map[string]interface{}{"errors": utils.CreateInvalidResponse("Parameter")}, http.StatusBadRequest)
		return
	}
	validate := validator.New()
	err = validate.Struct(registerValidator)
	if err != nil {
		writeResponse(w, map[string]interface{}{"errors": utils.CreateInvalidResponse("Parameter")}, http.StatusBadRequest)
		return
	}
	// Create User to database
	hash, _ := bcrypt.GenerateFromPassword([]byte(registerValidator.User.Password), bcrypt.DefaultCost)
	user := models.User{
		Email: registerValidator.User.Email,
		Profile: models.Profile{
			Name:  registerValidator.User.Username,
			Bio:   "",
			Image: "https://static.productionready.io/images/smiley-cyrus.jpg", // default image
		},
		Hash: string(hash),
	}
	db := database.GetDB()
	err = db.Create(&user).Error
	if err != nil {
		log.Println(err.Error())
		writeResponse(w, map[string]interface{}{"errors": utils.CreateInvalidResponse("Parameter")}, http.StatusNotFound)
		return
	}
	err = db.Save(&user).Error
	if err != nil {
		log.Println(err.Error())
		writeResponse(w, map[string]interface{}{"errors": utils.CreateInvalidResponse("Parameter")}, http.StatusNotFound)
		return
	}

	serializer := models.UserSerializer{user}
	writeResponse(w, map[string]interface{}{"user": serializer.Response(db)}, http.StatusCreated)

	// Update Context?
	// ctx := context.WithValue(r.Context(), utils.ContextKeyIsAuthenticated, true)
	// r = r.WithContext(ctx)
}

func LoginUser(w http.ResponseWriter, r *http.Request) {
	// get user data from request
	var loginValidator models.LoginValidator
	err := json.NewDecoder(r.Body).Decode(&loginValidator)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	validate := validator.New()
	err = validate.Struct(loginValidator)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Match with user from DB
	db := database.GetDB()
	var user models.User
	err = db.Model(&user).Preload(clause.Associations).Where("email = ?", loginValidator.User.Email).First(&user).Error
	if err != nil {
		log.Println(err.Error())
		writeResponse(w, map[string]interface{}{"errors": utils.CreateInvalidResponse("email or password")}, http.StatusForbidden)
		return
	}

	if err = user.CheckPassword(loginValidator.User.Password); err != nil {
		log.Println(err.Error())
		writeResponse(w, map[string]interface{}{"errors": utils.CreateInvalidResponse("email or password")}, http.StatusForbidden)
		return
	}

	// Update Context?
	// ctx := context.WithValue(r.Context(), utils.ContextKeyIsAuthenticated, true)
	// r = r.WithContext(ctx)

	serializer := models.UserSerializer{user}
	writeResponse(w, map[string]interface{}{"user": serializer.Response(db)}, http.StatusOK)
}

// func GetCurrentUser(w http.ResponseWriter, r *http.Request) {
// 	isAuthenticated, ok := r.Context().Value(utils.ContextKeyIsAuthenticated).(bool)
// 	if !ok {
// 		writeResponse(w, map[string]interface{}{"errors": utils.CreateInvalidResponse("Authentication")}, http.StatusForbidden)
// 		return
// 	}
// 	if !isAuthenticated {
// 		writeResponse(w, map[string]interface{}{"errors": utils.CreateInvalidResponse("Authentication")}, http.StatusForbidden)
// 		return
// 	}
// 	log.Println("Authentication okay")
// }
