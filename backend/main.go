package main

import (
	"log"
	"net/http"

	"github.com/matsu0122-png/linkvault/backend/config"
	"github.com/matsu0122-png/linkvault/backend/database"
	"github.com/matsu0122-png/linkvault/backend/fetcher"
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
	linkService := service.NewLinkService(linkRepo, fetcher.New())
	linkHandler := handler.NewLinkHandler(linkService)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/links", linkHandler.ListLinks)
	mux.HandleFunc("POST /api/links", linkHandler.CreateLink)
	mux.HandleFunc("POST /api/links/bulk", linkHandler.BulkCreateLinks)
	mux.HandleFunc("POST /api/links/check", linkHandler.CheckLinks)
	mux.HandleFunc("PUT /api/links/{id}", linkHandler.UpdateLink)
	mux.HandleFunc("DELETE /api/links/{id}", linkHandler.DeleteLink)

	log.Println("LinkVault backend started on :8080")

	if err := http.ListenAndServe(":8080", withCORS(cfg.CORSAllowedOrigins, mux)); err != nil {
		log.Fatal(err)
	}
}

// withCORS reflects the request's Origin header only when it's in
// allowedOrigins, rather than always sending back one fixed value — this is
// what lets both the web frontend and the browser extension (whose
// chrome-extension://<id> origin differs per install) be allowed at once.
func withCORS(allowedOrigins []string, next http.Handler) http.Handler {
	allowed := make(map[string]bool, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		allowed[origin] = true
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if allowed[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
