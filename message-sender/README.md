# Message Sender Service

This is a simple microservice that sends messages through webhook APIs. It handles all the tricky stuff like batching, retries, and keeping track of message status.

## Setup

1. Run dependencies:

```
make docker-run-dependencies
```

## Endpoints

- `POST /api/service` - Start or stop the service
- `GET /api/messages/sent` - See what messages have been sent (with pagination)
- `GET /health` - Check if everything's working
- `GET /swagger/*` - Browse the API documentation

## Usage

### Access Swagger UI

Open Swagger in your browser: [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)

### View Messages

To view mock messages that are automatically written to the database when Docker Compose is running:

```
curl -X 'GET' \
  'http://localhost:8080/api/messages/sent' \
  -H 'accept: application/json'
```

### Start Message Sender Service

```
curl -X 'POST' \
  'http://localhost:8080/api/service' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "action": "start"
}'
```

Response:

```
{
  "status": "success",
  "message": "Service started successfully"
}
```

### Stop Message Sender Service

```
curl -X 'POST' \
  'http://localhost:8080/api/service' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "action": "stop"
}'
```

Response:

```
{
  "status": "success",
  "message": "Service stopped successfully"
}
```

### Check Service Health

```
curl -X 'GET' \
  'http://localhost:8080/health' \
  -H 'accept: application/json'
```

Response:

```
{
  "service": "running",
  "status": "ok",
  "time": "2025-05-15T17:11:55+03:00"
}
```

## Technical Details

- **Webhook Testing**: [https://webhook.site/](https://webhook.site/) is used for testing webhooks. You can view request and response details there.
- **Redis**: Used to track message status before database persistence
- **Database**: Connection password is simulated using Vault at the Docker level
