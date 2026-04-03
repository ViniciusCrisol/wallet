.PHONY: run test populate docker-up docker-down

run:
	@set -a; . ./.env; set +a; \
	go run cmd/monolith/monolith.go

populate:
	node ../../sandbox/populate-wallet.js

docker-up:
	docker compose -f ../../sandbox/docker-compose.yaml up -d

docker-down:
	docker compose -f ../../sandbox/docker-compose.yaml down -v
