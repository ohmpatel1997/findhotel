#!make

build-app:
	docker build -f cmd/client-api/Dockerfile .

run-app:
	docker-compose build
	docker-compose up -d database server migration

reset-app:
	docker-compose down

import:
	docker-compose up -d database migration importer