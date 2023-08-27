psql:
	docker compose -f docker-compose.yml exec -it db psql -U postgres -d usertest
setup:
	docker compose -f docker-compose.yml up -d
migrate_up:
	migrate -path=./migrations -database=postgres://postgres:postgres@localhost:5431/usertest\?sslmode=disable up
migrate_down:
	migrate -path=./migrations -database=postgres://postgres:postgres@localhost:5431/usertest\?sslmode=disable down
test:
	go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out
