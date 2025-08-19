# Gig-Economy App Project Guide Template

This template provides a structured starting point for your project guide. It's designed as a living document (e.g., in Google Docs, Notion, or a Markdown file in your GitHub repo) that the junior developer can reference daily. Customize it with specifics from our discussions (e.g., AWS services, JSON Schemas, payment integrations). Include links to resources, diagrams, and code examples where possible. Update it as the project evolves, but keep changes minimal to avoid confusion.

---

## 1. Project Overview and Vision
### Goal
Build a gig-economy platform where consumers post jobs and gig workers accept them via mobile apps (iOS/Android with role-based UX). Use AWS for server-side (serverless architecture) and integrate flexible features like multi-provider payments (Stripe, Square, stablecoins).

### Key Features
- User roles: Consumers (post jobs), Gig Workers (accept/manage schedules), Admins.
- Core Models: People (Customers/Employees), Jobs, Transactions (with settlement/reconciliation), Schedules.
- Integrations: Payment providers via adapters; no data sync—use APIs.
- Tech Stack: AWS (Lambda, API Gateway, DynamoDB, EventBridge, Step Functions); Golang for web-service code; JSON Schemas for data validation.

### Success Metrics
- MVP: Functional CRUD for models, one payment provider integrated.
- Full Launch: Handles 1,000 daily transactions with 99.99% financial accuracy.

[Insert Diagram: High-level architecture flow (e.g., from Draw.io or Lucidchart).]

## 2. Development Principles
- **Start Small**: Build/test one component at a time (e.g., DynamoDB table before Lambda).
- **Best Practices**: Write tests (Jest), use version control (GitHub PRs), validate against schemas.
- **Risk Management**: If stuck, search AWS docs/Stack Overflow first; log questions for weekly check-in.
- **Learning Mindset**: Dedicate 20% time to tutorials; document what you learn.

## 3. Setup and Environment
### Prerequisites
- Install: Go, Docker, AWS CLI, SAM CLI (for local dev), VS Code (with extensions: AWS Toolkit, Prettier).
- AWS Account: Use free tier; create IAM user with limited permissions (e.g., Lambda/DynamoDB access only).
- GitHub Repo: This project is in GitHub.

### Initial Setup Steps
1. Configure AWS credentials: `aws configure` with access keys.
2. Install dependencies: `npm install` in repo root.
3. Local Testing: Use SAM to run Lambdas locally (e.g., `sam local start-api`).
4. Deploy Test: Follow AWS tutorial to deploy a sample Lambda.

[Code Example: Basic package.json with dependencies like aws-sdk, ajv for schema validation.]

## 4. Data Models and Schemas
Use the JSON Schema library we defined (copy/paste full schema here or link to file).

- **Key Models**: 
  - Person (abstract with subtypes Customer/Employee).
  - Job, Transaction (enhanced for payments), Schedule.
  - New: SettlementBatch, ReconciliationRecord.
- **Usage**: Validate all data in Lambdas (e.g., with Ajv library). Map external API responses to these schemas.

[Example Code: Go snippet for validating a Transaction object against schema.]

## 5. Architecture Breakdown
[Insert Detailed Diagram: Components like API Gateway -> Lambda Adapters -> DynamoDB.]

- **API Layer**: API Gateway for endpoints (e.g., /transactions).
- **Compute**: Lambda for logic (e.g., payment adapters).
- **Data**: DynamoDB tables (one per model, e.g., TransactionsTable).
- **Events/Workflows**: EventBridge for triggers; Step Functions for settlement/reconciliation.
- **Integrations**: Adapters for providers (Stripe/Square/Circle); webhooks for updates.

## 6. Milestones and Task List
Break into 2-week sprints. Each task includes: Description, Resources, Acceptance Criteria.

### Sprint 1: Setup and Core Models (Weeks 1-2)
- Task 1: Deploy basic AWS resources (DynamoDB tables from schemas).
  - Resources: AWS DynamoDB Getting Started (docs.aws.amazon.com/dynamodb).
  - Criteria: Tables created; simple insert/query works.
- Task 2: Implement Person schema validation in a test Lambda.
  - Resources: Ajv docs (ajv.js.org).
  - Criteria: Lambda returns error on invalid data.

### Sprint 2: Transaction Basics (Weeks 3-4)
- Task 1: Build Transaction Lambda 
  - Resources: Need to identify payment provider (Prefer Clover)
  - Criteria: Creates payment intent; stores in Postgres.

[Continue with sprints for reconciliation, mobile integrations, etc. Adjust based on progress.]

## 7. Resources and Learning Path
### AWS-Specific
- Tutorials: AWS Serverless Labs (aws.amazon.com/getting-started/hands-on/build-serverless-applications/).
- Docs: Lambda (docs.aws.amazon.com/lambda), DynamoDB (docs.aws.amazon.com/dynamodb).
- Videos: "AWS Lambda for Beginners" on YouTube (freeCodeCamp channel).

### General Dev
- Go: https://medium.com/@itskenzylimon/best-practices-when-developing-golang-backend-database-apis-part-3-2645738a76bc
- Testing: Postman and Newman
- Payments: TBD.

### Troubleshooting
- Common Issues: Check AWS Console for errors; use CloudWatch Logs.
- Ask Questions: Post in Slack channel; prepare for weekly call.

## 8. Workflow and Tools
- **Daily Routine**: Commit code daily; update task board.
- **Code Reviews**: Submit PRs; wait for approval.
- **Tools**: GitHub for code/issues; Postman for API testing; AWS Console for monitoring.
- **Weekly Check-In**: Prepare demo of progress; list blockers.

## 9. Risks and Contingencies
- If delayed: Prioritize core (e.g., skip stablecoins initially).
- Security: Always use Secrets Manager for keys; no hard-coded creds.
- Questions: If urgent, Slack; otherwise, batch for call.

This guide is your primary reference—read it fully before starting. Track your progress and celebrate small wins!

--- 
