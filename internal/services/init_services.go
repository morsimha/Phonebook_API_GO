package services

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	_ "github.com/lib/pq"
)

func InitDB() (*sql.DB, error) {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("sql.Open error: %w", err)
	}

// Set connection pool settings after opening the connection
	for i := 1; i <= 10; i++ {
		err = db.Ping()
		if err == nil {
			break
		}
		fmt.Printf("Waiting for DB to be ready... attempt %d/10\n", i)
		time.Sleep(1 * time.Second)
	}
	if err != nil {
		return nil, fmt.Errorf("db.Ping error: %w", err)
	}

	return db, nil
}

func InitRedis() (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: "",
		DB:       0,
	})
	ctx := context.Background()
	var err error
	for i := 1; i <= 10; i++ {
		_, err = rdb.Ping(ctx).Result()
		if err == nil {
			break
		}
		fmt.Printf("Waiting for Redis to be ready... attempt %d/10\n", i)
		time.Sleep(1 * time.Second)
	}
	return rdb, err
}
