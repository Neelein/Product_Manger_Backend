.PHONY: db db-stop db-rm test test-integration

db:
	docker compose up -d db

db-stop:
	docker compose stop db

db-rm:
	docker compose down

test:
	go test -count=1 ./src/...

test-integration:
	go test -tags=integration -count=1 -p=1 -v ./src/test/...
