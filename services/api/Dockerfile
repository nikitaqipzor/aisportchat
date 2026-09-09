FROM golang:1.23-bookworm AS build
RUN apt-get update && apt-get install -y --no-install-recommends gcc libpq-dev && rm -rf /var/lib/apt/lists/*
WORKDIR /src
COPY services/api/go.mod ./go.mod
COPY services/api ./
RUN CGO_ENABLED=1 go build -trimpath -ldflags="-s -w" -o /out/fitness-api ./cmd/api

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates libpq5 && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=build /out/fitness-api /app/fitness-api
ENV API_PORT=8080 STORE_BACKEND=postgres
EXPOSE 8080
CMD ["/app/fitness-api"]
