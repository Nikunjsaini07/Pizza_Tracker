# 🍕 Real-Time Go Pizza Tracker

A secure, concurrent, and real-time pizza ordering and tracking application built from scratch in Go. 

This application allows customers to place orders for multiple pizzas and track their preparation status in real-time, while providing an administrative dashboard for order management and status updates protected by session authentication.

---

## 🛠️ Tech Stack

*   **Language**: Go (Golang)
*   **Web Framework**: Gin Gonic
*   **Database**: SQLite
*   **ORM**: GORM
*   **Real-time Communication**: Server-Sent Events (SSE) via HTML5 EventSource
*   **Security**: Bcrypt password hashing & Session Middleware
*   **Frontend**: HTML5, Vanilla CSS & Server-Side Rendering (SSR) via Go HTML Templates

---

## ✨ Features

1.  **Dynamic Order Form**: Supports multi-pizza ordering in a single submission with input binding and validation.
2.  **Unpredictable Order IDs**: Uses automated GORM hooks and the `shortid` library to generate secure, alphanumeric string IDs (e.g. `y3bK8R9`) instead of sequential integers.
3.  **Real-Time Status Tracker**: Streams server updates instantly to the customer's browser via Server-Sent Events (SSE) when order statuses are modified.
4.  **Admin Dashboard**: Protected by session middleware. Admins can view order details, update statuses, or cancel/delete orders.
5.  **Secure Login**: Protects administrative actions using salted `bcrypt` password hashes and GORM database-persisted session cookies.

---

## 📂 Project Structure

```
pizza-tracker/
├── templates/                    # Server-Side HTML Templates
│   ├── order.html                # Pizza Order Form Page
│   ├── customer.html             # Real-time Customer Tracking Screen
│   ├── admin.html                # Admin dashboard order table
│   └── login.html                # Secure admin login form
├── internal/
│   └── models/                   # GORM Schemas & Logic
│       ├── order.go              # Order/OrderItem structs & hooks
│       └── user.go               # User struct & bcrypt helpers
├── main.go                       # Web server, DB init & route definitions
├── notifications.go              # SSE Broadcaster & Mutex client lists
├── go.mod                        # Go dependency declaration
└── pizza.db                      # SQLite database file (auto-generated)
```

---

## 🚀 Getting Started

### Prerequisites
*   [Go](https://go.dev/doc/install) (version 1.20+ recommended)

### Installation
1. Clone the repository to your local machine.
2. Install the required Go packages:
   ```bash
   go get github.com/gin-gonic/gin
   go get gorm.io/gorm
   go get gorm.io/driver/sqlite
   go get github.com/teris-io/shortid
   go get github.com/gin-contrib/sessions
   go get github.com/gin-contrib/sessions/gorm
   go get golang.org/x/crypto/bcrypt
   ```

### Running the App
Run the application by compiling all files in the root folder:
```bash
go run .
```

The server will automatically:
1. Connect to or create a SQLite database file named `pizza.db`.
2. Perform GORM migrations to create SQL tables.
3. Seed a default administrative user account:
   *   **Username**: `admin`
   *   **Password**: `password123`
4. Start listening on **`http://localhost:8080`**.

---

## 🔒 Security & Concurrency Details

*   **Mutex Thread-Safety**: The SSE broadcaster (`NotificationManager`) uses a `sync.RWMutex` to safely manage client maps across concurrent HTTP goroutines, preventing memory data races.
*   **One-Way Hashing**: Admin credentials use `bcrypt` at a work factor cost of 14, rendering password hashes stored in `pizza.db` computationally secure against reverse-engineering.
*   **Database Sessions**: User session cookies are stored directly in the database (`gormsessions`) rather than local cookie state, ensuring cryptographically secure identity verification.
