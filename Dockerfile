FROM golang:1.22-alpine

RUN apk add --no-cache git gcc musl-dev

WORKDIR /app

COPY go.mod go.sum* ./
RUN go mod download 2>/dev/null || true

COPY . .

CMD ["go", "build", "-o", "bin/ariavox", "./cmd/ariavox"]
