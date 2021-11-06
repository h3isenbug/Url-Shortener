FROM golang:1.17-alpine3.14 AS build
WORKDIR /src
RUN apk add gcc musl-dev curl
RUN go install github.com/google/wire/cmd/wire@latest
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.15.1/migrate.linux-amd64.tar.gz | tar xvz migrate && mv migrate /usr/bin

COPY go.mod go.sum ./
RUN go mod download -x

COPY . ./
RUN wire $PWD/cmd/url-shortener/di
RUN go build -o url-shortener $PWD/cmd/url-shortener/
CMD ["./docker-entrypoint.sh", "test"]

FROM alpine:3.14 AS execute
RUN apk --no-cache add ca-certificates curl
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.15.1/migrate.linux-amd64.tar.gz | tar xvz migrate && mv migrate /usr/bin
WORKDIR /srv
COPY --from=build /src/url-shortener .
COPY docker-entrypoint.sh secrets.test.yaml ./
COPY migrations migrations
CMD ["./docker-entrypoint.sh", "run"]
