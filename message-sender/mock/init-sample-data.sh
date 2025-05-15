#!/bin/bash
set -e

echo "Initializing message-sender sample data..."

# Use environment variables or default values
POSTGRES_HOST="${POSTGRES_HOST:-localhost}"
POSTGRES_USER="${POSTGRES_USER:-postgres}"
POSTGRES_DB="${POSTGRES_DB:-message_sender}"

# Get database password from vault or use default
if [ -f "./mock/vault-secrets.sh" ]; then
  DB_PASSWORD=$(./mock/vault-secrets.sh)
else
  DB_PASSWORD="${POSTGRES_PASSWORD:-postgres}"
  echo "Using default or environment password"
fi

echo "Connecting to PostgreSQL at $POSTGRES_HOST as $POSTGRES_USER"

# Wait for PostgreSQL to be ready
until PGPASSWORD=$DB_PASSWORD psql -h "$POSTGRES_HOST" -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c '\q'; do
  echo "PostgreSQL unavailable - waiting..."
  sleep 2
done

echo "PostgreSQL is ready"

# Create messages table if it doesn't exist
PGPASSWORD=$DB_PASSWORD psql -h "$POSTGRES_HOST" -U "$POSTGRES_USER" -d "$POSTGRES_DB" <<EOF
CREATE TABLE IF NOT EXISTS messages (
    id SERIAL PRIMARY KEY,
    content VARCHAR(160) NOT NULL,
    recipient VARCHAR(15) NOT NULL,
    is_sent BOOLEAN DEFAULT FALSE,
    sent_at TIMESTAMP,
    message_id VARCHAR(36)
);

-- Ensure the extension is available
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Insert sample messages
INSERT INTO messages (content, recipient, is_sent, sent_at, message_id)
VALUES 
    ('Message 1', '+905500000000', false, NULL, NULL),
    ('Message 2', '+905500000001', false, NULL, NULL),
    ('Message 3', '+905500000002', false, NULL, NULL);

-- Insert sample messages with different delivery times
INSERT INTO messages (content, recipient, is_sent, sent_at, message_id)
VALUES 
    ('Message 1', '+905500000000', true, NOW() - INTERVAL '2 days', 'msg-id-1'),
    ('Message 2', '+905500000000', true, NOW() - INTERVAL '1 day', 'msg-id-2');
EOF

echo "Sample data initialization completed!" 