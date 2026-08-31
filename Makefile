.PHONY: db db-stop db-rm server server-stop test test-integration test-integration-productdb seed-storefront-e2e

db:
	docker compose up -d db

db-stop:
	docker compose stop db

db-rm:
	docker compose down

server:
	export $$(grep -v '^#' dotenv.env | xargs) && go run .

server-stop:
	pkill -f "go run|./server" || true

test:
	go test -count=1 ./src/...

test-integration:
	go test -tags=integration -count=1 -p=1 -v ./src/test/...

test-integration-productdb:
	@DATABASE_URL="$${DATABASE_URL:?Set DATABASE_URL to localhost productdb}" go test -tags=integration -count=1 -p=1 -v ./src/test/...

seed-storefront-e2e:
	go run ./src/test/seed_storefront
