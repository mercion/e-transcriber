FROM golang:1.21-bookworm AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 go build -o /out/e-transcriber .

FROM debian:bookworm-slim

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=build /out/e-transcriber /app/e-transcriber
COPY web /app/web
RUN mkdir -p /app/models

ENV ADDR=:8080 \
    WHISPER_MODELS_DIR=/app/models

EXPOSE 8080
CMD ["/app/e-transcriber"]
