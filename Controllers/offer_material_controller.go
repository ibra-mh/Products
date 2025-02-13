package controllers

import (
	"Products/models"
	"database/sql"
	"encoding/json"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func GetOfferMaterials(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT * FROM offer_material WHERE deleted_at IS NULL")
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		offerMaterials := []models.OfferMaterial{}
		for rows.Next() {
			var offerMaterial models.OfferMaterial
			if err := rows.Scan(&offerMaterial.ID, &offerMaterial.OfferID, &offerMaterial.MaterialID, &offerMaterial.CreatedAt, &offerMaterial.UpdatedAt, &offerMaterial.DeletedAt); err != nil {
				log.Fatal(err)
			}
			offerMaterials = append(offerMaterials, offerMaterial)
		}
		if err := rows.Err(); err != nil {
			log.Fatal(err)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(offerMaterials)
	}
}

func GetOfferMaterialByID(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		var offerMaterial models.OfferMaterial
		err := db.QueryRow("SELECT * FROM offer_material WHERE id = $1 AND deleted_at IS NULL", id).
			Scan(&offerMaterial.ID, &offerMaterial.OfferID, &offerMaterial.MaterialID, &offerMaterial.CreatedAt, &offerMaterial.UpdatedAt, &offerMaterial.DeletedAt)
		if err != nil {
			http.Error(w, "OfferMaterial not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(offerMaterial)
	}
}

func CreateOfferMaterial(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var offerMaterial models.OfferMaterial
		if err := json.NewDecoder(r.Body).Decode(&offerMaterial); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err := db.QueryRow("INSERT INTO offer_material (offer_id, material_id) VALUES ($1, $2) RETURNING id, created_at, updated_at", offerMaterial.OfferID, offerMaterial.MaterialID).
			Scan(&offerMaterial.ID, &offerMaterial.CreatedAt, &offerMaterial.UpdatedAt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(offerMaterial)
	}
}

func UpdateOfferMaterial(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		var offerMaterial models.OfferMaterial
		if err := json.NewDecoder(r.Body).Decode(&offerMaterial); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		_, err := db.Exec("UPDATE offer_material SET offer_id = $1, material_id = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $3 AND deleted_at IS NULL", offerMaterial.OfferID, offerMaterial.MaterialID, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(offerMaterial)
	}
}

func DeleteOfferMaterial(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		_, err := db.Exec("UPDATE offer_material SET deleted_at = CURRENT_TIMESTAMP WHERE id = $1", id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
