psql:
	docker compose -f docker-compose.yml exec -it db psql -U postgres -d usertest
setup:
	docker compose -f docker-compose.yml up -d
tearDown:
	docker compose -f docker-compose.yml down
tearDownTest:
	@echo 'Tearing down test container...'
	docker stop $(container_name)
	docker rm $(container_name)
migration:
	@echo 'Creating migration files for ${name}...'
	migrate create -ext=.sql -dir=./migrations ${name}
migrate_up:
	migrate -path=./migrations -database=postgres://postgres:postgres@localhost:5431/usertest\?sslmode=disable up
migrate_down:
	migrate -path=./migrations -database=postgres://postgres:postgres@localhost:5431/usertest\?sslmode=disable down
test:
	go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out
run:
	go run ./cmd/api
