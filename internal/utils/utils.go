package utils

import (
	"fmt"
	"os"
)

func GetDsnFromEnv() string {
	var dsn string
	if os.Getenv("ENV") == "PROD" {
		dsn = os.Getenv("SOME_SECRET_DSN")
	} else {
		addr := os.Getenv("DATABASE_URL")
		if addr != "" {
			dsn = fmt.Sprintf("postgres://postgres:postgres@%v:5432/postgres?sslmode=disable", addr)
		} else {
			dsn = "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
		}
	}
	return dsn
}
