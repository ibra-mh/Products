package main

import (
	"Products/app"
	"log"
	"Products/config"

	_ "github.com/lib/pq"
)

func main() {
	config.ConnectDB()
	defer config.CloseDB()

	
	app.InitializeRoute(config.DB) 
	log.Println("Server started")  

}

