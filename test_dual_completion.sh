#!/bin/bash
set -e

echo "🧪 Testing Dual Completion Feature"
echo "=================================="
echo ""

# Login as consumer
echo "1️⃣  Logging in as consumer..."
CONSUMER_RESPONSE=$(curl -s -X POST 'http://localhost:8080/api/v1/auth/login' \
  -H 'Content-Type: application/json' \
  -d '{"email":"testconsumer@gigco.dev","password":"test123"}')
CONSUMER_TOKEN=$(echo $CONSUMER_RESPONSE | python3 -c "import sys, json; print(json.load(sys.stdin)['token'])")
CONSUMER_ID=$(echo $CONSUMER_RESPONSE | python3 -c "import sys, json; print(json.load(sys.stdin)['id'])")
echo "   ✅ Consumer logged in (ID: $CONSUMER_ID)"

# Login as worker
echo "2️⃣  Logging in as worker..."
WORKER_RESPONSE=$(curl -s -X POST 'http://localhost:8080/api/v1/auth/login' \
  -H 'Content-Type: application/json' \
  -d '{"email":"worker1@gigco.dev","password":"test123"}')
WORKER_TOKEN=$(echo $WORKER_RESPONSE | python3 -c "import sys, json; print(json.load(sys.stdin)['token'])")
WORKER_ID=$(echo $WORKER_RESPONSE | python3 -c "import sys, json; print(json.load(sys.stdin)['id'])")
echo "   ✅ Worker logged in (ID: $WORKER_ID)"

# Create a job
echo "3️⃣  Creating a job..."
JOB_RESPONSE=$(curl -s -X POST 'http://localhost:8080/api/v1/jobs/create' \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer $CONSUMER_TOKEN" \
  -d "{\"title\":\"Test Job for Dual Completion\",\"description\":\"Testing the new dual completion feature\",\"consumer_id\":$CONSUMER_ID,\"category\":\"testing\",\"total_pay\":50.00}")
JOB_ID=$(echo $JOB_RESPONSE | python3 -c "import sys, json; print(json.load(sys.stdin)['id'])")
echo "   ✅ Job created (ID: $JOB_ID)"

# Worker accepts the job
echo "4️⃣  Worker accepting job..."
curl -s -X POST "http://localhost:8080/api/v1/jobs/$JOB_ID/accept" \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer $WORKER_TOKEN" > /dev/null
echo "   ✅ Job accepted by worker"

# Worker starts the job
echo "5️⃣  Worker starting job..."
curl -s -X POST "http://localhost:8080/api/v1/jobs/$JOB_ID/start" \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer $WORKER_TOKEN" > /dev/null
echo "   ✅ Job started"

# Worker completes the job (first confirmation)
echo "6️⃣  Worker marking job as complete..."
WORKER_COMPLETE=$(curl -s -X POST "http://localhost:8080/api/v1/jobs/$JOB_ID/complete" \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer $WORKER_TOKEN")
echo "$WORKER_COMPLETE" | python3 -m json.tool
echo ""

# Consumer completes the job (second confirmation)
echo "7️⃣  Consumer confirming job completion..."
CONSUMER_COMPLETE=$(curl -s -X POST "http://localhost:8080/api/v1/jobs/$JOB_ID/complete" \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer $CONSUMER_TOKEN")
echo "$CONSUMER_COMPLETE" | python3 -m json.tool
echo ""

# Check final job status
echo "8️⃣  Checking final job status..."
JOB_STATUS=$(curl -s -X GET "http://localhost:8080/api/v1/jobs/$JOB_ID" \
  -H "Authorization: Bearer $CONSUMER_TOKEN")
echo "$JOB_STATUS" | python3 -m json.tool | grep -A1 -E '"status"|"worker_completed_at"|"consumer_completed_at"|"fully_completed"'

echo ""
echo "✅ Dual Completion Test Complete!"
