FROM golang:1.25-bookworm

WORKDIR /app

# curl 给 compose healthcheck 用；Air 负责 Go 源码热重载。
RUN apt-get update \
    && apt-get install -y --no-install-recommends curl \
    && rm -rf /var/lib/apt/lists/* \
    && go install github.com/air-verse/air@v1.67.4

COPY go.mod go.sum ./
RUN go mod download

COPY . .

EXPOSE 8000

CMD ["air", "-c", ".air.toml"]
