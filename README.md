# MOH SSO Dashboard

## 1. Introduction

The **MOH SSO Dashboard** is a centralized authentication,
authorization, and client management system built for Ministry of Health
(MOH) applications.\
It integrates all internal systems under a unified Single Sign-On (SSO)
platform powered by **Keycloak**, enabling users to access multiple apps
through one login.

This documentation provides a **fully detailed and production-ready**
overview of the project.

------------------------------------------------------------------------

## 2. System Architecture Overview

The system is composed of three major parts:

### **2.1 Frontend (React + Carbon)**

-   React 19 with TypeScript
-   Carbon Design System UI components
-   Uses AuthContext for token refresh & state management
-   Displays applications as tiles based on user role
-   Fetches user details from backend using JWT

### **2.2 Backend (Go + Gin)**

The backend provides: - Auth endpoints (`/auth/me`, `/auth/logout`) -
Client management (`/clients`) - User management (`/users`) -
SQLC-generated DB queries - Keycloak Admin API integrations

Uses: - Gin Web Framework - PostgreSQL - SQLC for strict type-safe
database access - Zerolog for logging

### **2.3 Keycloak (Identity Provider)**

Handles: - Authentication (OIDC) - User roles - Service accounts - Realm
configuration - Token issuance & validation

------------------------------------------------------------------------

## 3. Features in Detail

### **3.1 Authentication & Authorization**

-   Full login via Keycloak Authorization Code Flow
-   Backend extracts and validates tokens using middleware
-   Defines two main roles: `admin` and `user`
-   Admin can manage system users & clients
-   Users only view applications they are assigned to

------------------------------------------------------------------------

### **3.2 Client Management**

The dashboard allows administrators to manage applications connected
through SSO.

Features include: - Create client in Keycloak + local DB sync - Update
client information and roles - Delete client (with safety checks) -
Assign icon, baseURL, and visibility - Frontend displays apps
dynamically

Client model contains:

    ID
    ClientID
    Name
    Description
    BaseURL
    Icon
    PublicClient
    Enabled
    Attributes (extensible)

------------------------------------------------------------------------

### **3.3 User Management**

Admins can: - Create Keycloak + local DB users - Assign user roles
(`admin`, `user`) - Enable/disable users - Delete or update user profile
data

------------------------------------------------------------------------

## 4. API Reference (Backend)

### **4.1 Authentication API**

#### `GET /api/v1/auth/me`

Returns user profile from Keycloak token.

#### `POST /api/v1/auth/logout`

Logs the user out of Keycloak.

------------------------------------------------------------------------

### **4.2 Client API**

#### `GET /api/v1/clients`

List all clients visible to the active user.

#### `POST /api/v1/clients`

Create a new client.

#### `PUT /api/v1/clients/:id`

Update an existing client.

#### `DELETE /api/v1/clients/:id`

Remove a client from system.

------------------------------------------------------------------------

### **4.3 User API**

#### `POST /api/v1/users`

Create user in Keycloak & database.

#### `GET /api/v1/users`

List all users.

#### `GET /api/v1/users/:id`

Fetch user by ID.

#### `DELETE /api/v1/users/:id`

Delete user from system.

------------------------------------------------------------------------

## 5. Development Setup

### **5.1 Clone Repository**

    git clone https://github.com/moh-sso-dashboard.git
    cd sso-dashboard

### **5.2 Backend Setup**

Install dependencies:

    go mod tidy

Run:

    go run ./cmd/server

Optional auto-reload:

    air

### **5.3 Frontend Setup**

    npm install
    npm run dev

------------------------------------------------------------------------

## 6. Docker Deployment

### Start everything:

    docker-compose up -d

Includes: - Keycloak - PostgreSQL - Backend - Frontend

### Production build:

    docker build -t moh-sso-backend ./backend
    docker build -t moh-sso-frontend ./frontend

------------------------------------------------------------------------

## 7. Environment Variables

### Backend `.env`

    DB_SOURCE=postgresql://postgres:postgres@db:5432/sso?sslmode=disable
    KEYCLOAK_BASE_URL=http://keycloak:8080
    KEYCLOAK_REALM=moh-realm
    KEYCLOAK_CLIENT_ID=dashboard
    KEYCLOAK_CLIENT_SECRET=changeme
    SERVER_PORT=9000

### Frontend `.env`

    VITE_API_BASE=http://localhost:9000/api/v1
    VITE_KEYCLOAK_URL=http://localhost:8081
    VITE_KEYCLOAK_REALM=moh-realm
    VITE_KEYCLOAK_CLIENT_ID=dashboard

------------------------------------------------------------------------

## 8. Database Schema Overview

Tables include: - `user` - `client` - `client_roles` - `activity_logs` -
`audit_trails`

SQLC code is generated into `internal/db/sqlc`.

------------------------------------------------------------------------

## 9. Keycloak Realm Requirements

Your realm must include: - Realm: **moh-realm** - Roles: - admin -
user - Dashboard Client: - clientId: dashboard - serviceAccountsEnabled:
true - redirectUris: "http://localhost:3000/*" - webOrigins: "*"

------------------------------------------------------------------------

## 10. Future Enhancements

-   Multi-tenant support
-   App analytics dashboard
-   Audit log UI
-   Invite-based user onboarding
-   Role-based menu customization

------------------------------------------------------------------------

## 11. License

MIT License
