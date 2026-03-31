# API Testing Guide

This guide provides a step-by-step walkthrough for testing the complete authentication and user management lifecycle in this Go template. It covers registering a user, obtaining an authentication token, and accessing protected endpoints.

---

## Prerequisites

Before testing, ensure your application dependencies (PostgreSQL and Redis) and the Go server are running:

1. **Start the database and cache using Docker:**
   ```bash
   docker compose up -d
   ```
2. **Start the Go server:**
   ```bash
   go run ./cmd/api
   ```

By default, the server listens on `http://localhost:8080`.

---

## Step 1: Register a New User

First, we need to create a new user account. This endpoint is **Public** and does not require any authorization. 

**Endpoint:** `POST /api/v1/auth/register`

**Request (`curl`):**
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "johndoe",
    "name": "John Doe",
    "email": "john@example.com",
    "password": "supersecretpassword",
    "gender": "male"
  }'
```

**Expected Response (201 Created):**
```json
{
  "data": {
    "id": 2,
    "username": "johndoe",
    "name": "John Doe",
    "email": "john@example.com",
    "gender": "male",
    "created_at": "2026-03-31T15:00:00Z",
    "updated_at": "2026-03-31T15:00:00Z"
  }
}
```
*(Note: The system securely hashes the password via bcrypt and will never return it in responses).*

---

## Step 2: Login to Obtain JWT

Now that you have an account, securely login using those credentials to obtain your JSON Web Token (JWT). This endpoint is **Public**.

**Endpoint:** `POST /api/v1/auth/login`

**Request (`curl`):**
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "supersecretpassword"
  }'
```

**Expected Response (200 OK):**
```json
{
  "data": {
    "id": 2,
    "username": "johndoe",
    "name": "John Doe",
    "email": "john@example.com",
    "gender": "male",
    "created_at": "2026-03-31T15:00:00Z",
    "updated_at": "2026-03-31T15:00:00Z"
  },
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**⚠️ IMPORTANT:** Copy the `token` string from the response. You will need to attach it as a `Bearer` token to the `Authorization` header for all subsequent protected requests.

---

## Step 3: Get Your Profile (Get Me)

This endpoint retrieves the currently authenticated user's profile. 
- *Architecture Note:* This call checks **Redis** first. If a cached profile exists, it serves it instantly. Otherwise, it queries PostgreSQL and seeds the Redis cache with a 5-minute TTL.

**Endpoint:** `GET /api/v1/users/me`
**Requirement:** `Authorization: Bearer <token>`

**Request (`curl`):**
```bash
curl -X GET http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Expected Response (200 OK):**
```json
{
  "data": {
    "id": 2,
    "username": "johndoe",
    "name": "John Doe",
    "email": "john@example.com",
    "gender": "male",
    "created_at": "2026-03-31T15:00:00Z",
    "updated_at": "2026-03-31T15:00:00Z"
  }
}
```

---

## Step 4: List All Users

Finally, try accessing another protected route to list all users in the system. The default application setup also automatically seeds an "admin" user account.

**Endpoint:** `GET /api/v1/users`
**Requirement:** `Authorization: Bearer <token>`

**Request (`curl`):**
```bash
# Add ?limit=10&offset=0 to strictly paginate results
curl -X GET "http://localhost:8080/api/v1/users?limit=10" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Expected Response (200 OK):**
```json
{
  "data": [
    {
      "id": 1,
      "username": "admin",
      "name": "Admin",
      "email": "admin@email.com",
      "gender": "unspecified",
      "created_at": "2026-03-31T14:55:00Z",
      "updated_at": "2026-03-31T14:55:00Z"
    },
    {
      "id": 2,
      "username": "johndoe",
      "name": "John Doe",
      "email": "john@example.com",
      "gender": "male",
      "created_at": "2026-03-31T15:00:00Z",
      "updated_at": "2026-03-31T15:00:00Z"
    }
  ]
}
```

---

## 🛑 Unauthorized Error Handling

If you attempt to call a protected endpoint (like `/api/v1/users` or `/api/v1/users/me`) **without** attaching the authorization header, or if you use an expired token, the global JWT Middleware will intercept the request and immediately reject it.

**Request (Missing Token):**
```bash
curl -v -X GET http://localhost:8080/api/v1/users/me
```

**Expected Rejection (401 Unauthorized):**
```json
{
  "error": "missing authorization header"
}
```
