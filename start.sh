#!/bin/sh

set -e
# Only source the file if it exists (prevents the 'not found' error)
if [ -f "/app/app.env" ]; then
    echo "loading constants from /app/app.env"
    . /app/app.env
fi

# Run migrations with our DB_SOURCE flag
echo "run db migration"
/app/migrate -path /app/migration -database "$DB_SOURCE" -verbose up

echo "start the app"
exec "$@"
