package main

import (
	"database/sql"
	"log"
	"os"
	"Products/app" 
	_ "github.com/lib/pq"
)

func main() {
	// Open connection to the PostgreSQL database
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create the products, materials, and offer_material tables if they don't exist
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS offer (
			id SERIAL PRIMARY KEY,
			name VARCHAR NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS material (
			id SERIAL PRIMARY KEY,
			name VARCHAR NOT NULL,
			active BOOLEAN DEFAULT true,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS offer_material (
			id SERIAL PRIMARY KEY,
			offer_id INT NOT NULL REFERENCES offer(id),
			material_id INT NOT NULL REFERENCES material(id),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP
		);
	`)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize routes and start the server
	app.InitializeRoute(db)
}
