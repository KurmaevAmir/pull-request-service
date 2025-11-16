FROM golang:1.25 AS build
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/pr-service .

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=build /bin/pr-service usr/local/bin/pr-service
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/pr-service"]