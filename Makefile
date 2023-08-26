psql:
	docker compose -f docker-compose.yml exec -it db psql -U postgres -d usertest
setup:
	docker compose -f docker-compose.yml up -d