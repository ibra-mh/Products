package app

import (
	"database/sql"
	"Products/Controllers"
	"github.com/gorilla/mux"
)

func OfferRoutes(db *sql.DB, r *mux.Router) {
	// Offer Routes
	r.HandleFunc("/offers", controllers.GetOffers(db)).Methods("GET")
	r.HandleFunc("/offers/{id}", controllers.GetOfferByID(db)).Methods("GET")
	r.HandleFunc("/offers", controllers.CreateOffer(db)).Methods("POST")
	r.HandleFunc("/offers/{id}", controllers.UpdateOffer(db)).Methods("PUT")
	r.HandleFunc("/offers/{id}", controllers.DeleteOffer(db)).Methods("DELETE")
}
