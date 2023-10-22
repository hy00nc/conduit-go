package app

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gosimple/slug"
	"github.com/hy00nc/conduit-go/internal/database"
	"github.com/hy00nc/conduit-go/internal/models"
	"github.com/hy00nc/conduit-go/internal/utils"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm/clause"
)

func RetrieveArticle(slug string) (models.Article, error) {
	db := database.GetDB()
	var article models.Article

	err := db.Model(&article).Preload(clause.Associations).Preload("Tags").Find(&article, "slug = ?", slug).Error
	return article, err
}

func RetrieveArticles(tag_param, author_param, limit_param, offset_param, favorited_param string) ([]models.Article, int64, error) {
	db := database.GetDB()
	var articles []models.Article
	var count int64
	var err error

	offset_int, err := strconv.Atoi(offset_param)
	if err != nil {
		offset_int = 0 // set default value
	}

	limit_int, err := strconv.Atoi(limit_param)
	if err != nil {
		limit_int = 20 // set default default value
	}
	if tag_param != "" {
		var tag models.Tag
		db.First(&tag, "name = ?", tag_param)
		var arr []string
		db.Table("article_tags").Select("article_id").Where("tag_id = ?", tag.ID).Find(&arr)
		err = db.Model(&articles).Order("created_at desc").Offset(offset_int).Limit(limit_int).Preload(clause.Associations).Find(&articles, "id IN ?", arr).Count(&count).Error
	} else if author_param != "" {
		var profile models.Profile
		db.First(&profile, "name = ?", author_param)
		err = db.Model(&articles).Order("created_at desc").Offset(offset_int).Limit(limit_int).Preload(clause.Associations).Find(&articles, "author_id = ?", profile.ID).Count(&count).Error
	} else if favorited_param != "" {
		var profile models.Profile
		db.First(&profile, "name = ?", favorited_param)
		var favoriteArticleIds []uint
		db.Table("favorites").Select("article_id").Where("favorited_by_id = ?", profile.ID).Where("deleted_at IS NULL").Find(&favoriteArticleIds)
		err = db.Model(&articles).Where("id IN ?", favoriteArticleIds).Order("created_at desc").Offset(offset_int).Limit(limit_int).Preload(clause.Associations).Find(&articles).Count(&count).Error
	} else {
		err = db.Model(&articles).Order("created_at desc").Offset(offset_int).Limit(limit_int).Preload(clause.Associations).Find(&articles).Count(&count).Error
	}
	return articles, count, err
}

func RetrieveArticlesFeed(userData models.User, limitParam, offsetParam string) ([]models.Article, int64, error) {
	db := database.GetDB()
	var articles []models.Article
	var count int64

	offset_int, err := strconv.Atoi(offsetParam)
	if err != nil {
		offset_int = 0 // set default value
	}

	limit_int, err := strconv.Atoi(limitParam)
	if err != nil {
		limit_int = 20 // set default default value
	}

	var followingIds []uint
	db.Table("follows").Select("following_id").Where("user_id = ?", userData.ProfileID).Where("deleted_at IS NULL").Find(&followingIds)
	err = db.Where("author_id IN ?", followingIds).Order("created_at desc").Offset(offset_int).Limit(limit_int).Preload(clause.Associations).Find(&articles).Count(&count).Error
	return articles, count, err
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
	writeResponse(w, map[string]interface{}{"articles": serializer.Response(database.GetDB(), r), "articlesCount": count}, http.StatusOK)
}

func GetArticle(w http.ResponseWriter, r *http.Request) {
	slug := mux.Vars(r)["slug"]
	article, err := RetrieveArticle(slug)
	if err != nil {
		log.Println(err)
		writeResponse(w, map[string]interface{}{"errors": utils.CreateInvalidResponse("Parameter")}, http.StatusBadRequest)
		return
	}
	serializer := models.ArticleSerializer{article}
	writeResponse(w, map[string]interface{}{"article": serializer.Response(database.GetDB(), r)}, http.StatusOK)
}

func GetFeed(w http.ResponseWriter, r *http.Request) {
	// Retrieve optional params
	limit := r.URL.Query().Get("limit")
	offset := r.URL.Query().Get("offset")

	userData := r.Context().Value(utils.ContextKeyUserData).(models.User)
	articles, count, err := RetrieveArticlesFeed(userData, limit, offset)
	if err != nil {
		log.Println(err)
		writeResponse(w, map[string]interface{}{"errors": utils.CreateInvalidResponse("Parameter")}, http.StatusBadRequest)
		return
	}
	serializer := models.ArticlesSerializer{articles}
	writeResponse(w, map[string]interface{}{"articles": serializer.Response(database.GetDB(), r), "articlesCount": count}, http.StatusOK)
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
	writeResponse(w, map[string]interface{}{"comments": serializer.Response(database.GetDB(), r)}, http.StatusOK)
}

func AddComments(w http.ResponseWriter, r *http.Request) {
	userData := r.Context().Value(utils.ContextKeyUserData).(models.User)

	// get comment data from request
	var commentValidator models.CommentValidator
	if err := json.NewDecoder(r.Body).Decode(&commentValidator); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		writeResponse(w, map[string]interface{}{"errors": utils.CreateInvalidResponse("Parameter")}, http.StatusBadRequest)
		return
	}
	validate := validator.New()
	if err := validate.Struct(commentValidator); err != nil {
		writeResponse(w, map[string]interface{}{"errors": utils.CreateInvalidResponse("Parameter")}, http.StatusBadRequest)
		return
	}
	// Create comment in database
	slugParam := mux.Vars(r)["slug"]
	db := database.GetDB()
	var article models.Article
	db.Model(&article).Find(&article, "slug = ?", slugParam)
	comment := models.Comment{
		Body:      commentValidator.Comment.Body,
		ArticleID: article.ID,
		AuthorID:  userData.ID,
	}
	db.Create(&comment)
	db.Save(&comment)
	serializer := models.CommentSerializer{comment}
	writeResponse(w, map[string]interface{}{"comment": serializer.Response(db, r)}, http.StatusCreated)
}

func DeleteComment(w http.ResponseWriter, r *http.Request) {
	idParam := mux.Vars(r)["id"]
	db := database.GetDB()
	var comment models.Comment
	db.Model(&comment).Find(&comment, "id = ?", idParam)
	db.Delete(&comment)
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
	writeResponse(w, map[string]interface{}{"profile": serializer.Response(database.GetDB(), r)}, http.StatusOK)
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	// get user data from request
	var registerValidator models.RegisterValidator
	if err := json.NewDecoder(r.Body).Decode(&registerValidator); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		writeResponse(w, map[string]interface{}{"errors": utils.CreateInvalidResponse("Parameter")}, http.StatusBadRequest)
		return
	}
	validate := validator.New()
	if err := validate.Struct(registerValidator); err != nil {
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
	if err := db.Create(&user).Error; err != nil {
		log.Println(err.Error())
		writeResponse(w, map[string]interface{}{"errors": utils.CreateInvalidResponse("Parameter")}, http.StatusNotFound)
		return
	}
	if err := db.Save(&user).Error; err != nil {
		log.Println(err.Error())
		writeResponse(w, map[string]interface{}{"errors": utils.CreateInvalidResponse("Parameter")}, http.StatusNotFound)
		return
	}

	serializer := models.UserSerializer{user}
	writeResponse(w, map[string]interface{}{"user": serializer.Response()}, http.StatusCreated)
}

func LoginUser(w http.ResponseWriter, r *http.Request) {
	// get user data from request
	var loginValidator models.LoginValidator
	if err := json.NewDecoder(r.Body).Decode(&loginValidator); err != nil {
		log.Println(err.Error())
		writeResponse(w, map[string]interface{}{"errors": utils.CreateInvalidResponse("Data")}, http.StatusBadRequest)
		return
	}
	validate := validator.New()
	if err := validate.Struct(loginValidator); err != nil {
		log.Println(err.Error())
		writeResponse(w, map[string]interface{}{"errors": utils.CreateInvalidResponse("Data")}, http.StatusBadRequest)
		return
	}
	// Match user from DB
	db := database.GetDB()
	var user models.User
	db.Model(&user).Preload(clause.Associations).Where("email = ?", loginValidator.User.Email).First(&user)

	if err := user.CheckPassword(loginValidator.User.Password); err != nil || user.ID == 0 {
		log.Println(err.Error())
		writeResponse(w, map[string]interface{}{"errors": utils.CreateInvalidResponse("email or password")}, http.StatusForbidden)
		return
	}

	serializer := models.UserSerializer{user}
	writeResponse(w, map[string]interface{}{"user": serializer.Response()}, http.StatusOK)
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	userData := r.Context().Value(utils.ContextKeyUserData).(models.User)
	serializer := models.UserSerializer{userData}
	writeResponse(w, map[string]interface{}{"user": serializer.Response()}, http.StatusOK)
}

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	userData := r.Context().Value(utils.ContextKeyUserData).(models.User)
	// get user data from request
	var userRequest models.UserRequest
	err := json.NewDecoder(r.Body).Decode(&userRequest)
	if err != nil {
		log.Println(err.Error())
		writeResponse(w, map[string]interface{}{"errors": utils.CreateInvalidResponse("Data")}, http.StatusBadRequest)
		return
	}
	// perform update on DB
	db := database.GetDB()

	// profile update
	if userRequest.User.Bio != "" || userRequest.User.Username != "" || userRequest.User.Image != "" {
		var profile models.Profile
		db.Model(&profile).Find(&profile, "id = ?", userData.ProfileID)
		db.Model(&profile).Updates(
			models.Profile{
				Bio:   userRequest.User.Bio,
				Name:  userRequest.User.Username,
				Image: userRequest.User.Image,
			},
		)
	}

	// user update
	if userRequest.User.Password != "" || userRequest.User.Email != "" {
		var hash string
		if userRequest.User.Password != "" {
			hashByte, _ := bcrypt.GenerateFromPassword([]byte(userRequest.User.Password), bcrypt.DefaultCost)
			hash = string(hashByte)
		}
		db.Model(&userData).Updates(
			models.User{
				Email: userRequest.User.Email,
				Hash:  hash,
			},
		)
	}

	// Retrieve updated
	db.Model(&userData).Preload(clause.Associations).Find(&userData, "id = ?", userData.ID)

	// Return response
	serializer := models.UserSerializer{userData}
	writeResponse(w, map[string]interface{}{"user": serializer.Response()}, http.StatusOK)
}

func CreateArticle(w http.ResponseWriter, r *http.Request) {
	// get article data from request
	userData := r.Context().Value(utils.ContextKeyUserData).(models.User)
	var articleValidator models.ArticleValidator
	if err := json.NewDecoder(r.Body).Decode(&articleValidator); err != nil {
		log.Println(err.Error())
		writeResponse(w, map[string]interface{}{"errors": utils.CreateInvalidResponse("Data")}, http.StatusBadRequest)
		return
	}
	validate := validator.New()
	if err := validate.Struct(articleValidator); err != nil {
		log.Println(err.Error())
		writeResponse(w, map[string]interface{}{"errors": utils.CreateInvalidResponse("Data")}, http.StatusBadRequest)
		return
	}
	// Create Article data
	db := database.GetDB()
	tagList := articleValidator.Article.TagList
	tagLen := len(tagList)
	tags := make([]models.Tag, tagLen)
	for i := 0; i < len(tagList); i++ {
		var tag models.Tag
		db.Where("name = ?", tagList[i]).Find(&tag)
		if tag.ID != 0 {
			tags[i] = tag
		} else {
			tags[i].Name = tagList[i]
		}
	}
	article := models.Article{
		Slug:        slug.Make(articleValidator.Article.Title + " " + uuid.NewString()),
		Title:       articleValidator.Article.Title,
		Description: articleValidator.Article.Description,
		Body:        articleValidator.Article.Body,
		AuthorID:    userData.ProfileID,
		Tags:        tags,
	}
	db.Create(&article)
	db.Save(&article)
	serializer := models.ArticleSerializer{article}
	writeResponse(w, map[string]interface{}{"article": serializer.Response(db, r)}, http.StatusOK)
}

func ArticleSlugEndpointAuthenticated(w http.ResponseWriter, r *http.Request) {
	// Get article by slug
	slugParam := mux.Vars(r)["slug"]
	db := database.GetDB()
	var article models.Article
	db.Model(&article).Find(&article, "slug = ?", slugParam)

	if r.Method == "DELETE" {
		db.Delete(&article)
	} else { // PUT
		// Get request data
		var articleRequest models.ArticleRequest
		err := json.NewDecoder(r.Body).Decode(&articleRequest)
		if err != nil {
			log.Println(err.Error())
			writeResponse(w, map[string]interface{}{"errors": utils.CreateInvalidResponse("Data")}, http.StatusBadRequest)
			return
		}
		db.Model(&article).Updates(
			models.Article{
				Title:       articleRequest.Article.Title,
				Description: articleRequest.Article.Description,
				Body:        articleRequest.Article.Body,
			},
		)
		if articleRequest.Article.Title != "" {
			db.Model(&article).Updates(
				models.Article{
					Slug: slug.Make(article.Title) + uuid.NewString(),
				},
			)
		}

		// Retrieve updated
		db.Model(&article).Preload(clause.Associations).Find(&article, "id = ?", article.ID)

		// Return response
		serializer := models.ArticleSerializer{article}
		writeResponse(w, map[string]interface{}{"article": serializer.Response(db, r)}, http.StatusOK)
	}
}

func FollowUserEndpoint(w http.ResponseWriter, r *http.Request) {
	currUser := r.Context().Value(utils.ContextKeyUserData).(models.User)
	// get use data to un/follow
	db := database.GetDB()
	targetUsername := mux.Vars(r)["username"]
	var targetUserProfile models.Profile
	db.Model(&targetUserProfile).Find(&targetUserProfile, "name = ?", targetUsername)

	var followData models.Follow
	db.FirstOrCreate(
		&followData,
		models.Follow{
			UserID:      currUser.ProfileID,
			FollowingID: targetUserProfile.ID,
		},
	)
	if r.Method == "DELETE" {
		db.Delete(&followData)
	}
	serializer := models.ProfileSerializer{targetUserProfile}
	writeResponse(w, map[string]interface{}{"profile": serializer.Response(db, r)}, http.StatusOK)
}

func FavoriteArticleEndpoint(w http.ResponseWriter, r *http.Request) {
	// get article data to unfavorite
	articleSlug2Favorite := mux.Vars(r)["slug"]
	db := database.GetDB()
	var targetArticle models.Article
	db.Model(&targetArticle).Preload(clause.Associations).Find(&targetArticle, "slug = ?", articleSlug2Favorite)

	var favoriteData models.Favorite
	userData := r.Context().Value(utils.ContextKeyUserData).(models.User)
	db.FirstOrCreate(
		&favoriteData,
		models.Favorite{
			ArticleID:     targetArticle.ID,
			FavoritedByID: userData.ProfileID,
		},
	)
	if r.Method == "DELETE" {
		db.Delete(&favoriteData)
	}
	serializer := models.ArticleSerializer{targetArticle}
	writeResponse(w, map[string]interface{}{"article": serializer.Response(db, r)}, http.StatusOK)
}
