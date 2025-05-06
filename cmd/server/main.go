package main

import (
	"log"
	"net/http"
	"os"

	"phoneBook/api/handlers"
	"phoneBook/api/middleware"
	"phoneBook/internal/services"
	"phoneBook/pkg/metrics"
	
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	if err := initConfig(); err != nil {
		log.Fatalf("[ERROR] Failed to load environment: %v", err)
	}

	db, err := services.InitDB()
	if err != nil {
		log.Fatalf("[ERROR] Failed to initialize database: %v", err)
	}
	log.Println("[INFO] Successfully connected to PostgreSQL")

	rdb, err := services.InitRedis()
	if err != nil {
		log.Fatalf("[ERROR] Failed to connect to Redis: %v", err)
	}
	log.Println("[INFO] Successfully connected to Redis")

	log.Println("[DEBUG] Redis connection established")
	h := handlers.NewHandler(db, rdb)
	r := setupRouter(h)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("[INFO] Server is listening on port %s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("[ERROR] Server failed to start: %v", err)
	}
}

func initConfig() error {
	if err := godotenv.Load(); err != nil {
		log.Println("[WARN] .env file not found, using system environment variables.")
		return nil
	}
	log.Println("[INFO] Loaded environment variables from .env")
	return nil
}

func setupRouter(h *handlers.Handler) *mux.Router {
	prometheus.MustRegister(metrics.RequestCount)
	r := mux.NewRouter()
	r.Use(middleware.LoggingMiddleware)
	r.Use(middleware.MetricsMiddleware)

	r.HandleFunc("/contacts", func(w http.ResponseWriter, r *http.Request) {
		log.Println("[INFO] Handling GET /contacts")
		h.GetContacts(w, r)
	}).Methods("GET")

	r.HandleFunc("/contacts/search", h.SearchContacts).Methods("GET")
	r.HandleFunc("/contacts", h.AddContact).Methods("POST")
	r.HandleFunc("/contacts/{id}", h.UpdateContact).Methods("PUT")
	r.HandleFunc("/contacts/{id}", h.DeleteContact).Methods("DELETE")
	r.Handle("/metrics", promhttp.Handler()).Methods("GET")
	return r
}
