module bangungrid/services/catalog-service

go 1.21.0

require (
	bangungrid/pkg v0.0.0
	github.com/go-chi/chi/v5 v5.0.10
	github.com/google/uuid v1.3.1
	github.com/jackc/pgx/v5 v5.4.3
	github.com/joho/godotenv v1.5.1
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	golang.org/x/crypto v0.13.0 // indirect
	golang.org/x/sync v0.3.0 // indirect
	golang.org/x/text v0.13.0 // indirect
)

replace bangungrid/pkg => ../../pkg
