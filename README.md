# Appointment Booking System

A RESTful API built in **Go** using the **Gin** framework and **PostgreSQL** for a 30-minute appointment booking system. Coaches set their weekly availability, and users browse available slots and book appointments — with full concurrency protection, timezone handling, and cancellation support.

---

## Table of Contents

- [Design Choices](#design-choices)
- [Architecture](#architecture)
- [Database Schema](#database-schema)
- [Setup & Run](#setup--run)
- [API Documentation](#api-documentation)
- [Concurrency Handling](#concurrency-handling)
- [Timezone Handling](#timezone-handling)
- [Running Tests](#running-tests)

---

## Design Choices

### Why Layered Architecture?
The codebase follows a strict **Handler → Service → Repository** layered pattern. Each layer has a single responsibility:
- **Handlers** parse HTTP requests, validate input, and format responses.
- **Services** contain business logic (slot generation, overlap detection, booking validation).
- **Repositories** are the only layer that touches the database.

This separation makes the code testable (each layer can be tested independently), maintainable, and easy to extend.

### Why PostgreSQL Partial Unique Index?
Double-booking prevention uses a **two-layer defense**:
1. **Application-level**: `SELECT ... FOR UPDATE` row-level locking inside a transaction.
2. **Database-level**: A partial unique index `UNIQUE(coach_id, start_time) WHERE status = 'booked'` ensures the DB itself rejects duplicates even if the application logic has a bug.

The partial index only applies to active (`booked`) bookings — cancelled bookings don't block the slot.

### Why Soft Deletes for Cancellation?
Bookings are cancelled by setting `status = 'cancelled'` rather than deleting the row. This preserves audit history and allows the partial unique index to free the slot for rebooking.

### Assumptions
- A **coach** must be created before setting availability or booking slots.
- Availability is defined as **weekly recurring** windows (e.g., "Every Monday 9 AM – 3 PM").
- All database times are stored in **UTC**. Coach availability times are interpreted in the coach's configured timezone.
- Users are identified by `user_id` passed in the request (no authentication layer).
- Each booking slot is exactly **30 minutes**.

---

## Architecture

```
Client → Router (Gin) → Middleware (Logger, Recovery, CORS) → Handlers → Services → Repositories → PostgreSQL
```

```
appointment-booking-system/
├── main.go                          # Entry point
├── config/config.go                 # Environment-based configuration
├── router/router.go                 # Route definitions + middleware
├── middleware/
│   ├── logger.go                    # Request/response logging (Zap)
│   └── cors.go                      # CORS configuration
├── handlers/                        # HTTP request/response handling
├── business/                        # Business logic / services
├── repositories/                    # Database access layer
├── models/
│   ├── dbModels.go                  # GORM entity models
│   ├── requestModels.go             # Request DTOs with validation
│   ├── responseModels.go            # Response DTOs
│   └── genericResponse.go          # Standardized API response wrapper
├── commons/
│   ├── commons.go                   # Slot generation, time utilities
│   └── constants/                   # Routes, errors, app constants
├── tests/                           # Integration + concurrency tests
├── Makefile
└── .env.example
```

---

## Database Schema

Three tables:

| Table | Purpose |
|-------|---------|
| `coaches` | Coach profiles (name, email, timezone) |
| `coach_availabilities` | Weekly recurring availability windows |
| `bookings` | User appointment bookings |

Key indexes:
- `idx_unique_active_booking`: `UNIQUE(coach_id, start_time) WHERE status = 'booked'` — prevents double booking at DB level
- `idx_availability_coach`: `(coach_id, day_of_week)` — fast availability lookups
- `idx_bookings_user`: `(user_id, status)` — fast user booking queries

---

## Setup & Run

### Prerequisites
- **Go** 1.21+ installed
- **PostgreSQL** running locally (or accessible via network)

### 1. Clone the repository
```bash
git clone <repository-url>
cd appointment-booking-system
```

### 2. Create the database
```bash
psql -U postgres -c "CREATE DATABASE appointment_booking;"
```

For running tests, also create:
```bash
psql -U postgres -c "CREATE DATABASE appointment_booking_test;"
```

### 3. Configure environment
```bash
cp .env.example .env
```

Edit `.env` with your PostgreSQL credentials:
```
APP_PORT=8080
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=appointment_booking
DB_SSL_MODE=disable
```

### 4. Install dependencies
```bash
go mod tidy
```

### 5. Run the server
```bash
make run
# or
go run main.go
```

The server starts on `http://localhost:8080`. Database tables are auto-migrated on startup.

---

## API Documentation

All responses follow this format:
```json
{
  "status": "success" | "error",
  "message": "Human-readable message",
  "data": { ... }
}
```

### Health Check

```
GET /health
```

**Response** `200 OK`:
```json
{"status": "success", "message": "service is healthy"}
```

---

### Create Coach

```
POST /v1/coaches
```

**Request Body**:
```json
{
  "name": "Coach A",
  "email": "coach.a@example.com",
  "timezone": "America/New_York"
}
```

**Response** `201 Created`:
```json
{
  "status": "success",
  "message": "coach created successfully",
  "data": {
    "id": 1,
    "name": "Coach A",
    "email": "coach.a@example.com",
    "timezone": "America/New_York"
  }
}
```

---

### Set Coach Availability

```
POST /v1/coaches/availability
```

**Request Body**:
```json
{
  "coach_id": 1,
  "day_of_week": 1,
  "start_time": "09:00",
  "end_time": "14:00"
}
```

`day_of_week`: 0 = Sunday, 1 = Monday, ..., 6 = Saturday.
`start_time`/`end_time`: Accepts both `HH:MM` and `HH:MM:SS` formats.

**Response** `201 Created`:
```json
{
  "status": "success",
  "message": "availability set successfully",
  "data": {
    "id": 1,
    "coach_id": 1,
    "day_of_week": 1,
    "start_time": "09:00:00",
    "end_time": "14:00:00"
  }
}
```

**Error Responses**:
- `400` — Invalid time range, invalid day, bad format
- `409` — Overlapping availability window

---

### Get Coach Availability

```
GET /v1/coaches/availability?coach_id=1
```

**Response** `200 OK`:
```json
{
  "status": "success",
  "message": "availability retrieved successfully",
  "data": [
    {
      "id": 1,
      "coach_id": 1,
      "day_of_week": 1,
      "start_time": "09:00:00",
      "end_time": "14:00:00"
    }
  ]
}
```

---

### Get Available Slots

```
GET /v1/users/slots?coach_id=1&date=2026-10-26&timezone=America/New_York
```

| Param | Required | Description |
|-------|----------|-------------|
| `coach_id` | Yes | Coach to check |
| `date` | Yes | Date in `YYYY-MM-DD` format |
| `timezone` | No | Convert slots to this timezone (default: UTC) |

**Response** `200 OK`:
```json
{
  "status": "success",
  "message": "available slots retrieved",
  "data": [
    {"start_time": "2026-10-26T09:00:00Z", "end_time": "2026-10-26T09:30:00Z"},
    {"start_time": "2026-10-26T09:30:00Z", "end_time": "2026-10-26T10:00:00Z"}
  ]
}
```

**Error Responses**:
- `400` — Invalid date, past date, invalid timezone
- `404` — Coach not found

---

### Book an Appointment

```
POST /v1/users/bookings
```

**Request Body**:
```json
{
  "user_id": 101,
  "coach_id": 1,
  "start_time": "2026-10-26T09:30:00Z",
  "user_timezone": "America/New_York"
}
```

`start_time` must be in RFC3339 format, aligned to `:00` or `:30` minutes, and fall within the coach's availability.

**Response** `201 Created`:
```json
{
  "status": "success",
  "message": "booking created successfully",
  "data": {
    "id": 1,
    "user_id": 101,
    "coach_id": 1,
    "start_time": "2026-10-26T09:30:00Z",
    "end_time": "2026-10-26T10:00:00Z",
    "status": "booked",
    "user_timezone": "America/New_York"
  }
}
```

**Error Responses**:
- `400` — Invalid input, misaligned time, past slot, outside availability
- `404` — Coach not found
- `409` — Slot already booked (double booking prevented)

---

### Get User Bookings

```
GET /v1/users/bookings?user_id=101
```

**Response** `200 OK`:
```json
{
  "status": "success",
  "message": "bookings retrieved successfully",
  "data": [
    {
      "id": 1,
      "user_id": 101,
      "coach_id": 1,
      "start_time": "2026-10-26T09:30:00Z",
      "end_time": "2026-10-26T10:00:00Z",
      "status": "booked",
      "user_timezone": "America/New_York",
      "created_at": "2026-03-22T10:00:00Z"
    }
  ]
}
```

---

### Cancel a Booking

```
DELETE /v1/users/bookings/:id?user_id=101
```

**Response** `200 OK`:
```json
{"status": "success", "message": "booking cancelled successfully"}
```

**Error Responses**:
- `400` — Missing or invalid user_id / booking ID
- `403` — Booking does not belong to this user
- `404` — Booking not found
- `409` — Booking is already cancelled

---

## Concurrency Handling

The system prevents double-booking through a **two-layer defense**:

### Layer 1: Application-Level (Transaction + Row Lock)
```go
tx.Clauses(clause.Locking{Strength: "UPDATE"}).
    Where("coach_id = ? AND start_time = ? AND status = ?", ...).
    First(&existing)
```
Uses PostgreSQL `SELECT ... FOR UPDATE` to acquire a row-level lock inside a transaction. If another concurrent transaction tries to book the same slot, it blocks until the first transaction completes.

### Layer 2: Database-Level (Partial Unique Index)
```sql
CREATE UNIQUE INDEX idx_unique_active_booking
ON bookings(coach_id, start_time) WHERE status = 'booked';
```
Even if the application logic has a race condition bug, the database itself rejects duplicate active bookings.

The concurrency test (`tests/concurrency_test.go`) spawns 10 goroutines all trying to book the same slot simultaneously and verifies exactly 1 succeeds with `201 Created` while the other 9 receive `409 Conflict`.

---

## Timezone Handling

- **Storage**: All times in the database are in **UTC**.
- **Coach availability**: `start_time`/`end_time` are clock times in the coach's configured timezone. When generating slots, these are converted to UTC for the specific date.
- **Slot viewing**: Users can pass a `timezone` query parameter to see slots converted to their local time.
- **Booking**: The `user_timezone` field is stored with each booking for display purposes.

Example: A coach in `America/New_York` is available Monday 9 AM – 12 PM. On a Monday in October (EDT, UTC-4), the generated slots are 13:00–15:30 UTC.

---

## Running Tests

### Unit tests (no database required)
```bash
go test ./commons/... -v
```

### Integration tests (requires PostgreSQL)
Create the test database first:
```bash
psql -U postgres -c "CREATE DATABASE appointment_booking_test;"
```

Run all tests:
```bash
go test ./tests/... -v
```

Run with race detector (validates concurrency safety):
```bash
go test ./tests/... -v -race
```

### Test coverage
```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```
