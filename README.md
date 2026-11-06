# BiteShare API

Backend API for **BiteShare**, a platform that connects food donors with a food recovery workflow and provides an e-commerce catalog for products created from recovered food.

The backend is organized as two independently runnable Go services:

- **Donation API** — receives food donation submissions, stores donor and food data, manages review and pickup scheduling, and sends donor email notifications.
- **E-Commerce API** — exposes the processed-product catalog, handles checkout and payment-proof workflows, and provides admin product management.

Both services connect to the **same PostgreSQL database**. Food processing takes place offline and is outside the API workflow. After processing, an administrator manually lists the resulting products in the e-commerce service.

## Architecture at a glance

```text
Donor / Public User                  Admin
       |                               |
       v                               v
+---------------------+       +----------------------+
| Donation API :8081  |       | E-Commerce API :8080 |
| - Submit donation   |       | - Admin login        |
| - Review donation   |       | - Product catalog    |
| - Schedule pickup   |       | - Checkout           |
| - Email updates     |       | - Payment proof      |
+----------+----------+       +-----------+----------+
           |                              |
           +--------------+---------------+
                          v
                +-------------------+
                | Shared PostgreSQL |
                | (Supabase/local)  |
                +-------------------+

External services: ImageKit (images) and Resend (email)
```

## Tech stack

| Area | Technology |
|---|---|
| Language | Go 1.26.10 (see `go.mod`) |
| HTTP framework | Gin |
| ORM | GORM |
| Database | PostgreSQL / Supabase PostgreSQL |
| Authentication | JWT bearer token, shared admin identity |
| Image storage | ImageKit |
| Email | Resend |
| Configuration | Environment variables and `.env` (`godotenv`) |
| API specification | OpenAPI 3.0 in `docs/openapi.yaml` |

## Repository structure

```text
.
├── cmd/
│   ├── donation/       # Donation API entry point
│   ├── ecommerce/      # E-Commerce API entry point
│   ├── migrate/        # Database migration command
│   ├── seed-admin/     # Initial admin seeding command
│   └── seed-products/  # Sample product seeding command
├── donation-api/
│   ├── dto/            # Donation request/response DTOs
│   ├── handler/        # Donation HTTP handlers
│   ├── model/          # Donation domain models
│   ├── repository/     # Donation database operations
│   └── service/        # Donation business rules and integrations
├── ecommerce-api/
│   ├── dto/
│   ├── handler/
│   ├── model/
│   ├── repository/
│   └── service/
├── internal/
│   ├── auth/           # JWT and password utilities
│   ├── database/       # Shared PostgreSQL connection
│   ├── imagekit/       # ImageKit client
│   ├── mailer/         # Resend email client
│   ├── payment/        # Payment workflow
│   └── server/         # Server and route registration
├── schema/
│   ├── migrations/     # SQL up/down migrations
│   └── schema.sql      # Database schema snapshot
├── docs/
│   ├── openapi.yaml    # Swagger / OpenAPI specification
│   └── TECHNICAL.md    # Technical documentation
├── .env.example
├── go.mod
└── Makefile
```

## Prerequisites

- Go version compatible with `go.mod` (Go 1.26.10 in this repository).
- PostgreSQL database. A Supabase PostgreSQL project can be used.
- ImageKit account for donation, product, and payment-proof images.
- Resend account for transactional email.

## Local setup

1. Clone the repository and enter the project directory.

   ```bash
   git clone <repository-url>
   cd biteshare-api
   ```

2. Copy the example environment file.

   ```powershell
   # PowerShell
   Copy-Item .env.example .env
   ```

   ```bash
   # Linux / macOS
   cp .env.example .env
   ```

3. Fill in the required values in `.env`. Do not commit `.env` or share it publicly.

4. Download Go dependencies.

   ```bash
   go mod download
   ```

5. Prepare the PostgreSQL database. Configure the `BLUEPRINT_DB_*` variables to point to your database. Apply the SQL migrations in `schema/migrations` in ascending order using the repository migration command:

   ```bash
   go run ./cmd/migrate -action up
   ```

   The command records applied migration filenames in `schema_migrations` and skips migrations that have already been applied. Review the SQL files and take a backup before running migrations against production.

6. Seed the initial administrator account(s), if this is a fresh environment:

   ```bash
   go run ./cmd/seed-admin
   ```

   Configure `ADMIN_1_*` (and optionally `ADMIN_2_*`, `ADMIN_3_*`) in `.env`. Run the seeder only when needed; do not repeatedly seed production data without checking its behavior.

## Environment variables

Use `.env.example` as the canonical template. Main settings:

| Variable | Purpose |
|---|---|
| `PORT` | E-Commerce API port; defaults to `8080` |
| `DONATION_PORT` | Donation API port; defaults to `8081` |
| `DONATION_PORT` | Donation API port; defaults to `8081` |
| `APP_ENV` | Application environment label |
| `JWT_SECRET` | Secret used to sign and validate admin access tokens; use a random secret of at least 32 characters |
| `JWT_ISSUER` | JWT issuer; defaults to `biteshare-api` |
| `JWT_ACCESS_TOKEN_TTL` | Access-token lifetime; defaults to `24h` |
| `BLUEPRINT_DB_HOST` | PostgreSQL host |
| `BLUEPRINT_DB_PORT` | PostgreSQL port |
| `BLUEPRINT_DB_DATABASE` | Database name |
| `BLUEPRINT_DB_USERNAME` | Database username |
| `BLUEPRINT_DB_PASSWORD` | Database password |
| `BLUEPRINT_DB_SCHEMA` | Database schema, commonly `public` |
| `BLUEPRINT_DB_SSLMODE` | PostgreSQL SSL mode; use the value required by your provider |
| `IMAGEKIT_PRIVATE` | ImageKit private API key (backend only) |
| `IMAGEKIT_PUBLIC` | ImageKit public key (used where applicable by the client) |
| `IMAGEKIT_ID` | ImageKit account ID / URL identifier |
| `IMAGEKIT_ENDPOINT` | ImageKit delivery endpoint |
| `RESEND_API_KEY` | Resend API key |
| `RESEND_FROM_EMAIL` | Verified sender address configured in Resend |
| `ADMIN_1_NAME`, `ADMIN_1_EMAIL`, `ADMIN_1_PASSWORD` | First administrator seed values |
| `ADMIN_2_*`, `ADMIN_3_*` | Optional additional administrator seed values |
| `SEED_ADMIN_EMAIL` | Admin email used by the seed command where applicable |

**Important:** the Donation API and E-Commerce API must use the same database and JWT configuration so that the admin token issued by the E-Commerce API can be validated by the Donation API.

## Run the services

Run each service in a separate terminal from the repository root.

```bash
# Terminal 1 — E-Commerce API (default :8080)
go run ./cmd/ecommerce
```

```bash
# Terminal 2 — Donation API (default :8081)
go run ./cmd/donation
```

Health checks:

```text
GET http://localhost:8080/
GET http://localhost:8080/health
GET http://localhost:8081/
GET http://localhost:8081/health
```

The root endpoint returns a simple service message. The health endpoint reports database connection-pool health.

## API documentation

The OpenAPI specification is available at [`docs/openapi.yaml`](docs/openapi.yaml), and an interactive Swagger UI is served by the E-Commerce API at `http://localhost:8080/swagger/index.html`. The specification is maintained as an OpenAPI YAML file (not generated from Go annotations). You can also open it with [Swagger Editor](https://editor.swagger.io/) to inspect or edit the contract.

Base URLs:

- E-Commerce API: `http://localhost:8080`
- Donation API: `http://localhost:8081`

## Main workflows

### Donation

1. A donor submits their contact details, one or more food items, and at least one image for every food item.
2. The Donation API validates the multipart request and uploads images to ImageKit.
3. Donor, donation request, food items, and image metadata are stored in PostgreSQL.
4. The donor receives a submission email when Resend is configured.
5. An authenticated admin reviews the donation and approves or rejects it.
6. For an approved request, an admin schedules pickup. The donor receives a schedule email.
7. Pickup and receipt by BiteShare are operational milestones. Food processing is performed offline and is not an API operation in this repository.

### E-Commerce

1. An administrator manually enters processed food products into the product catalog.
2. Public users browse available products and submit checkout requests.
3. Buyers submit a payment proof when using manual transfer.
4. Admins review payment proofs. A payment webhook endpoint can update payment and order status.

## Tests and build

```bash
# Run all Go tests
go test ./...

# Build both services (Makefile targets)
make build

# Run tests through Makefile
make test
```

Integration tests may require Docker because the project includes Testcontainers-based PostgreSQL tests. Review the test setup before running them in a restricted environment.

## Security notes

- Never commit `.env`, JWT secrets, database passwords, ImageKit private keys, or Resend API keys.
- Admin-only endpoints require `Authorization: Bearer <access_token>`.
- Obtain the admin access token from `POST /auth/login` on the E-Commerce API.
- Configure a verified sender domain in Resend before sending email to arbitrary donor addresses. Resend's testing sender may restrict recipients.
- The payment webhook route is currently protected by the admin JWT middleware in the route registration. Confirm the intended payment-provider authentication/signature mechanism before exposing it to a third-party provider.

## Project status and scope

The Donation API and E-Commerce API are separate HTTP services but share one PostgreSQL database. The Donation API workflow ends at donation intake, review, pickup scheduling, and the operational handoff of food to BiteShare. Food processing occurs offline. E-Commerce begins when an admin lists the processed output as a product.
