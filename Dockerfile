FROM golang:1.25-alpine AS builder

WORKDIR /app
ENV GOTOOLCHAIN=auto
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /glassbase ./cmd/server

FROM chromedp/headless-shell:latest AS chrome

FROM alpine:3.19

WORKDIR /app

COPY --from=chrome /headless-shell /headless-shell

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /glassbase /app/glassbase
COPY migrations/ /app/migrations/

ENV CHROME_BIN=/headless-shell/headless-shell

EXPOSE 8080
CMD ["/app/glassbase"]
