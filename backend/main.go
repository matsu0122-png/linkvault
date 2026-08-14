package main

import (
	"log"
	"net/http"

	"github.com/matsu0122-png/linkvault/backend/config"
	"github.com/matsu0122-png/linkvault/backend/database"
	"github.com/matsu0122-png/linkvault/backend/handler"
	"github.com/matsu0122-png/linkvault/backend/repository"
	"github.com/matsu0122-png/linkvault/backend/service"
)

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	linkRepo := repository.NewLinkRepository(db)
	linkService := service.NewLinkService(linkRepo)
	linkHandler := handler.NewLinkHandler(linkService)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/links", linkHandler.ListLinks)
	mux.HandleFunc("POST /api/links", linkHandler.CreateLink)
	mux.HandleFunc("PUT /api/links/{id}", linkHandler.UpdateLink)
	mux.HandleFunc("DELETE /api/links/{id}", linkHandler.DeleteLink)

	log.Println("LinkVault backend started on :8080")

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
