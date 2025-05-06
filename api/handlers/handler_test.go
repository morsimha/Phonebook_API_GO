package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestAddContact(t *testing.T) {
	// Mock DB
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	// Mock Redis (just a dummy client here, no commands used in AddContact)
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	h := NewHandler(db, rdb)

	// Set up expected insert query
	mock.ExpectExec("INSERT INTO contacts").
		WithArgs("John", "Doe", "1234567890", "Test Street").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Prepare JSON body
	jsonData := []byte(`{
        "first_name": "John",
        "last_name": "Doe",
        "phone": "1234567890",
        "address": "Test Street"
    }`)

	req := httptest.NewRequest("POST", "/contacts", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.AddContact(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestAddContact_InvalidJSON(t *testing.T) {
	db, _, _ := sqlmock.New()
	defer db.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	h := NewHandler(db, rdb)

	// Invalid JSON
	badJSON := []byte(`{ "first_name": "MissingQuote }`)
	req := httptest.NewRequest("POST", "/contacts", bytes.NewBuffer(badJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.AddContact(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateContact(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	h := NewHandler(db, rdb)

	// Expect the update query
	mock.ExpectExec("UPDATE contacts").
		WithArgs("Jane", "Smith", "987654321", "New Address", "1").
		WillReturnResult(sqlmock.NewResult(1, 1))

	jsonData := []byte(`{
        "first_name": "Jane",
        "last_name": "Smith",
        "phone": "987654321",
        "address": "New Address"
    }`)

	req := httptest.NewRequest("PUT", "/contacts/1", bytes.NewBuffer(jsonData))
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.UpdateContact(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDeleteContact(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	h := NewHandler(db, rdb)

	// Expect delete query
	mock.ExpectExec("DELETE FROM contacts").
		WithArgs("1").
		WillReturnResult(sqlmock.NewResult(1, 1))

	req := httptest.NewRequest("DELETE", "/contacts/1", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	w := httptest.NewRecorder()

	h.DeleteContact(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
func TestSearchContacts(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	h := NewHandler(db, rdb)

	// Mock expected query
	mock.ExpectQuery("SELECT id, first_name, last_name, phone, address FROM contacts WHERE").
		WithArgs("%John%", "%John%").
		WillReturnRows(sqlmock.NewRows([]string{"id", "first_name", "last_name", "phone", "address"}).
			AddRow(1, "John", "Doe", "1234567890", "Test Street").
			AddRow(2, "Johnny", "Smith", "987654321", "Another Street"))

	req := httptest.NewRequest("GET", "/contacts/search?q=John", nil)
	w := httptest.NewRecorder()

	h.SearchContacts(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `[{"id":1,"first_name":"John","last_name":"Doe","phone":"1234567890","address":"Test Street"},{"id":2,"first_name":"Johnny","last_name":"Smith","phone":"987654321","address":"Another Street"}]`, w.Body.String())
}

func TestGetContacts(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	h := NewHandler(db, rdb)

	rows := sqlmock.NewRows([]string{"id", "first_name", "last_name", "phone", "address"}).
		AddRow(1, "John", "Doe", "1234567890", "Test Address").
		AddRow(2, "Jane", "Smith", "9876543210", "Another Address")

	mock.ExpectQuery("SELECT .* FROM contacts .*").
		WillReturnRows(rows)

	req := httptest.NewRequest("GET", "/contacts?page=1", nil)
	w := httptest.NewRecorder()

	h.GetContacts(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}


func TestUpdateContact_MissingID(t *testing.T) {
    db, _, _ := sqlmock.New()
    defer db.Close()

    rdb := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
    })

    h := NewHandler(db, rdb)

    jsonData := []byte(`{
        "first_name": "Jane",
        "last_name": "Smith",
        "phone": "987654321",
        "address": "New Address"
    }`)

    req := httptest.NewRequest("PUT", "/contacts/", bytes.NewBuffer(jsonData))
    // Missing mux.SetURLVars
    w := httptest.NewRecorder()

    h.UpdateContact(w, req)

    assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteContact_MissingID(t *testing.T) {
    db, _, _ := sqlmock.New()
    defer db.Close()

    rdb := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
    })

    h := NewHandler(db, rdb)

    req := httptest.NewRequest("DELETE", "/contacts/", nil)
    // Missing mux.SetURLVars
    w := httptest.NewRecorder()

    h.DeleteContact(w, req)

    assert.Equal(t, http.StatusBadRequest, w.Code)
}
