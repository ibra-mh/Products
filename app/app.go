package app

import (
	"database/sql"
	"log"
	"Products/utils"
	"net/http"
	"github.com/gorilla/mux"
)

func InitializeRoute(db *sql.DB) {
	r := mux.NewRouter()
	OfferRoutes(db, r)
	MaterialRoutes(db, r)
	OfferMaterialRoutes(db, r)

	// Start the server
	log.Fatal(http.ListenAndServe(":8003", utils.JsonContentTypeMiddleware(r))) // Running on port 8002
}
