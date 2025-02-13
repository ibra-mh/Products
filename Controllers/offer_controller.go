package controllers

import (
	"Products/models"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func GetOffers(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT * FROM offer WHERE deleted_at IS NULL")
		if err != nil {
			log.Printf("Error querying database: %v", err) // Use log.Printf instead of Fatal
			http.Error(w, "database error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		offers := []models.Offer{}
		for rows.Next() {
			var offer models.Offer
			if err := rows.Scan(&offer.ID, &offer.Name, &offer.CreatedAt, &offer.UpdatedAt, &offer.DeletedAt); err != nil {
				log.Printf("Error scanning rows: %v", err)
				http.Error(w, "error processing database results", http.StatusInternalServerError)
				return
			}
			offers = append(offers, offer)
		}
		if err := rows.Err(); err != nil {
			log.Printf("Error iterating rows: %v", err)
			http.Error(w, "error iterating over results", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(offers)
	}
}

// func GetOffers(db *sql.DB) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		rows, err := db.Query("SELECT * FROM offer WHERE deleted_at IS NULL")
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 		defer rows.Close()

// 		offers := []models.Offer{}
// 		for rows.Next() {
// 			var offer models.Offer
// 			if err := rows.Scan(&offer.ID, &offer.Name, &offer.CreatedAt, &offer.UpdatedAt, &offer.DeletedAt); err != nil {
// 				log.Fatal(err)
// 			}
// 			offers = append(offers, offer)
// 		}
// 		if err := rows.Err(); err != nil {
// 			log.Fatal(err)
// 		}

// 		w.Header().Set("Content-Type", "application/json")
// 		json.NewEncoder(w).Encode(offers)
// 	}
// }

func GetOfferByID(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		var offer models.Offer
		err := db.QueryRow("SELECT * FROM offer WHERE id = $1 AND deleted_at IS NULL", id).
			Scan(&offer.ID, &offer.Name, &offer.CreatedAt, &offer.UpdatedAt, &offer.DeletedAt)
		if err != nil {
			http.Error(w, "Offer not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(offer)
	}
}

// func CreateOffer(db *sql.DB) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		var offer models.Offer
// 		if err := json.NewDecoder(r.Body).Decode(&offer); err != nil {
// 			http.Error(w, err.Error(), http.StatusBadRequest)
// 			return
// 		}

// 		err := db.QueryRow("INSERT INTO offer (name) VALUES ($1) RETURNING id, created_at, updated_at", offer.Name).
// 			Scan(&offer.ID, &offer.CreatedAt, &offer.UpdatedAt)
// 		if err != nil {
// 			http.Error(w, err.Error(), http.StatusInternalServerError)
// 			return
// 		}

//			w.Header().Set("Content-Type", "application/json")
//			json.NewEncoder(w).Encode(offer)
//		}
//	}
func CreateOffer(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var offer models.Offer
		if err := json.NewDecoder(r.Body).Decode(&offer); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err := db.QueryRow("INSERT INTO offer (name) VALUES ($1) RETURNING id, created_at, updated_at", offer.Name).
			Scan(&offer.ID, &offer.CreatedAt, &offer.UpdatedAt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated) // Explicitly set the status code to 201
		json.NewEncoder(w).Encode(offer)
	}
}

func UpdateOffer(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		var offer models.Offer
		if err := json.NewDecoder(r.Body).Decode(&offer); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		_, err := db.Exec("UPDATE offer SET name = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2 AND deleted_at IS NULL", offer.Name, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(offer)
	}
}

// func DeleteOffer(db *sql.DB) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		vars := mux.Vars(r)
// 		id := vars["id"]

// 		_, err := db.Exec("UPDATE offer SET deleted_at = CURRENT_TIMESTAMP WHERE id = $1", id)
// 		if err != nil {
// 			http.Error(w, err.Error(), http.StatusInternalServerError)
// 			return
// 		}

//			w.WriteHeader(http.StatusNoContent)
//		}
//	}
func DeleteOffer(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		idStr, exists := vars["id"]
		if !exists {
			http.Error(w, "missing id", http.StatusBadRequest)
			return
		}

		// Convert ID to int
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}

		// Execute soft delete query
		res, err := db.Exec("UPDATE offer SET deleted_at = CURRENT_TIMESTAMP WHERE id = $1", id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Check if any row was affected
		rowsAffected, err := res.RowsAffected()
		if err != nil || rowsAffected == 0 {
			http.Error(w, "Offer not found", http.StatusNotFound)
			return
		}

		// Return 204 No Content if deletion was successful
		w.WriteHeader(http.StatusNoContent)
	}
}
