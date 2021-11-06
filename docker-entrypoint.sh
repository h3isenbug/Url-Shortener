#!/usr/bin/env sh

test() {
  while ! nc -z "$DB_HOST" 5432; do sleep .5; done
  migrate -source "file://migrations" -database "postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?application_name=url_shortener-migrations-$DEPLOY_TAG-$HOSTNAME&sslmode=disable" up
  go test -race -coverprofile=coverage.out -coverpkg=./... ./...
  TEST_RESULT=$?

  go tool cover -html=coverage.out -o coverage.html

  return $TEST_RESULT
}

run() {
  while ! nc -z "$DB_HOST" 5432; do sleep .5; done

  ./url-shortener
}

"$@"
