package app

import (
	"log"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/hy00nc/conduit-go/internal/database"
)

func RunServer() {
	// Get db
	db := database.InitDB()
	database.MigrateDB(db)
	defer database.CloseDB(db)

	// Headers
	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
	originsOk := handlers.AllowedOrigins([]string{"http://localhost:4100", "http://0.0.0.0:4100"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS", "DELETE"})
	allowCredentials := handlers.AllowCredentials()
	// ignoreOptions := handlers.IgnoreOptions()

	// Start serving
	log.Fatal(http.ListenAndServe(":8000", handlers.CORS(originsOk, headersOk, methodsOk, allowCredentials)(MakeWebHandler(true))))
}
