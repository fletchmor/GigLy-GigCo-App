# Clover Payment Integration Guide

## Overview

This guide covers the complete Clover payment integration for GigCo, including escrow payments, payment splitting, and refunds.

## What Was Built

### 1. Database Schema (`scripts/clover_payment_schema.sql`)

**New Tables:**
- `transactions` (enhanced) - Tracks all payment transactions with Clover integration
- `payment_splits` - Tracks fee distribution (platform fees, worker payments, tips, etc.)
- `payment_events` - Audit log of all payment events
- `user_payment_methods` (enhanced) - Stores tokenized payment methods

**Key Features:**
- Pre-authorization (escrow) and capture workflow
- Payment splitting for platform fees and worker payments
- Refund tracking with parent transaction references
- Complete audit trail via payment_events table
- Helper functions for fee calculations and payment summaries

**To Apply:**
```bash
# 1. Connect to your database
PGPASSWORD=bamboo psql -h localhost -p 5433 -U postgres -d gigco

# 2. Run the schema update
\i scripts/clover_payment_schema.sql
```

### 2. Go Backend Implementation

#### Models (`internal/model/payment.go`)
- `EnhancedTransaction` - Complete transaction model with Clover fields
- `PaymentSplit` - Fee distribution tracking
- `PaymentEvent` - Audit log entries
- `CloverConfig`, `CloverChargeRequest`, `CloverChargeResponse`, etc. - Clover API models
- Request/Response models for all payment endpoints

#### Configuration (`config/payment.go`)
- `PaymentConfig` - Main payment configuration
- `CloverConfig` - Clover-specific settings
- Fee calculation helpers (platform fee, processing fee)
- Environment-based endpoint configuration (sandbox/production)

#### Services

**Clover Service (`internal/payment/clover_service.go`)**
- Direct integration with Clover REST API
- Card tokenization
- Payment authorization (pre-auth)
- Payment capture
- Refund processing

**Payment Service (`internal/payment/payment_service.go`)**
- Business logic layer
- Database operations for transactions
- Authorization with escrow
- Capture with fee splitting
- Refund processing with job status updates

#### API Handlers (`api/payment_handlers.go`)
- `POST /api/v1/payments/authorize` - Pre-authorize payment (hold funds in escrow)
- `POST /api/v1/payments/capture` - Capture payment (release from escrow)
- `POST /api/v1/payments/refund` - Refund a payment
- `GET /api/v1/jobs/{id}/payments` - Get all transactions for a job
- `GET /api/v1/jobs/{id}/payment-summary` - Get payment summary for a job

### 3. Environment Configuration

Add these environment variables to your `.env` file or `docker-compose.yml`:

```bash
# Clover Configuration
PAYMENT_PROVIDER=clover
CLOVER_ENVIRONMENT=sandbox  # or 'production'
CLOVER_MERCHANT_ID=your_merchant_id
CLOVER_ACCESS_TOKEN=your_access_token
CLOVER_API_ACCESS_KEY=your_pakms_key
CLOVER_WEBHOOK_SECRET=your_webhook_secret

# Platform Configuration
PLATFORM_FEE_PERCENT=10.0  # 10% platform fee
```

## Payment Workflow

### 1. Job Creation & Payment Authorization (Escrow)

When a consumer posts a job and a worker accepts:

```bash
POST /api/v1/payments/authorize
Content-Type: application/json
X-User-ID: <consumer_id>

{
  "job_id": 123,
  "amount": 100.00,
  "card_details": {
    "number": "4242424242424242",
    "exp_month": "12",
    "exp_year": "2025",
    "cvv": "123",
    "name": "John Doe"
  },
  "save_card": true
}
```

**What Happens:**
1. Card is tokenized with Clover
2. Payment is pre-authorized (funds held but not captured)
3. Transaction record created with `status=completed`, `transaction_type=authorization`
4. Funds held in escrow (`escrow_held_at` timestamp set)
5. Platform and processing fees calculated
6. Payment event logged for audit trail

### 2. Job Completion & Payment Capture (Release from Escrow)

When both parties confirm job completion:

```bash
POST /api/v1/payments/capture
Content-Type: application/json
X-User-ID: <consumer_or_worker_id>

{
  "transaction_id": 456
}
```

**What Happens:**
1. Clover captures the pre-authorized payment
2. Transaction updated with `captured_at`, `escrow_released_at` timestamps
3. Payment splits created:
   - Platform fee (10% default)
   - Worker payment (net amount after fees)
4. Job status updated to `paid`
5. Capture event logged

### 3. Job Cancellation & Refund

If job is cancelled:

```bash
POST /api/v1/payments/refund
Content-Type: application/json
X-User-ID: <consumer_id>

{
  "transaction_id": 456,
  "amount": 100.00,  # Optional: omit for full refund
  "reason": "Job cancelled by consumer"
}
```

**What Happens:**
1. Clover processes the refund
2. New refund transaction created with `transaction_type=refund`
3. Original transaction marked as `status=refunded`
4. Job status updated to `cancelled`
5. Refund event logged

### 4. Payment Summary

Get complete payment details for a job:

```bash
GET /api/v1/jobs/123/payment-summary
X-User-ID: <user_id>
```

**Response:**
```json
{
  "job_id": 123,
  "total_authorized": 100.00,
  "total_captured": 100.00,
  "total_refunded": 0.00,
  "platform_fees": 10.00,
  "worker_payment": 87.40,
  "escrow_status": "released"
}
```

## Fee Structure

### Default Configuration

**Platform Fee:** 10% of total amount
- Configurable via `PLATFORM_FEE_PERCENT` environment variable
- Calculated as: `amount * (platform_fee_percent / 100)`

**Processing Fee:** ~2.6% + $0.10 (Clover's typical rate)
- Calculated as: `(amount * 0.026) + 0.10`

**Net Amount to Worker:**
- `net_amount = total_amount - platform_fee - processing_fee`
- Example: $100 job = $10 platform fee + $2.70 processing fee = $87.30 to worker

### Customizing Fees

Edit `config/payment.go`:

```go
func (c *CloverConfig) CalculatePlatformFee(amount float64) float64 {
    return amount * (c.PlatformFeePercent / 100.0)
}

func (c *CloverConfig) CalculateProcessingFee(amount float64) float64 {
    percentage := 2.6  // Customize this
    fixedFee := 0.10    // Customize this
    return (amount * (percentage / 100.0)) + fixedFee
}
```

## Temporal Workflow Integration

### Recommended Integration Points

1. **Job Acceptance Workflow** - Trigger payment authorization
2. **Job Completion Workflow** - Trigger payment capture
3. **Job Cancellation Workflow** - Trigger refund

### Example Temporal Activity

Create `internal/temporal/activities/payment_activities.go`:

```go
package activities

import (
    "context"
    "app/config"
    "app/internal/model"
    "app/internal/payment"
)

type PaymentActivities struct {
    paymentService *payment.PaymentService
}

func NewPaymentActivities() *PaymentActivities {
    return &PaymentActivities{
        paymentService: payment.NewPaymentService(config.DB, &config.Payment.Clover),
    }
}

func (a *PaymentActivities) AuthorizeJobPayment(ctx context.Context, req model.PaymentAuthorizeRequest) error {
    _, err := a.paymentService.AuthorizeJobPayment(req.ConsumerID, req)
    return err
}

func (a *PaymentActivities) CaptureJobPayment(ctx context.Context, transactionID, userID int) error {
    req := model.PaymentCaptureRequest{TransactionID: transactionID}
    _, err := a.paymentService.CaptureJobPayment(userID, req)
    return err
}

func (a *PaymentActivities) RefundJobPayment(ctx context.Context, transactionID, userID int, reason string) error {
    req := model.PaymentRefundRequest{
        TransactionID: transactionID,
        Reason: reason,
    }
    _, err := a.paymentService.RefundJobPayment(userID, req)
    return err
}
```

## iOS App Integration

### APIService Updates

Add payment endpoints to `Services/APIService.swift`:

```swift
func authorizePayment(jobId: Int, cardDetails: CardDetails) async throws -> PaymentAuthResponse {
    let url = baseURL.appendingPathComponent("/payments/authorize")

    let requestBody: [String: Any] = [
        "job_id": jobId,
        "amount": calculateJobAmount(jobId),
        "card_details": [
            "number": cardDetails.number,
            "exp_month": cardDetails.expMonth,
            "exp_year": cardDetails.expYear,
            "cvv": cardDetails.cvv,
            "name": cardDetails.name
        ],
        "save_card": true
    ]

    // Make request...
}

func capturePayment(transactionId: Int) async throws -> PaymentCaptureResponse {
    let url = baseURL.appendingPathComponent("/payments/capture")
    let requestBody = ["transaction_id": transactionId]
    // Make request...
}

func refundPayment(transactionId: Int, reason: String) async throws -> PaymentRefundResponse {
    let url = baseURL.appendingPathComponent("/payments/refund")
    let requestBody = [
        "transaction_id": transactionId,
        "reason": reason
    ]
    // Make request...
}
```

### Payment Flow in iOS

1. **Job Creation** - Add payment method entry
2. **Job Acceptance** - Trigger payment authorization
3. **Job Completion** - Show "Release Payment" button
4. **Payment Capture** - Call capture endpoint when both parties confirm

### UI Components Needed

- `PaymentMethodView` - Add/manage credit cards
- `JobPaymentView` - Display payment status for a job
- `PaymentConfirmationView` - Confirm payment release

## Testing

### Test with Clover Sandbox

**Test Card Numbers:**
- Success: `4242424242424242`
- Decline: `4000000000000002`
- Insufficient Funds: `4000000000009995`

### Test Workflow

1. **Setup Environment:**
```bash
docker compose up --build
```

2. **Apply Schema:**
```bash
PGPASSWORD=bamboo psql -h localhost -p 5433 -U postgres -d gigco < scripts/clover_payment_schema.sql
```

3. **Test Authorization:**
```bash
curl -X POST http://localhost:8080/api/v1/payments/authorize \
  -H "Content-Type: application/json" \
  -H "X-User-ID: 1" \
  -d '{
    "job_id": 1,
    "amount": 100.00,
    "card_details": {
      "number": "4242424242424242",
      "exp_month": "12",
      "exp_year": "2025",
      "cvv": "123"
    }
  }'
```

4. **Verify in Database:**
```sql
SELECT * FROM transactions WHERE job_id = 1;
SELECT * FROM payment_splits WHERE transaction_id = <id>;
SELECT * FROM payment_events WHERE transaction_id = <id>;
```

5. **Test Capture:**
```bash
curl -X POST http://localhost:8080/api/v1/payments/capture \
  -H "Content-Type: application/json" \
  -H "X-User-ID: 1" \
  -d '{"transaction_id": <id>}'
```

6. **Test Refund:**
```bash
curl -X POST http://localhost:8080/api/v1/payments/refund \
  -H "Content-Type: application/json" \
  -H "X-User-ID: 1" \
  -d '{
    "transaction_id": <id>,
    "reason": "Test refund"
  }'
```

## Security Considerations

### PCI Compliance

✅ **Implemented:**
- Card data never stored in your database
- Tokenization handled by Clover
- Secure API communication with Clover

⚠️ **Additional Steps Needed:**
- Implement proper JWT authentication (currently using placeholder)
- Add rate limiting on payment endpoints
- Set up webhook signature verification
- Enable HTTPS in production
- Regular security audits

### Authentication

**Current Implementation:** Placeholder using `X-User-ID` header

**TODO:** Replace with proper JWT authentication:
```go
func getUserIDFromContext(r *http.Request) int {
    // TODO: Extract user ID from JWT token
    claims := r.Context().Value("jwt_claims")
    // Validate and return user ID
}
```

## Production Checklist

Before going live:

- [ ] Obtain production Clover credentials
- [ ] Set `CLOVER_ENVIRONMENT=production`
- [ ] Update Clover endpoints to production URLs
- [ ] Implement proper JWT authentication
- [ ] Add webhook handlers for Clover events
- [ ] Set up monitoring and alerts for failed payments
- [ ] Test refund and dispute workflows
- [ ] Verify PCI compliance requirements
- [ ] Add payment retry logic for failed captures
- [ ] Implement payment reconciliation process
- [ ] Set up customer support for payment issues

## Support & Resources

- **Clover Documentation:** https://docs.clover.com/
- **Clover Sandbox:** https://sandbox.dev.clover.com/
- **API Reference:** https://docs.clover.com/dev/reference/api-reference-overview
- **Ecommerce API:** https://docs.clover.com/dev/docs/ecommerce-api-tutorials

## Troubleshooting

### Payment Authorization Fails

1. Check Clover credentials in environment variables
2. Verify PAKMS API key is correct
3. Check card details format (exp_month should be "12", not 12)
4. Review payment_events table for error details

### Capture Fails

1. Verify authorization hasn't expired (typically 7 days)
2. Check that transaction type is "authorization"
3. Ensure transaction hasn't been captured already
4. Review Clover dashboard for auth status

### Refund Fails

1. Verify original transaction was captured
2. Check refund amount doesn't exceed original amount
3. Ensure transaction hasn't been fully refunded already
4. Review Clover merchant account refund policies

## Next Steps

1. ✅ Backend implementation complete
2. ⏳ Apply database schema
3. ⏳ Configure Clover credentials
4. ⏳ Integrate with Temporal workflows
5. ⏳ Build iOS payment UI
6. ⏳ Test end-to-end payment flow
7. ⏳ Production deployment

---

**Need Help?** Check the code comments in:
- `internal/payment/payment_service.go` - Business logic
- `internal/payment/clover_service.go` - Clover API integration
- `api/payment_handlers.go` - API endpoints
