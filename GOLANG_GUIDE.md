# Go Language Guide & Best Practices

This guide covers Go language concepts ranging from beginner topics to advanced architectural patterns. It builds upon the structure of the Go Fast Start Template and introduces core language features, idioms, and standard practices that every Go developer should know.

---

## 1. Fundamentals

### 1.1 Package and Naming Rules

- One folder usually maps to one package.
- `package main` + `func main()` creates an executable program.
- Export rule is **naming-based**:
  - **Uppercase first letter**: exported / public (`NewService`, `HTTPHandler`)
  - **Lowercase first letter**: package-private (`newService`, `repository`)

### 1.2 Variable Declaration (`:=` vs `var` vs `=`)

In Go, there are multiple ways to declare variables.

Short declaration (declare + infer type + assign):
```go
cfg := config.FromEnv()
```

Equivalent long form:
```go
var cfg config.Config
cfg = config.FromEnv()
```

Or inferred type using `var`:
```go
var cfg = config.FromEnv()
```

**Rules:**
- `:=` can only be used inside functions.
- `=` assigns a value to an existing variable.
- At least one variable on the left side of `:=` must be new.

### 1.3 Control Structures

Go is intentionally minimalist and has only one looping construct: `for`.

**For Loops:**
```go
// Standard C-like for loop
for i := 0; i < 10; i++ {
    // ...
}

// "While" loop equivalent
for condition {
    // ...
}

// Infinite loop
for {
    // ...
}

// Iterating over slices or maps
for index, value := range mySlice {
    // ...
}
```

**If / Else statements:**
Go allows an initialization statement before the condition:
```go
if err := doSomething(); err != nil {
    return err // Handle the error right away
}
```

**Switch statement:**
By default, switch cases do **not** fall through. You don't need `break`.
```go
switch status {
case "active":
    // Do active logic
case "inactive", "suspended":
    // Do inactive logic
default:
    // Handle unknown
}
```

---

## 2. Types and Data Structures

### 2.1 Structs and Methods

Go is not a traditional class-based object-oriented language. It uses `struct`s to define state, and methods attached to types for behavior.

```go
type User struct {
    ID    int
    Name  string
    Email string
}

// Method with a value receiver - cannot modify the original struct fields
func (u User) GetDisplayName() string {
    return u.Name + " (" + u.Email + ")"
}

// Method with a pointer receiver - CAN modify the struct fields
func (u *User) UpdateEmail(newEmail string) {
    u.Email = newEmail
}
```

**Pointer vs Value Receivers:**
- Use a **pointer receiver** when a method needs to mutate state or to avoid copying large structs (performance).
- Use a **value receiver** for small immutable-like data or when mutation is not intended.

### 2.2 Arrays, Slices, and Maps

Before understanding slices, it helps to understand arrays.

**Arrays:**
An array has a fixed length defined at compile time. Because of their fixed size, they are less commonly used directly in Go compared to slices.

```go
// Declaring a fixed-size array of 3 strings
var names [3]string
names[0] = "Alice"
names[1] = "Bob"
names[2] = "Charlie"

// Short declaration and initialization
colors := [2]string{"red", "blue"}
```

**Slices (Preferred):**
Slices are dynamically-sized, flexible views into the elements of an array. In almost all Go code, you'll use slices rather than arrays because you can easily append new elements to them as your data grows.

```go
// Creating a slice (notice there is no number in the brackets `[]`)
users := make([]User, 0, 10) // Type, initial length, optionally capacity

// Appending dynamically (which you cannot do with a fixed array)
users = append(users, User{Name: "Alice"})
```
*Note: Appending to a slice might allocate a new underlying array if capacity is reached, so always assign the result back to the slice variable.*

**Maps:**
A map is an unordered group of elements of one type, indexed by a set of unique keys of another type.

```go
// Creating a map
userCache := make(map[string]User)

// Adding to map
userCache["alice@example.com"] = User{Name: "Alice"}

// Retrieving and checking existence safely (the "comma ok" idiom)
if user, exists := userCache["bob@example.com"]; exists {
    fmt.Println(user.Name)
}
```

---

## 3. Interfaces and Contracts

### 3.1 Interfaces as Behavior Contracts

Interfaces in Go define a set of methods. If a type provides those methods, it implements the interface. There is no `implements` keyword.

- **Prefer small, focused interfaces** (often just one or two methods).
- Define the interface near the **consumer**, not the producer.
- In clean architecture (like `handler -> service -> repository`), each layer depends on interfaces, decoupling it from concrete implementations. This significantly improves testability via mocking.

```go
// Consumer defines what it needs
type UserRepository interface {
    GetByID(ctx context.Context, id int) (*User, error)
}
```

### 3.2 Implicit Implementation

Because interfaces are implemented implicitly, you can compose existing types easily.

```go
// Any type that implements `Read(p []byte) (n int, err error)` is an `io.Reader`.
type MyReader struct{}

func (m MyReader) Read(p []byte) (int, error) {
    return 0, io.EOF
}
// MyReader is now an io.Reader anywhere io.Reader is expected!
```

---

## 4. Error Handling

### 4.1 The Error-First Style

Go does not use exceptions for normal control flow. It relies on explicit error checking.

```go
dbPool, err := database.NewPostgresPool(rootCtx, cfg.PostgresDSN())
if err != nil {
    log.Fatalf("failed to initialize postgres pool: %v", err)
}
```

- Return errors as the last value from a function.
- Handle errors early (exit early pattern) to keep the "happy path" unindented.
- Avoid nesting `if else` excessively.

### 4.2 Advanced Errors (`errors.Is` and `errors.As`)

Since Go 1.13, you can unwrap nested errors and accurately check against sentinel (pre-defined) errors.

```go
var ErrNotFound = errors.New("record not found")

// Wrapping errors provides context
func findUser() error {
    return fmt.Errorf("user search failed: %w", ErrNotFound)
}

// Checking for specific errors
err := findUser()
if errors.Is(err, ErrNotFound) {
    // We specifically hit a not found scenario
}
```

---

## 5. Concurrency Basics

### 5.1 Goroutines and Channels

Goroutines are lightweight threads managed by the Go runtime.
Channels provide typed communication between goroutines.

```go
serverErr := make(chan error, 1) // Buffered channel
go func() { // Start a concurrent goroutine
    serverErr <- httpServer.ListenAndServe() // Send error to channel
}()
```

### 5.2 WaitGroup and MutEX

When managing multiple tasks, use `sync.WaitGroup` to wait for a collection of goroutines to finish.
Use `sync.Mutex` to safely access shared data across multiple goroutines.

```go
var wg sync.WaitGroup
var count int
var mu sync.Mutex

for i := 0; i < 10; i++ {
    wg.Add(1) // Tell the WaitGroup we added one task
    go func() {
        defer wg.Done() // Announce task completion when returning

        mu.Lock()
        count++ // Mutex ensures this is thread-safe
        mu.Unlock()
    }()
}

wg.Wait() // Block here until all 10 tasks report Done
```

### 5.3 Select Statement

The `select` statement lets a goroutine wait on multiple communication operations. It's like a `switch` for channels.

```go
select {
case msg := <-messageChan:
    fmt.Println("Received:", msg)
case <-time.After(5 * time.Second):
    fmt.Println("Timeout reached!")
case <-ctx.Done():
    fmt.Println("Context cancelled")
}
```

---

## 6. Architecture & Ecosystem

### 6.1 Why `internal/` exists

- `internal` is a special directory recognized by the Go compiler.
- It acts as an **import boundary**; code inside `internal` cannot be imported by external projects.
- Use it to protect your project's implementation details from becoming a public, breaking API constraint.

### 6.2 Dependency Wiring in `main`

- `main.go` acts as your **composition root**.
- Create config, database pools, repositories, services, and handlers explicitly.
- **Avoid global variables** (e.g., global DB connections). By having explicit dependencies passed explicitly (Dependency Injection), you make the code much easier to reason about and test.

### 6.3 Context Propagation

- `context.Context` carries deadlines, cancellation signals, and request-scoped values across API boundaries.
- Pass `ctx` as the **first argument** to functions that perform I/O (DB queries, network calls, etc.).
- When a server shuts down or a client aborts a request, the context is cancelled, efficiently aborting the downstream work.

### 6.4 Graceful Shutdown Pattern

- Use `signal.NotifyContext` to capture `SIGINT`/`SIGTERM` (e.g., Ctrl+C or Kubernetes termination).
- On signal, stop accepting new HTTP requests and let active requests finish cleanly within a timeout using `httpServer.Shutdown(ctx)`.
- It ensures you don't abruptly drop active user connections during deployments.

### 6.5 Go Modules and Private Dependencies

- Go manages dependencies using `go.mod`.
- If your company uses private repos, the standard `go get` will fail trying to verify checksums on public sumdb servers.
- Configure `GOPRIVATE` to tell Go which domains hold private code:
```bash
go env -w GOPRIVATE=github.com/your-org/*
```

---

## 7. Testing Mindset

### 7.1 Table-Driven Tests

- Keep tests deterministic, small, and behavior-oriented (test paths, inputs, and outputs, not implementation mechanics).
- Use **table-driven tests** to neatly define multiple test scenarios in a struct slice.

```go
func TestSomething(t *testing.T) {
    tests := []struct {
        name string
        in   string
        want string
    }{
        {name: "normal uppercase", in: "a", want: "A"}, // Test cases defined here
        {name: "already uppercase", in: "A", want: "A"},
    }

    for _, tc := range tests {
        // Evaluate each case independently
        t.Run(tc.name, func(t *testing.T) {
            got := strings.ToUpper(tc.in)
            if got != tc.want {
                t.Errorf("got %q want %q", got, tc.want) // t.Errorf reports failure but doesn't abort
            }
        })
    }
}
```
