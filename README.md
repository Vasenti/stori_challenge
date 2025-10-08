# Stori challenge

Ingest a CSV of account transactions, persist them in Postgres (via GORM), compute a monthly summary (total balance, transactions count per month, average debit/credit), and email an HTML report with a customizable template. We assume this project is compatible with Windows and Linux.

- Input: CSV with headers `Id,Date,Transaction`
- Storage: PostgreSQL
- ORM: GORM
- Email: SMTP (MailHog in dev, SES/SMTP in prod)
- File source: local path or `s3://bucket/key`
- Architecture: **hexagonal (ports & adapters)**
- Config: env vars (via `caarlos0/env`)
- Deployment: CLI, Docker Compose, AWS Lambda (image), SAM for local Lambda testing

---

## Contents

- [Quick Start](#quick-start)
- [Architecture](#architecture)
- [Data Model & Parsing Rules](#data-model--parsing-rules)
- [Technical Decisions](#technical-decisions--rationale)
- [Configuration](#configuration)
- [Run with Docker Compose](#run-with-docker-compose)
- [Run the CLI](#run-the-cli)
- [Email Templates](#email-templates)
- [AWS Lambda (Local & Cloud)](#aws-lambda-local--cloud)
- [Troubleshooting](#troubleshooting)

---

## Quick Start

```bash
# 0) Prereqs: Docker, Docker Compose, Go >= 1.25 (for local build), AWS SAM CLI (optional)

# 1) Put a sample CSV
mkdir -p data
cp ./transactions.csv ./data/transactions.csv  # or create your own CSV

# 2) Bring up infra (Postgres + DB init + MailHog)
docker compose up -d postgres pg-init mailhog

# 3) Run the job (containerized CLI)
docker compose run --rm transaction_manager   --email=you@example.com   --src=/data/transactions.csv   --template=/templates/report.html.tmpl   # optional

# 4) See the email
# MailHog UI: http://localhost:8025
```

---

## Architecture

**Hexagonal (ports & adapters)**:

```
cmd/
  transaction_manager/           # CLI entrypoint
  lambda/             # AWS Lambda handler
internal/
  application/
    ports/            # interfaces (Reader, Repos, EmailSender, Service)
    services/         # TransactionReportService (use-case orchestration)
  domain/             # Entities (User, Transaction) + MonthlySummary
  intrastructure/     # (typo kept as folder name) adapters
    db/               # GORM setup
      repositories/   # UserRepository, TransactionRepository
      reader/         # LocalFileReader, S3Reader
    email/            # SMTPSender (SMTP/STARTTLS)
    parser/           # CSV parsing
    templating/       # HTML templating (default embedded + custom)
```

Flow:

1. **Reader** opens CSV (local/S3)
2. **Parser** converts rows → `[]Transaction`
3. **UserRepository** ensures user exists (by email)
4. **TransactionRepository** bulk-upsert with conflict strategy
5. **TransactionRepository** computes monthly summary
6. **Templating** renders HTML from summary (+ user + time)
7. **EmailSender** sends via SMTP

---

## Data Model & Parsing Rules

**CSV Header**: `Id,Date,Transaction`  
- **Id**: `uint` (can be `0`, `1`, `100000`, etc.)  
- **Date**:
  - Formats:
    - `"M/D"` → assumed **current year**
    - `"YYYY/M/D"` → use provided year
  - Parsed to `OccurredAt time.Time`
- **Transaction**: `float64`
  - Positive = **credit**
  - Negative = **debit**

**Domain**
```go
type Transaction struct {
  ID         uint      `gorm:"primaryKey;autoIncrement:false"`
  UserEmail  string    `gorm:"primaryKey;not null"` // acts as FK to users(email)
  OccurredAt time.Time `gorm:"index;not null"`
  Amount     float64   `gorm:"not null"`
  RawDate    string    `gorm:"not null"`
  RawAmount  string    `gorm:"not null"`
}
```

**Summary**:
- `BalanceTotal`: sum of all amounts
- `TransactionsByMonth`: `map[time.Month]int`
- `AvgDebit`: average **absolute** value of negatives
- `AvgCredit`: average of positives

---

## Technical Decisions

- **Composite key (`user_email`, `id`)** for `transactions`  
  The CSV’s `id` is **not globally unique**; it may repeat per user. Using a composite PK makes imports **idempotent** and prevents cross-user collisions (e.g., `(alice,0)` and `(bob,0)` both valid). We also set `autoIncrement:false` so `ID=0` is preserved.

- **Upsert with conflict on (`user_email`,`id`)**  
  `ON CONFLICT (user_email, id) DO NOTHING` → safe reprocessing, no duplicates.  
  Requires PK/UNIQUE on those columns.

- **Monthly summary computed in Go** (readability over SQL)  
  For this challenge size it’s simpler to maintain, easy to test, and avoids DB-specific syntax.

- **SMTP adapter**  
  Dev: MailHog (no auth)  
  Prod: SES (SMTP) with STARTTLS/Auth or SES SDK adapter if preferred.  
  Email HTML uses inline styles & tables for client compatibility.

- **Config via env** (`caarlos0/env`)  
  All credentials/addresses come from environment variables → containers/Lambda friendly

---

## Configuration

Environment variables (main ones):

```
APP_ENV=dev

# Postgres
DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=stori
DB_SSLMODE=disable
DB_MAX_OPEN=10
DB_MAX_IDLE=5
DB_MAX_LIFETIME_SECS=600

# SMTP (MailHog for dev)
SMTP_HOST=mailhog
SMTP_PORT=1025
SMTP_FROM=no-reply@example.com
SMTP_USERNAME=
SMTP_PASSWORD=

# S3 / MinIO (optional, only for s3://src or s3://template)
S3_REGION=us-east-1
S3_ENDPOINT=
S3_ACCESS_KEY=
S3_SECRET_KEY=
S3_FORCE_PATH_STYLE=false
```

> Do **not** commit real secrets. Use `.env` locally; in Lambda/Cloud use function environment/config.

---

## Run with Docker Compose

The compose file brings up:
- `postgres` (db)
- `pg-init` (idempotently creates DB `stori` if missing)
- `mailhog` (smtp + web ui)
- `transaction_manager` (your binary; run it on demand)

**Compose snippet (key parts):**
```yaml
services:
  postgres:
    image: postgres:16
    container_name: stori_pg
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
    ports: ["5432:5432"]
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 5s
      timeout: 5s
      retries: 10
    volumes:
      - pgdata:/var/lib/postgresql/data

  pg-init:
    image: postgres:16
    depends_on:
      postgres:
        condition: service_healthy
    environment:
      PGPASSWORD: postgres
    command: ["sh","-c",
      "until pg_isready -h postgres -U postgres >/dev/null 2>&1; do sleep 1; done;        exists=$(psql 'postgresql://postgres:postgres@postgres:5432/postgres?sslmode=disable' -tc "SELECT 1 FROM pg_database WHERE datname='stori'" | tr -d '[:space:]');        if [ "$exists" != "1" ]; then          psql 'postgresql://postgres:postgres@postgres:5432/postgres?sslmode=disable' -c 'CREATE DATABASE stori';        fi"
    ]
    restart: "no"

  mailhog:
    image: mailhog/mailhog:v1.0.1
    container_name: stori_mailhog
    ports: ["1025:1025", "8025:8025"]
    healthcheck:
      test: ["CMD", "nc", "-z", "localhost", "1025"]
      interval: 5s
      timeout: 3s
      retries: 10

  transaction_manager:
    build:
      context: .
      dockerfile: Dockerfile
    container_name: stori_transaction_manager
    depends_on:
      postgres: { condition: service_healthy }
      pg-init:  { condition: service_completed_successfully }
      mailhog:  { condition: service_healthy }
    profiles: ["manual"]
    environment: # (see Configuration section)
      DB_HOST: postgres
      ...
    volumes:
      - ./data:/data:ro
      - ./templates:/templates:ro
    restart: "no"

volumes:
  pgdata:
```

**Run:**
```bash
docker compose up -d postgres pg-init mailhog

# With a template file:
docker compose run --rm transaction_manager   --email=you@example.com   --src=/data/transactions.csv   --template=/templates/report.html.tmpl

# Or default embedded template:
docker compose run --rm transaction_manager   --email=you@example.com   --src=/data/transactions.csv
```

Open MailHog: http://localhost:8025

> **Windows tip:** ensure `./data/transactions.csv` exists *before* running; otherwise Docker may create a **directory** at `/data/transactions.csv` inside the container.

---

## Run the CLI

You can also run the binary directly:

```bash
# Go toolchain >= 1.25
go run ./cmd/importer   --email=you@example.com   --src=./data/transactions.csv   --template=./templates/report.html.tmpl
```

Flags:
- `--email` (required): recipient AND user key
- `--src` (required): CSV path; local or `s3://bucket/key`
- `--template` (optional): path to HTML template; if empty, uses the embedded default

---

## Email Templates

- **Default template** is embedded (via `go:embed`) and used when `--template` is empty.
- To provide a custom one, pass a **file path** (CLI) or **HTML string** (service), or (Lambda) a path that the handler reads into a string before rendering.

**A polished English template is already included** (inline-styled, email-client friendly) with the company SVG logo. If you see the raw path as email body, it means you passed the path as **content**—ensure the handler reads the file and passes HTML to the renderer (already handled in `cmd/lambda`).

---

## AWS Lambda (Local & Cloud)

### Local (SAM, image-based)

`Dockerfile.lambda` (multi-stage) builds `cmd/lambda` and copies sample data/templates into `/var/task`:

```bash
sam build --clean
# Attach to the same Docker network as compose (replace with your network name)
sam local invoke ImporterFn   --event event.json   --docker-network <your_compose_network>   # e.g. stori-challenge_default
```

`event.json`:
```json
{
  "email": "user@example.com",
  "src": "/var/task/data/transactions.csv",
  "template": "/var/task/templates/report.html.tmpl"
}
```

**Notes**
- Ensure Go **1.25** in the builder (`golang:1.25-alpine`) since `go.mod` requires it.
- In the template handler we **load the template file content** (or S3) and pass HTML to the renderer.

### Cloud (ECR image)

- Build `Dockerfile.lambda`, push to ECR.
- Create Lambda (Package type: **Image**).
- Set ENV vars (DB, SMTP, S3…).
- Networking:
  - If DB is in a VPC, attach the function to that VPC/subnets and allow outbound to the DB SG.
  - For SES/SMTP or S3, ensure the function has Internet/NAT if needed.
- Test with a payload like `event.json` above (for S3, pass `"src":"s3://bucket/key"` and proper IAM permissions).

---

## Troubleshooting

- **“database ‘stori’ does not exist”**  
  Postgres is up, but DB wasn’t created. Use `pg-init` (compose), or create once:
  ```bash
  docker exec -it stori_pg psql -U postgres -c "CREATE DATABASE stori;"
  ```

- **“is a directory” when reading CSV**  
  Your host path was missing; Docker created a directory. Ensure `./data/transactions.csv` exists and mount `./data:/data:ro`. Use `--src=/data/transactions.csv`.

- **Email body shows a file path**  
  You passed the template *path* as content. The Lambda handler now reads the file and passes HTML to the renderer.

- **Upsert error: “no unique or exclusion constraint…”**  
  Add PK or UNIQUE on `(user_email, id)`:
  ```sql
  ALTER TABLE transactions DROP CONSTRAINT IF EXISTS transactions_pkey;
  ALTER TABLE transactions ALTER COLUMN id DROP DEFAULT;
  ALTER TABLE transactions ADD PRIMARY KEY (user_email, id);
  -- or:
  -- CREATE UNIQUE INDEX IF NOT EXISTS transactions_useremail_id_uq ON transactions(user_email,id);
  ```

- **IDs become 1..N instead of CSV values**  
  Ensure the model has `gorm:"autoIncrement:false"` on `ID` and the PK/UNIQUE is on `(user_email,id)`.

- **Lambda: “no such file or directory /var/task/data/transactions.csv”**  
  Copy `data/` into the image (Dockerfile.lambda) or mount it when running locally. For S3 sources, pass `s3://…` and configure S3 credentials/endpoint.

---

## License

MIT (or your preferred license)


**Enjoy!** The project was largely achieved through my own knowledge and research for the necessary information, but also with the help of artificial intelligence (chatgpt) for infrastructure development. Thank you so much for coming this far and for everyone's time.