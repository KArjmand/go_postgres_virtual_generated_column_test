# PostgreSQL Virtual Generated Column Test

A Go server comparing CPU usage and network performance between:
1. **PostgreSQL Virtual Generated Column** - `total_cents` computed by the database
2. **Application-level Calculation** - `total_cents` computed in Go

## Architecture

This project follows **Clean Architecture** principles inspired by [Three Dots Labs](https://threedots.tech/post/microservices-or-monolith-its-detail/).

```
├── cmd/
│   └── monolith/               # Application entry point
│       └── main.go
├── pkg/
│   ├── common/                 # Shared utilities
│   │   ├── cmd/                # Router, signals
│   │   └── http/               # HTTP error handling
│   └── invoices/               # Invoices bounded context
│       ├── domain/             # Domain layer (entities, repository interfaces)
│       │   └── invoices/
│       ├── application/        # Application layer (use cases, services)
│       ├── infrastructure/     # Infrastructure layer (PostgreSQL implementation)
│       │   └── postgres/
│       └── interfaces/         # Interface layer (HTTP handlers)
│           └── http/
├── docker-compose.yml
├── Dockerfile
└── Makefile
```

### Layers

- **Domain**: Core business entities (`Invoice`), repository interfaces, value objects
- **Application**: Business logic, orchestrates domain objects
- **Infrastructure**: Database implementations, external services
- **Interfaces**: HTTP handlers, API endpoints

## Prerequisites

You need **Docker** and **docker-compose** installed.

Everything runs in Docker containers, so you don't need Go installed locally.

## Running

Just run:

```bash
make up
```

It will build the Docker image and start the server with PostgreSQL.

### Running locally

1. Create `.env` file:
```bash
cp .env.example .env
```

2. Start PostgreSQL (or use your own):
```bash
docker compose up postgres -d
```

3. Run the server:
```bash
make run
```

The server will:
- Create both tables automatically
- Seed 100,000 rows of random data (configurable via `SEED_COUNT`)
- Start listening on port 8080

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/invoices/virtual` | Uses PostgreSQL virtual generated column |
| GET | `/api/invoices/calculated` | Calculates `total_cents` in Go |
| GET | `/api/stats` | Returns row counts for both tables |
| GET | `/health` | Health check endpoint |

## Response Format

```json
{
  "data": [...],
  "count": 10000,
  "query_time_ms": 12.5,
  "total_time_ms": 45.2,
  "cpu_time_ns": 45200000,
  "memory_bytes": 1234567
}
```

## Benchmarking

Use `curl` or tools like `wrk`/`hey` to compare:

```bash
# Compare response times
curl -w "@curl-format.txt" -o /dev/null -s http://localhost:8080/api/invoices/virtual
curl -w "@curl-format.txt" -o /dev/null -s http://localhost:8080/api/invoices/calculated

# Load testing with hey
hey -n 100 -c 10 http://localhost:8080/api/invoices/virtual
hey -n 100 -c 10 http://localhost:8080/api/invoices/calculated
```

## Database Schema

### Table with Virtual Generated Column
```sql
CREATE TABLE invoices_with_virtual (
    id           BIGSERIAL PRIMARY KEY,
    customer_id  BIGINT NOT NULL,
    amount_cents BIGINT NOT NULL,
    tax_rate     NUMERIC(4,2) NOT NULL,
    total_cents  BIGINT GENERATED ALWAYS AS (
        ROUND(amount_cents * (1 + tax_rate))
    ) STORED
);
```

### Table without Virtual Generated Column
```sql
CREATE TABLE invoices_without_virtual (
    id           BIGSERIAL PRIMARY KEY,
    customer_id  BIGINT NOT NULL,
    amount_cents BIGINT NOT NULL,
    tax_rate     NUMERIC(4,2) NOT NULL
);
```

## Benchmark Results

### Test Environment
- **PostgreSQL 18** (Alpine)
- **Go 1.25.4**
- **Data**: 100,000 invoices seeded
- **Query**: Fetch 10,000 rows per request

### Single Request Comparison

```bash
curl -s http://localhost:8080/api/benchmark | jq
```

| Metric | Virtual Column | App Calculated | Winner |
|--------|---------------|----------------|--------|
| Query Time | 6.67ms | 5.96ms | **Calculated** (-11%) |
| Total Time | 6.88ms | 6.09ms | **Calculated** (-11%) |
| Memory | 1,563KB | ~0KB | **Calculated** |
| Response Size | 875KB | 875KB | Tie |

### Load Test (10 concurrent users, 200 requests)

```bash
hey -n 200 -c 10 http://localhost:8080/api/invoices/virtual
hey -n 200 -c 10 http://localhost:8080/api/invoices/calculated
```

| Metric | Virtual Column | App Calculated | Winner |
|--------|---------------|----------------|--------|
| Requests/sec | 134.03 | 106.93 | **Virtual** (+25%) |
| Avg Latency | 68.0ms | 57.4ms | **Calculated** (-16%) |
| p50 Latency | 63.0ms | 51.5ms | **Calculated** |
| p95 Latency | 101.6ms | 80.8ms | **Calculated** |
| p99 Latency | 286.8ms | 310.6ms | **Virtual** |

### High Concurrency Test (50 concurrent users, 500 requests)

```bash
hey -n 500 -c 50 http://localhost:8080/api/invoices/virtual
hey -n 500 -c 50 http://localhost:8080/api/invoices/calculated
```

| Metric | Virtual Column | App Calculated | Winner |
|--------|---------------|----------------|--------|
| Requests/sec | 6.03 | 11.16 | **Calculated** (+85%) |
| Avg Latency | 440ms | 338ms | **Calculated** (-23%) |
| p50 Latency | 277ms | 240ms | **Calculated** |
| p95 Latency | 488ms | 431ms | **Calculated** |
| p99 Latency | 10.67s | 6.51s | **Calculated** |

## Analysis

### Why App Calculation Wins in This Test

1. **Simple formula**: `ROUND(amount * (1 + tax_rate))` is trivial for Go (~nanoseconds)
2. **Less data transfer**: 4 columns vs 5 columns from PostgreSQL
3. **STORED columns**: Computation happens at write time, not read time

### When to Use Virtual Generated Columns

Virtual columns shine when:
- **Complex calculations** that would be expensive in application code
- **Indexing derived values**: `CREATE INDEX ON invoices(total_cents)`
- **Multiple consumers** need the same computed value
- **Data consistency**: Formula enforced by database, not duplicated across services

### Key Takeaway

Both approaches handle **10K rows in <500ms under 50 concurrent users** without any cache layer. This validates the premise that **PostgreSQL 18 is fast enough** to serve many workloads directly, potentially eliminating the need for Redis or other cache tiers.

## References

- [Why Postgres 18 Let Me Delete a Whole Cache Tier](https://medium.com/@ArkProtocol1/why-postgres-18-let-me-delete-a-whole-cache-tier-bba2b4f1c742) - Article that inspired this benchmark
- [PostgreSQL Generated Columns](https://www.postgresql.org/docs/current/ddl-generated-columns.html)

## License

MIT
