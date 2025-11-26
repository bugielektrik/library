build:
	docker build -t library-service .

up:
	docker-compose up -d

down:
	docker-compose down

restart: down up

.PHONY: build up down restart
