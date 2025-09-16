# Caslette Casino Project

A complete casino platform with user management, authentication, roles, permissions, and diamond currency system.

## Project Structure

```
caslette/
├── caslette-server/     # Go backend API
├── caslette-web/        # React user frontend
└── castelle-web-admin/  # React admin dashboard
```

## Backend (caslette-server)

### Tech Stack

- **Go** with Gin web framework
- **GORM** for database ORM
- **MySQL** database (schema: 'castelle')
- **JWT** authentication
- **bcrypt** password hashing

### Features

- User authentication and registration
- Role-based access control (admin, moderator, user)
- Permission system
- Diamond currency system with transaction tracking
- RESTful API endpoints

### Setup

1. Navigate to `caslette-server`
2. Ensure MySQL is running with a database named 'castelle'
3. Update `.env` file with your database credentials
4. Run the server:
   ```bash
   go run main.go
   ```

### API Endpoints

- **Auth**: `/api/v1/auth/login`, `/api/v1/auth/register`, `/api/v1/auth/profile`
- **Users**: `/api/v1/users` (CRUD operations)
- **Diamonds**: `/api/v1/diamonds/user/:userId`, `/api/v1/diamonds/credit`, `/api/v1/diamonds/debit`

### Default Database Setup

The system automatically creates:

- **Roles**: admin, moderator, user
- **Permissions**: CRUD operations for users, roles, diamonds
- **Default diamond balance**: 1000 diamonds for new users

## Frontend User App (caslette-web)

### Tech Stack

- **React 18** with TypeScript
- **Vite** for build tooling
- **Tailwind CSS** for styling
- **React Router** for navigation
- **Axios** for API calls

### Features

- User login and registration
- Dashboard with profile information
- Diamond balance display
- Transaction history
- Responsive design

### Setup

1. Navigate to `caslette-web`
2. Install dependencies: `npm install`
3. Start development server: `npm run dev`

## Admin Dashboard (castelle-web-admin)

### Tech Stack

- **React 18** with TypeScript
- **Vite** for build tooling
- **Tailwind CSS** for styling
- **React Router** for navigation
- **Axios** for API calls

### Features

- Admin-only authentication
- User management (view, activate/deactivate users)
- Diamond transaction management (credit/debit diamonds)
- Real-time transaction history
- Role-based access control

### Setup

1. Navigate to `castelle-web-admin`
2. Install dependencies: `npm install`
3. Start development server: `npm run dev`

## Getting Started

### Prerequisites

- Go 1.24+
- Node.js 18+
- MySQL 8.0+

### Quick Start

1. **Setup Database**:

   - Create MySQL database named 'castelle'
   - User: root, Password: (empty) - or update `.env` file

2. **Start Backend**:

   ```bash
   cd caslette-server
   go run main.go
   ```

3. **Start User Frontend**:

   ```bash
   cd caslette-web
   npm install
   npm run dev
   ```

4. **Start Admin Dashboard**:
   ```bash
   cd castelle-web-admin
   npm install
   npm run dev
   ```

### Default Access

- **User App**: http://localhost:5173
- **Admin Dashboard**: http://localhost:5174
- **API Server**: http://localhost:8080

### Creating Admin User

1. Register a user through the regular interface
2. Manually assign admin/moderator role in the database, or
3. Use the database seeding to create default admin users

## Database Schema

### Users

- Basic user information
- Password hashing with bcrypt
- Active/inactive status
- Many-to-many relationship with roles

### Roles & Permissions

- Hierarchical role system
- Fine-grained permissions
- Resource-based access control

### Diamonds

- Transaction-based currency system
- Running balance tracking
- Transaction types: credit, debit, bonus, purchase
- Audit trail with transaction IDs

## Security Features

- JWT-based authentication
- Password hashing with bcrypt
- Role-based access control
- Protected API routes
- CORS configuration
- Admin privilege validation

## Development Notes

- Backend runs on port 8080
- Frontend apps run on ports 5173 and 5174
- All API calls use JWT tokens
- Admin dashboard requires admin/moderator roles
- New users start with 1000 diamonds
