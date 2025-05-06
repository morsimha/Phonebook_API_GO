package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"phoneBook/internal/models"

)

type Handler struct {
	db    *sql.DB
	cache *redis.Client
}

func NewHandler(db *sql.DB, cache *redis.Client) *Handler {
	return &Handler{db: db, cache: cache}
}

func (h *Handler) AddContact(w http.ResponseWriter, r *http.Request) {
	var c models.Contact
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	_, err := h.db.Exec(
		"INSERT INTO contacts (first_name, last_name, phone, address) VALUES ($1, $2, $3, $4)",
		c.FirstName, c.LastName, c.Phone, c.Address,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = h.cache.Del(context.Background(), "contacts_page_1")
	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) UpdateContact(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var c models.Contact
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	_, err := h.db.Exec(
		"UPDATE contacts SET first_name=$1, last_name=$2, phone=$3, address=$4 WHERE id=$5",
		c.FirstName, c.LastName, c.Phone, c.Address, id,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = h.cache.Del(context.Background(), "contacts_page_1")
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetContacts(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	page := r.URL.Query().Get("page")
	if page == "" {
		page = "1"
	}
	cacheKey := "contacts_page_" + page

	// ניסיון למשוך מהקאש
	cached, err := h.cache.Get(ctx, cacheKey).Result()
	if err == nil {
		log.Printf("[DEBUG] Cache hit for key: %s", cacheKey)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(cached))
		return
	}
	log.Printf("[DEBUG] Cache miss for key: %s", cacheKey)

	// שליפה מה־DB
	p, _ := strconv.Atoi(page)
	if p < 1 {
		p = 1
	}
	offset := (p - 1) * 10
	rows, err := h.db.Query("SELECT id, first_name, last_name, phone, address FROM contacts LIMIT 10 OFFSET $1", offset)
	if err != nil {
		log.Printf("[ERROR] DB query failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var contacts []models.Contact
	for rows.Next() {
		var c models.Contact
		err := rows.Scan(&c.ID, &c.FirstName, &c.LastName, &c.Phone, &c.Address)
		if err != nil {
			log.Printf("[ERROR] Failed to scan row: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		contacts = append(contacts, c)
	}

	res, err := json.Marshal(contacts)
	if err != nil {
		log.Printf("[ERROR] Failed to marshal contacts: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// שמירה בקאש
	if err := h.cache.Set(ctx, cacheKey, res, 30*time.Second).Err(); err != nil {
		log.Printf("[WARN] Failed to set cache for key: %s, err: %v", cacheKey, err)
	} else {
		log.Printf("[INFO] Data cached for key: %s", cacheKey)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(res)
}

func (h *Handler) SearchContacts(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	rows, err := h.db.Query(
		"SELECT id, first_name, last_name, phone, address FROM contacts WHERE first_name ILIKE $1 OR last_name ILIKE $2",
		"%"+query+"%", "%"+query+"%",
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var results []models.Contact
	for rows.Next() {
		var c models.Contact
		err := rows.Scan(&c.ID, &c.FirstName, &c.LastName, &c.Phone, &c.Address)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		results = append(results, c)
	}
	json.NewEncoder(w).Encode(results)
}

func (h *Handler) DeleteContact(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	_, err := h.db.Exec("DELETE FROM contacts WHERE id=$1", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = h.cache.Del(context.Background(), "contacts_page_1")
	w.WriteHeader(http.StatusOK)
}
