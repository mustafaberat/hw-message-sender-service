FROM golang:1.21-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o message-sender .

FROM alpine:3.19

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata postgresql-client bash

COPY --from=builder /app/message-sender .
COPY --from=builder /app/mock ./mock/

RUN chmod +x ./mock/init-sample-data.sh ./mock/vault-secrets.sh

ENV TZ=UTC
ENV POSTGRES_HOST=postgres
ENV POSTGRES_USER=postgres
ENV POSTGRES_DB=message_sender

EXPOSE 8080

ENTRYPOINT ["sh", "-c", "./mock/init-sample-data.sh && ./message-sender"]
