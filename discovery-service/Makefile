.PHONY: up down drop_random_container

up:
	@docker-compose up -d
down:
	@docker-compose down
drop_random_container:
	@docker-compose stop discovery-$$(($$RANDOM % 3 + 1)); \
	docker-compose rm -f discovery-$$(($$RANDOM % 3 + 1))