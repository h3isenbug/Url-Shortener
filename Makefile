help:
	@echo "$$HELP_TEXT"

dependencies:
	go mod download -x

dev-dependencies:
	go install github.com/golang/mock/mockgen@v1.6.0
	go install github.com/google/wire/cmd/wire@latest
	go mod download -x

build:
	wire $$PWD/cmd/url-shortener/di
	go build -o url-shortener $$PWD/cmd/url-shortener/
docker:
	docker build . -t h3isenbug/url-shortener:build --target build --cache-from h3isenbug/url-shortener:build
	docker build . -t h3isenbug/url-shortener:execute --target execute --cache-from h3isenbug/url-shortener:build --cache-from h3isenbug/url-shortener:execute

ACTION=up
N=1
migrate:
	migrate -source "file:/$(shell realpath migrations)" -database "postgres://$$DB_USER:$$DB_PASSWORD@$$DB_HOST:$$DB_PORT/$$DB_NAME?application_name=url_shortener-migrations-$$DEPLOY_TAG-$$HOSTNAME" $(ACTION) $(N)

mocks:
	mockgen -source internal/repository/refreshToken/refreshToken.go  > internal/repository/refreshToken/mock/refreshToken.go
	mockgen -source internal/repository/url/url.go  > internal/repository/url/mock/url.go
	mockgen -source internal/repository/account/account.go  > internal/repository/account/mock/account.go

test:
	docker-compose -f docker-compose.test.yaml rm -fsv
	docker-compose -f docker-compose.test.yaml up --build --remove-orphans --abort-on-container-exit --exit-code-from url-shortener

run:
	docker-compose -f docker-compose.yaml up --build --remove-orphans --abort-on-container-exit

define HELP_TEXT
Use the following commands:
	make dependencies
		install dependencies needed for execution of this project

	make dev-dependencies
		install dependencies needed for development of this project

	make build
		build the project binary

	make docker
		build docker images for this project

	make migrate [up|down|force] N
		run migrations. default is ACTION=up with N=1

	make mock
		generate mock stubs

	make test
		run tests using docker compose

	make run
		run project using docker compose

	make help
		print this manual

endef

export HELP_TEXT