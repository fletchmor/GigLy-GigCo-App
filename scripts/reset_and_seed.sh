#!/bin/bash

# Reset and seed the GigCo database for testing
# This script will drop all data and recreate it with seed data

echo "🔄 Resetting GigCo database..."

# Stop containers if running
echo "Stopping containers..."
docker compose down

# Start containers
echo "Starting containers..."
docker compose up -d

# Wait for database to be ready
echo "Waiting for database to be ready..."
sleep 10

# Run the main initialization script
echo "Running database initialization..."
cat /Users/fletch/app/scripts/init.sql | docker compose exec -T postgres psql -U postgres -d gigco

# Run additional seed data
echo "Adding additional seed data..."
cat /Users/fletch/app/scripts/simple_seed.sql | docker compose exec -T postgres psql -U postgres -d gigco

# Test the API
echo "Testing API endpoints..."
echo "Health check:"
curl -s http://localhost:8080/health | jq '.'

echo -e "\nGigWorkers count:"
curl -s http://localhost:8080/api/v1/gigworkers | jq '.pagination.total'

echo -e "\nJobs count:"
curl -s http://localhost:8080/api/v1/jobs | jq '.pagination.total'

echo -e "\nCustomers count:"
curl -s http://localhost:8080/api/v1/customers/1 | jq '.name'

echo -e "\n✅ Database reset and seeded successfully!"
echo "You can now run your Postman tests."