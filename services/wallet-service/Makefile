.PHONY: run test docker-up docker-down

run:
	@set -a; . ./.env; set +a; \
	go run cmd/monolith/monolith.go

docker-up:
	docker compose -f ../../sandbox/docker-compose.yaml up -d

docker-down:
	docker compose -f ../../sandbox/docker-compose.yaml down -v
