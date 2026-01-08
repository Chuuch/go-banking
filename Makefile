DB_URL=postgresql://root:secret@localhost:5432/go_banking?sslmode=disable

postgres:
	docker run --name postgres16 --network banking-network -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:16-alpine

createdb:
	docker exec -it postgres16 createdb --username=root --owner=root go_banking

dropdb:
	docker exec -it postgres16 dropdb go_banking

migrateup:
	migrate -path db/migration -database "$(DB_URL)" -verbose up

migrateup1:
	migrate -path db/migration -database "$(DB_URL)" -verbose up 1

migratedown:
	migrate -path db/migration -database "$(DB_URL)" -verbose down

migratedown1:
	migrate -path db/migration -database "$(DB_URL)" -verbose down 1

sqlc:
	sqlc generate

test:
	go test -v -cover ./...

server:
	go run main.go

mock:
	mockgen --package mockdb --destination db/mock/store.go  github.com/chuuch/go-banking/db/sqlc Store

.PHONY: postgres createdb dropdb migrateup migrateup1 migratedown1 migratedown sqlc test server mock 
