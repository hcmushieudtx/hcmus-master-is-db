# Online Bookstore — Multi-Database System (N06)

> HCMUS Master — Information Systems Database Final Project
> Group N06 — Polyglot Persistence Architecture

---

## Table of Contents

- [System Overview](#system-overview)
- [Architecture](#architecture)
- [Backend](#backend)
  - [Technology Stack](#technology-stack)
  - [Database Responsibilities](#database-responsibilities)
  - [Project Structure](#project-structure)
  - [API Reference](#api-reference)
  - [Getting Started](#getting-started)
  - [Configuration](#configuration)
  - [Database Migrations](#database-migrations)
  - [Makefile Commands](#makefile-commands)
- [Frontend](#frontend)

---

## System Overview

The **Online Bookstore System** is a full-stack e-commerce application built around a **Polyglot Persistence** architecture — each business domain uses the database type best suited to its data characteristics.

| # | Data Characteristic | Technical Requirement | Selected Database |
|---|--------------------|-----------------------|-------------------|
| 1 | Transactional Data | Strong ACID, referential integrity | **PostgreSQL** |
| 2 | Catalog Data | Polymorphic schema, high read frequency | **MongoDB** |
| 3 | Graph Data | Complex multi-dimensional relationships, graph traversal | **Neo4j** |
| 4 | Ephemeral / Cached Data | Sub-millisecond in-memory access, short TTL | **Redis** |

**Actors**

| Actor | Type | Capabilities |
|-------|------|-------------|
| Guest | Unauthenticated | Browse catalog, search books, view recommendations |
| Customer | Authenticated (`role: user`) | Full shopping flow: cart, checkout, order history, profile |
| Admin | Authenticated (`role: admin`) | Catalog management, order tracking, user management, analytics |

---

## Architecture

```
┌──────────────┐        REST / JSON        ┌──────────────────────────────────────┐
│  Next.js FE  │ ────────────────────────► │           Gin HTTP Server            │
│  (Port 3000) │ ◄──────────────────────── │    internal/server  (Port 8080)      │
└──────────────┘                           └──────────────┬───────────────────────┘
                                                          │ JWT Middleware
                                                          │ (role: user | admin)
                                           ┌──────────────▼───────────────────────┐
                                           │          internal/domain             │
                                           │   Repository Interfaces + Models     │
                                           └──────┬──────┬──────┬────────────────┘
                                                  │      │      │      │
                                        ┌─────────▼─┐ ┌──▼──┐ ┌▼───┐ ┌▼──────┐
                                        │ PostgreSQL│ │Mongo│ │Neo4│ │Redis  │
                                        │ Users &   │ │Book │ │j   │ │Session│
                                        │ Orders    │ │Cat. │ │Rec.│ │Cart   │
                                        └───────────┘ └─────┘ └────┘ └───────┘
```

---

## Backend

### Technology Stack

| Layer | Technology |
|-------|-----------|
| Language | Go 1.22 |
| Web Framework | Gin |
| CLI | Cobra |
| Configuration | Viper (YAML + env var overrides) |
| PostgreSQL ORM | GORM + golang-migrate |
| PostgreSQL Typed Queries | sqlc |
| MongoDB Driver | go.mongodb.org/mongo-driver |
| Neo4j Driver | neo4j-go-driver/v5 |
| Redis Client | go-redis/v9 |
| Authentication | JWT (golang-jwt/jwt/v5) + bcrypt |
| Logging | Zap (uber-go/zap) |

### Database Responsibilities

#### PostgreSQL — Users & Orders (Groups A, D)
Handles all transactional data requiring full ACID guarantees.

- `users` table — accounts with role (`user` / `admin`), bcrypt password hash, active flag
- `books_ref` table — bridge table linking MongoDB book IDs to live stock quantities and prices (used for `SELECT ... FOR UPDATE` during checkout)
- `orders` table — order headers with status lifecycle (`pending → confirmed → shipping → completed | cancelled`)
- `order_items` table — line items with price snapshot at purchase time (immutable for history integrity)

#### MongoDB — Book Catalog (Group B)
Stores the flexible, polymorphic book catalog. A single document holds all variant attributes (binding type, dimensions, set volumes, signed edition notes, etc.) without schema migrations.

- Full-text search index on `title` and `authors`
- Compound index on `genres` + `price` for filtered browsing
- Sparse index on `series_name` + `volume_order` for series lookups

#### Neo4j — Recommendation Engine (Group E)
Models books as a graph with typed, weighted relationships:

- `(:Book)-[:SAME_GENRE {weight:3}]->(:Genre)`
- `(:Book)-[:SAME_AUTHOR {weight:2}]->(:Author)`
- `(:Book)-[:SAME_PUBLISHER {weight:1}]->(:Publisher)`
- `(:Book)-[:IN_SERIES {volume_order}]->(:Series)`

Similar-book scoring: `score = (shared genres × 3) + (shared authors × 2) + (shared publisher × 1)`

#### Redis — Sessions, Cart & Trending (Groups A, B, C, E)

| Key Pattern | Structure | Purpose |
|-------------|-----------|---------|
| `session:<userID>` | String | Active JWT token with TTL |
| `blacklist:<token>` | String | Revoked tokens (logout) |
| `cart:<userID>` | Hash | Shopping cart items (bookID → JSON) |
| `trending:books` | Sorted Set | Sales score per book (ZINCRBY on checkout) |
| `trending:top10` | String | Cached JSON top-10 list |

---

### Project Structure

```
backend/
├── main.go                          # Entry point → cmd.Run
├── go.mod / go.sum                  # Module: bookstore/backend
├── Makefile                         # Developer commands
├── .env.example                     # Environment variable reference
│
├── cmd/
│   ├── cmd.go                       # Cobra root + .env loading + config init
│   └── server.go                    # DB connections, repo wiring, Gin server, graceful shutdown
│
├── config/
│   ├── config.go                    # Typed Config struct + Viper loader
│   └── default.go                   # Embedded YAML defaults
│
├── internal/
│   ├── server/                      # HTTP layer (Gin handlers)
│   │   ├── server.go                # Route groups: public / RequireAuth / RequireUser / RequireAdmin
│   │   ├── service.go               # Service struct (all repo dependencies + jwtCfg)
│   │   ├── response.go              # Unified JSON response helpers
│   │   ├── user.go                  # Register, Login, Logout, GetProfile, UpdateProfile
│   │   ├── book.go                  # SearchBooks, GetBookDetail, GetNewBooks
│   │   ├── cart.go                  # AddToCart, GetCart, UpdateCartItem, RemoveCartItem
│   │   ├── order.go                 # Checkout (atomic TX), GetOrderHistory, GetOrderDetail
│   │   ├── recommendation.go        # GetSimilarBooks, GetSeriesBooks, GetTrending
│   │   ├── admin_book.go            # AdminCreateBook, AdminUpdateBook, AdminDeleteBook, AdminUpdateStock
│   │   ├── admin_order.go           # AdminListOrders, AdminGetOrder, AdminUpdateOrderStatus
│   │   └── admin_user.go            # AdminListUsers, AdminGetUser, AdminDeactivateUser, AdminGetSales
│   │
│   ├── domain/
│   │   ├── model.go                 # User, Book, BookRef, CartItem, Order, OrderItem, BookNode, ...
│   │   ├── repository.go            # All repository interfaces + PostgresTransactor
│   │   └── dto.go                   # Request / Response DTOs for all endpoints
│   │
│   ├── middleware/
│   │   ├── auth.go                  # RequireAuth, RequireUser, RequireAdmin
│   │   └── constants.go             # Context keys (userID, userRole, token)
│   │
│   └── repository/
│       ├── postgres/                # GORM-backed: user.go, order.go, postgres.go (Transaction)
│       ├── mongo/                   # MongoDB: book.go (search, CRUD)
│       ├── neo4j/                   # Neo4j: recommendation.go (graph traversal + upsert)
│       └── redis/                   # Redis: session.go, cart.go, trending.go
│
├── utils/
│   ├── database/                    # Connection factories: ConnectPostgres, ConnectMongo, ConnectNeo4j, ConnectRedis
│   ├── token/jwt.go                 # GenerateToken(userID, email, role), ParseToken, Claims
│   ├── password/bcrypt.go           # HashPassword, CheckPassword
│   ├── log/log.go                   # Zap logger factory
│   └── server/server.go             # Port / timeout constants
│
└── db/
    ├── postgres/
    │   ├── sqlc.yaml                # sqlc code-generation config
    │   ├── queries/                 # Named SQL queries (user.sql, order.sql)
    │   ├── migrations/              # golang-migrate up/down pairs (3 migrations)
    │   └── store/                   # sqlc-generated typed query code (run: make sqlc-generate)
    ├── mongo/indexes/               # MongoDB index definitions (books_indexes.json)
    └── neo4j/
        ├── migrations/              # Cypher constraint + index setup (up/down)
        └── queries/                 # Reusable Cypher queries (similar_books, series_books)
```

---

### API Reference

All endpoints are prefixed with `/api/v1`.

#### Public (no authentication required)

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/auth/register` | Create a new customer account |
| `POST` | `/auth/login` | Authenticate and receive JWT |
| `GET` | `/books` | Search and filter books (MongoDB) |
| `GET` | `/books/new` | Newest books |
| `GET` | `/books/:id` | Book detail |
| `GET` | `/books/:id/similar` | Neo4j similar-book recommendations (Top 10) |
| `GET` | `/books/:id/series` | All volumes in the same series |
| `GET` | `/trending` | Redis top-10 bestsellers |

#### Customer (JWT required, `role: user` only)

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/auth/logout` | Revoke JWT (Redis blacklist) |
| `GET` | `/users/me` | View own profile |
| `PUT` | `/users/me` | Update name / phone / default address |
| `GET` | `/cart` | Get cart contents |
| `POST` | `/cart` | Add book to cart |
| `PUT` | `/cart/:bookId` | Update item quantity |
| `DELETE` | `/cart/:bookId` | Remove item from cart |
| `POST` | `/orders/checkout` | Atomic checkout (PG transaction + Redis clear) |
| `GET` | `/orders` | List own orders |
| `GET` | `/orders/:id` | Order detail |

> Admin accounts (`role: admin`) are blocked from all customer endpoints with HTTP 403. The only shared endpoint is `POST /auth/logout`.

#### Admin (JWT required, `role: admin` only)

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/admin/books` | List all books with stock info |
| `POST` | `/admin/books` | Create book (MongoDB + PostgreSQL + Neo4j) |
| `PUT` | `/admin/books/:id` | Update book metadata / price |
| `DELETE` | `/admin/books/:id` | Soft-delete book (`is_active=false`) |
| `PATCH` | `/admin/books/:id/stock` | Set stock quantity |
| `GET` | `/admin/orders` | List all orders (filter by status) |
| `GET` | `/admin/orders/:id` | Full order detail |
| `PATCH` | `/admin/orders/:id/status` | Update order status |
| `GET` | `/admin/users` | List all users |
| `GET` | `/admin/users/:id` | View any user profile |
| `PATCH` | `/admin/users/:id/deactivate` | Activate / deactivate account |
| `GET` | `/admin/analytics/trending` | Trending scores from Redis |
| `GET` | `/admin/analytics/sales` | Sales summary by date range |

---

### Getting Started

**Prerequisites**

- Go 1.22+
- PostgreSQL 14+
- MongoDB 6+
- Neo4j 5+ (Community or Enterprise)
- Redis 7+
- [`golang-migrate`](https://github.com/golang-migrate/migrate) CLI
- [`sqlc`](https://sqlc.dev) CLI (optional — only needed to regenerate typed query code)

**1. Clone and install dependencies**

```bash
git clone <repo-url>
cd hcmus-master-is-db/backend
go mod tidy
```

**2. Configure environment**

```bash
cp .env.example .env
# Edit .env with your database credentials
```

**3. Apply PostgreSQL migrations**

```bash
make migrate-up
```

**4. (Optional) Apply Neo4j constraints and indexes**

Run the Cypher file against your Neo4j instance:
```bash
# Using cypher-shell
cypher-shell -u neo4j -p <password> < db/neo4j/migrations/001_init_graph.up.cypher
```

**5. (Optional) Apply MongoDB indexes**

Use `mongosh` or your preferred client with the definitions in `db/mongo/indexes/books_indexes.json`.

**6. Start the server**

```bash
make run
# Server starts on http://localhost:8080
```

---

### Configuration

All settings have embedded defaults in `config/default.go` and can be overridden via environment variables using `__` as the nested key separator.

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `ENV` | `development` | Runtime environment |
| `SERVER__PORT` | `8080` | HTTP listen port |
| `POSTGRES__HOST` | `localhost` | PostgreSQL host |
| `POSTGRES__PORT` | `5432` | PostgreSQL port |
| `POSTGRES__DB` | `bookstore` | Database name |
| `POSTGRES__USER` | `postgres` | Username |
| `POSTGRES__PASSWORD` | `secret` | Password |
| `POSTGRES__SSLMODE` | `disable` | SSL mode |
| `MONGO__URI` | `mongodb://localhost:27017` | MongoDB connection URI |
| `MONGO__DB` | `bookstore` | Database name |
| `NEO4J__URI` | `bolt://localhost:7687` | Neo4j Bolt URI |
| `NEO4J__USER` | `neo4j` | Username |
| `NEO4J__PASSWORD` | `password` | Password |
| `REDIS__ADDR` | `localhost:6379` | Redis address |
| `REDIS__PASSWORD` | _(empty)_ | Redis password |
| `REDIS__DB` | `0` | Redis logical DB index |
| `JWT__SECRET` | _(change this!)_ | HMAC signing secret |
| `JWT__ACCESS_TTL` | `24h` | Token expiry duration |
| `LOGGER__LEVEL` | `info` | Log level (`debug`, `info`, `warn`, `error`) |

---

### Database Migrations

PostgreSQL migrations are managed by **golang-migrate** and live in `db/postgres/migrations/`.

```bash
# Apply all pending migrations
make migrate-up

# Roll back one migration
make migrate-down

# Create a new migration pair
make migrate-create NAME=add_reviews_table
```

Migration files:

| File | Description |
|------|-------------|
| `202604231400_create_users.up.sql` | `users` table with `user_role` enum, indexes |
| `202604231401_create_books_ref.up.sql` | `books_ref` bridge table (MongoDB ID ↔ stock/price) |
| `202604231402_create_orders.up.sql` | `orders` + `order_items` tables with `order_status` enum |

---

### Makefile Commands

| Command | Description |
|---------|-------------|
| `make run` | Start the API server (reads `.env`) |
| `make build` | Compile binary to `bin/bookstore-api` |
| `make tidy` | Run `go mod tidy` |
| `make migrate-up` | Apply all pending PostgreSQL migrations |
| `make migrate-down` | Roll back one PostgreSQL migration |
| `make migrate-create NAME=<name>` | Create a new migration pair |
| `make sqlc-generate` | Regenerate typed query code from `db/postgres/queries/` |
| `make clean` | Remove build artifacts |

---

## Frontend

> Documentation for the frontend (Next.js) will be added here once implemented.