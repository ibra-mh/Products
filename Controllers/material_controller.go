package controllers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"github.com/gorilla/mux"
	"Products/models"
)

// GetMaterials retrieves all materials
func GetMaterials(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT * FROM material WHERE deleted_at IS NULL")
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		materials := []models.Material{}
		for rows.Next() {
			var material models.Material
			if err := rows.Scan(&material.ID, &material.Name, &material.Active, &material.CreatedAt, &material.UpdatedAt, &material.DeletedAt); err != nil {
				log.Fatal(err)
			}
			materials = append(materials, material)
		}
		if err := rows.Err(); err != nil {
			log.Fatal(err)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(materials)
	}
}

// GetMaterialByID retrieves a material by ID
func GetMaterialByID(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		var material models.Material
		err := db.QueryRow("SELECT * FROM material WHERE id = $1 AND deleted_at IS NULL", id).
			Scan(&material.ID, &material.Name, &material.Active, &material.CreatedAt, &material.UpdatedAt, &material.DeletedAt)
		if err != nil {
			http.Error(w, "Material not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(material)
	}
}

// CreateMaterial creates a new material
func CreateMaterial(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var material models.Material
		if err := json.NewDecoder(r.Body).Decode(&material); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err := db.QueryRow("INSERT INTO material (name, active) VALUES ($1, $2) RETURNING id, created_at, updated_at", material.Name, material.Active).
			Scan(&material.ID, &material.CreatedAt, &material.UpdatedAt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(material)
	}
}

// UpdateMaterial updates an existing material
func UpdateMaterial(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		var material models.Material
		if err := json.NewDecoder(r.Body).Decode(&material); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		_, err := db.Exec("UPDATE material SET name = $1, active = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $3 AND deleted_at IS NULL", material.Name, material.Active, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(material)
	}
}

// DeleteMaterial deletes a material (soft delete)
func DeleteMaterial(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		_, err := db.Exec("UPDATE material SET deleted_at = CURRENT_TIMESTAMP WHERE id = $1", id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
