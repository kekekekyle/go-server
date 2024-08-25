module github.com/kekekekyle/go-server

go 1.23.0

replace github.com/kekekekyle/database => ./internal/database

require (
  github.com/golang-jwt/jwt/v5 v5.2.1 // indirect
	github.com/joho/godotenv v1.5.1
	github.com/kekekekyle/database v0.0.0
	golang.org/x/crypto v0.26.0
)

