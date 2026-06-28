### 1. API Contract Decisions

> **Question:** What was one non-obvious design decision you made in the API surface—a naming choice, a response shape, a status code—and why did you make it?
> 
> 

**Answer:**
For `PATCH /menu/items/:id`, instead of passing a raw boolean like `{"is_available": false}`, I chose to accept an explicit status enum string like `{"availability": "out_of_stock"}`. While a boolean is simpler initially, F&B contexts often require more states later on, such as "discontinued" or "seasonal_unavailable". Designing the contract around an enum prevents breaking the API contract when these inevitable business requirements arise.

### 2. Versioning

> **Question:** If a mobile client were already consuming GET /menu and you needed to change the response shape in a breaking way, how would you handle that?
> 
> 

**Answer:**
I would introduce path-based versioning by introducing a new endpoint route: `/v2/menu`. The existing mobile client would safely continue hitting `/v1/menu` (or the unversioned route mapping to the legacy handler) without breaking. I would then collaborate with the mobile team to establish a deprecation timeline for the older version once the new application build is fully adopted by users.

### 3. What you'd do differently with more time

> **Question:** Name one thing you would change or add if you had another two hours. Be specific.
> 
> 

**Answer:**
Given an extra two hours, I would implement database transactions ($TX$) across the order creation flow. Right now, if an order successfully writes to the database but the RabbitMQ message publishing fails, the system enters an inconsistent state where the kitchen is never notified. Implementing an Outbox Pattern—where the event is saved to the Postgres database within the same transaction as the order and then pumped to RabbitMQ—would guarantee at-least-once delivery.

### 4. Production Gap

> **Question:** What is the most significant thing missing from this service that would concern you before shipping it to real users?
> 
> 

**Answer:**
The most glaring production gap is the lack of concurrency handling and inventory locks during order placement. If two users simultaneously order the very last available slice of cake, a race condition could allow both orders to succeed, resulting in a terrible real-world customer experience. Before shipping, I would introduce a pessimistic locking mechanism (`SELECT ... FOR UPDATE`) or an inventory deduction check during the order placement transaction.

---

## ⚡ Pro-Tips for Nailing the Execution

1. **Docker Compose is your friend:** Provide a `docker-compose.yml` that spins up a local `postgres` instance and a `rabbitmq` instance. This makes it so the reviewer can just run `docker-compose up` and immediately test your app.


2. **Structured Logging:** Use Go's built-in `log/slog` (introduced in Go 1.21, which fits their requirement). Printing JSON logs in both the API and the worker shows you care about production monitoring.


3. **Idempotency/Validation:** In your `POST /orders` endpoint, make sure you double-check that every item id in the payload actually exists in the database and has its `availability` set to `in_stock` before saving the order.


## 🚀 How to Run Locally

### Step 1: Start Infrastructure Containers

Launch the database and messaging infrastructure in the background:

```bash
docker-compose up -d

```

### Step 2: Download Dependencies

```bash
go mod tidy

```

### Step 3: Run the Async Worker Application

In a separate terminal window, launch the background worker:

```bash
go run cmd/worker/main.go

```

### Step 4: Run the Primary REST API Application

In another terminal window, start up the core web server:

```bash
go run cmd/api/main.go

```

---

## 🧪 Testing Your Work (cURL Examples)

#### 1. Fetching Entire Menu

```bash
curl http://localhost:8080/menu

```

#### 2. Modify Item Availability to Out of Stock

```bash
curl -X PATCH http://localhost:8080/menu/items/1 \
  -H "Content-Type: application/json" \
  -d '{"availability": "out_of_stock"}'

```

#### 3. Submit a New Order

```bash
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{"items": [{"menu_item_id": 2, "quantity": 2}, {"menu_item_id": 3, "quantity": 1}]}'

```

*Note: Check your running **Worker terminal** immediately after firing this post request; you'll see a clean, structured JSON receipt confirmation log instantly process via the message queue!*

#### 4. Checking Tracking Information Statuses

```bash
curl http://localhost:8080/orders/1

```