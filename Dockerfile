FROM golang:1.26.1-alpine AS build

WORKDIR /src

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/api ./cmd/api

FROM alpine:3.22

WORKDIR /app

COPY --from=build /out/api /app/api
COPY migrations /app/migrations

EXPOSE 8080

CMD ["/app/api"]
