check_defined = \
    $(strip $(foreach 1,$1, \
        $(call __check_defined,$1,$(strip $(value 2)))))
__check_defined = \
    $(if $(value $1),, \
      $(error Undefined $1$(if $2, ($2))))

cache_events:
	go run main.go cache_events
## to generate a migration file
## Eg: make migration NAME=initBookingSchema
migration:
	$(call check_defined, NAME)
	./scripts/migration_gen.sh -f '$(NAME)'

## to migrate all the sql migration files
migrate:
	go run main.go migrate

## to rollback the latest version of sql migration
migrate.rollback:
	go run main.go migrate --rollback

# to run the http server
run:
	go run main.go http_server

air:
	air http_server

start:
	docker compose up postgres && docker compose up redis