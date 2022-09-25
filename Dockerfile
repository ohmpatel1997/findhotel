FROM  golang:1.17.3-alpine3.13 AS builder

WORKDIR code

RUN echo $(pwd)

COPY go.mod .
COPY go.sum .

RUN go mod download

ADD . .

# build the app binary
RUN env CGO_ENABLED=0 GOOS=linux  go build -o /app cmd/client-api/main.go

# build the importer binary
RUN env CGO_ENABLED=0 GOOS=linux  go build -o /import cmd/import/main.go

# build the migration binary
RUN env CGO_ENABLED=0 GOOS=linux go build -o /migration migration/main.go


# final stage
FROM alpine:3.16.2

RUN apk add curl
COPY --from=builder /app /
COPY --from=builder /import /
COPY --from=builder /migration /

COPY migration/geolocation /geolocation
COPY cmd/client-api/config.yaml /cmd/client-api/
COPY cmd/import/config.yaml /cmd/import/
COPY wait-for.sh /
COPY cmd/import/*.csv /cmd/import/

RUN chmod +x /app
RUN chmod +x /import
RUN chmod +x /migration
