package app

import (
	"database/sql"
	"Products/Controllers"
	"github.com/gorilla/mux"
)

func OfferMaterialRoutes(db *sql.DB, r *mux.Router) {
	// OfferMaterial Routes
	r.HandleFunc("/offer-materials", controllers.GetOfferMaterials(db)).Methods("GET")
	r.HandleFunc("/offer-materials/{id}", controllers.GetOfferMaterialByID(db)).Methods("GET")
	r.HandleFunc("/offer-materials", controllers.CreateOfferMaterial(db)).Methods("POST")
	r.HandleFunc("/offer-materials/{id}", controllers.UpdateOfferMaterial(db)).Methods("PUT")
	r.HandleFunc("/offer-materials/{id}", controllers.DeleteOfferMaterial(db)).Methods("DELETE")
}
