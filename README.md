# 📞 Phonebook API (Go)

A simple and scalable phonebook REST API built with **Golang**, **PostgreSQL**, **Redis**, **Docker Compose**, and **Prometheus** monitoring.

---

## 🧱 Features

- CRUD operations for contacts
- Search by first/last name (ILIKE)
- Pagination (`GET /contacts?page=N`)
- Redis caching for paginated contact queries
- Prometheus metrics at `/metrics`
- Dockerized setup with PostgreSQL + Redis + App
- Clean architecture (modular code, SRP-compliant)

---

## 🚀 Endpoints

| Method | Endpoint               | Description                |
|--------|------------------------|----------------------------|
| GET    | `/contacts?page=N`     | Paginated list of contacts |
| GET    | `/contacts/search?q=`  | Search contacts by name    |
| POST   | `/contacts`            | Add a new contact          |
| PUT    | `/contacts/{id}`       | Update existing contact    |
| DELETE | `/contacts/{id}`       | Delete contact             |
| GET    | `/metrics`             | Prometheus metrics         |

---

## 🐳 Running with Docker

```bash
docker-compose up --build
```

> App runs on `http://localhost:8080`

---

## ⚙️ Environment Variables

These are required (see `.env.example`):

```env
PORT=8080
DB_HOST=db
DB_PORT=5432
DB_USER=user
DB_PASSWORD=pass
DB_NAME=phonebook
REDIS_ADDR=redis:6379
```

---

## 📦 Project Structure

```
.
├── main.go               # Entrypoint (calls setup + start)
├── init_services.go      # initDB(), initRedis()
├── middleware.go         # Logging + Metrics middleware
├── handler.go            # REST API handlers
├── db/init.sql           # Schema + indexes
├── docker-compose.yml
└── .env.example
```

---

## 📈 Observability

Prometheus-compatible endpoint exposed at `/metrics`:
- Total requests
- Ready for Grafana dashboards
- Can be extended to measure latency, errors, etc.

---

## ✅ Bonus

- Redis caching for faster paginated fetches
- Logging: HTTP requests, cache hits/misses, DB failures
- Modular, testable structure

## ✅ Testing

This project includes a complete suite of **unit tests** for all API endpoints.

### 🔍 Technologies Used
- `sqlmock` for mocking PostgreSQL queries
- `httptest` for simulating HTTP requests/responses
- `testify/assert` for readable assertions

### 🧪 Covered Endpoints
- `POST /contacts`
- `PUT /contacts/{id}`
- `DELETE /contacts/{id}`
- `GET /contacts`
- `GET /contacts/search?name=X`

### 📈 Code Coverage
To measure test coverage and view a full HTML report:

```bash
go test -coverprofile=coverage.out
go tool cover -html=coverage.out
```

To see coverage percentage in CLI:

```bash
go test -cover
```

📊 Example output:
```
ok      phoneBook       1.9s   coverage: 39.3% of statements
```
