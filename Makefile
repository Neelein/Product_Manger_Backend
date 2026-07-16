.PHONY: db db-stop db-rm

db:
	docker compose up -d db

db-stop:
	docker compose stop db

db-rm:
	docker compose down -v
