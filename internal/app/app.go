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

	// Create test db entry for test
	// database.CreateTestArticles(db)

	// Headers
	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With"})
	originsOk := handlers.AllowedOrigins([]string{"http://localhost:4100", "http://0.0.0.0:4100"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})

	// Start serving
	log.Fatal(http.ListenAndServe(":8000", handlers.CORS(originsOk, headersOk, methodsOk)(MakeWebHandler(db))))
}
