package app

import (
	"database/sql"
	"Products/Controllers"
	"github.com/gorilla/mux"
)

func MaterialRoutes(db *sql.DB, r *mux.Router) {
	// Material Routes
	r.HandleFunc("/materials", controllers.GetMaterials(db)).Methods("GET")
	r.HandleFunc("/materials/{id}", controllers.GetMaterialByID(db)).Methods("GET")
	r.HandleFunc("/materials", controllers.CreateMaterial(db)).Methods("POST")
	r.HandleFunc("/materials/{id}", controllers.UpdateMaterial(db)).Methods("PUT")
	r.HandleFunc("/materials/{id}", controllers.DeleteMaterial(db)).Methods("DELETE")
}
