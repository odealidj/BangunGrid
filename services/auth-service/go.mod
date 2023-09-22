module bangungrid/services/auth-service

go 1.21.0

require (
	bangungrid/pkg v0.0.0
	github.com/go-chi/chi/v5 v5.0.10
	github.com/golang-jwt/jwt/v5 v5.2.1
	github.com/google/uuid v1.3.1
	github.com/jackc/pgx/v5 v5.4.3
	github.com/joho/godotenv v1.5.1
	github.com/nats-io/nats.go v1.28.0
	github.com/redis/go-redis/v9 v9.1.0
	golang.org/x/crypto v0.13.0
)

require (
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/klauspost/compress v1.17.2 // indirect
	github.com/nats-io/nkeys v0.4.4 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	golang.org/x/sync v0.3.0 // indirect
	golang.org/x/sys v0.12.0 // indirect
	golang.org/x/text v0.13.0 // indirect
)

replace bangungrid/pkg => ../../pkg
