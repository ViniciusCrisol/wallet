.PHONY: run test populate docker-up docker-down

run:
	@set -a; . ./.env; set +a; \
	go run cmd/monolith/monolith.go

test:
	go clean -testcache && go test ./...

populate:
	node ../../sandbox/populate-wallet.js

docker-up:
	docker compose -f ../../sandbox/docker-compose.yaml up -d && \
	docker compose -f ../../sandbox/docker-compose.test.yaml up -d

docker-down:
	docker compose -f ../../sandbox/docker-compose.yaml down -v && \
	docker compose -f ../../sandbox/docker-compose.test.yaml down -v
