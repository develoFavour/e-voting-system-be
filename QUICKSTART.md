# eVote Backend - Quick Start Guide

## Prerequisites

1. **Go** (1.21 or higher)
2. **MongoDB** (running locally or cloud instance)

## Installation

### 1. Install Dependencies

```bash
cd backend

# Install all required packages
go get github.com/gin-gonic/gin
go get github.com/gin-contrib/cors
go get go.mongodb.org/mongo-driver/mongo
go get github.com/golang-jwt/jwt/v5
go get github.com/joho/godotenv
go get golang.org/x/crypto/bcrypt
```

### 2. Setup Environment

```bash
# Copy example env file
cp .env.example .env

# Edit .env with your settings
# IMPORTANT: Change JWT_SECRET and ENCRYPTION_KEY in production!
```

### 3. Start MongoDB

**Option A: Local MongoDB**
```bash
# Windows (if installed as service)
net start MongoDB

# Or run manually
mongod --dbpath C:\data\db
```

**Option B: MongoDB Atlas (Cloud)**
- Create free cluster at mongodb.com/cloud/atlas
- Get connection string
- Update MONGODB_URI in .env

### 4. Run the Server

```bash
go run cmd/server/main.go
```

Server will start on `http://localhost:8080`

## Creating Admin User

Since there's no admin user by default, you need to create one manually in MongoDB:

```javascript
// Connect to MongoDB
use evote

// Create admin user
db.users.insertOne({
  matric_number: "ADMIN001",
  full_name: "Admin User",
  department: "IT",
  faculty: "Administration",
  password_hash: "$2a$14$...", // Use bcrypt to hash "admin123"
  status: "APPROVED",
  role: "ADMIN",
  has_voted: false,
  created_at: new Date(),
  updated_at: new Date()
})
```

**Or use this Go script to create admin:**

```go
package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	password := "admin123"
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), 14)
	fmt.Println(string(hash))
}
```

## Testing the API

### 1. Health Check
```bash
curl http://localhost:8080/health
```

### 2. Student Registration
```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "matricNumber": "HU/2024/001",
    "fullName": "John Doe",
    "department": "Computer Science",
    "faculty": "Natural Sciences",
    "password": "password123",
    "idCardUrl": "https://example.com/id.jpg"
  }'
```

### 3. Admin Login
```bash
curl -X POST http://localhost:8080/api/auth/admin/login \
  -H "Content-Type: application/json" \
  -d '{
    "matricNumber": "ADMIN001",
    "password": "admin123"
  }'
```

### 4. Approve Voter (Admin)
```bash
curl -X PUT http://localhost:8080/api/admin/accreditation/{user_id}/approve \
  -H "Authorization: Bearer {admin_token}"
```

### 5. Add Candidate (Admin)
```bash
curl -X POST http://localhost:8080/api/admin/candidates \
  -H "Authorization: Bearer {admin_token}" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Jane Smith",
    "position": "President",
    "party": "Progressive Party",
    "manifesto": "I will improve student welfare...",
    "imageUrl": "https://example.com/candidate.jpg"
  }'
```

### 6. Cast Vote (Student)
```bash
curl -X POST http://localhost:8080/api/vote/cast \
  -H "Authorization: Bearer {student_token}" \
  -H "Content-Type: application/json" \
  -d '{
    "selections": {
      "President": "candidate_id_here",
      "Vice President": "candidate_id_here"
    }
  }'
```

## Project Structure

```
backend/
├── cmd/server/main.go          # Entry point
├── internal/
│   ├── config/                 # Configuration
│   ├── database/               # MongoDB connection
│   ├── handlers/               # HTTP handlers
│   ├── middleware/             # Auth, CORS, Rate limit
│   ├── models/                 # Data models
│   ├── repository/             # Database operations
│   └── services/               # Business logic
└── pkg/utils/                  # Utilities
```

## Common Issues

### MongoDB Connection Error
- Ensure MongoDB is running
- Check MONGODB_URI in .env
- Verify network connectivity

### JWT Token Invalid
- Check JWT_SECRET matches in .env
- Ensure token is sent as "Bearer {token}"

### Double Voting Error
- This is expected! The atomic operation prevents it
- User can only vote once

## Next Steps

1. ✅ Backend is complete
2. Connect frontend to backend
3. Test full flow: Register → Approve → Vote
4. Deploy to production

## Production Checklist

- [ ] Change JWT_SECRET to strong random string
- [ ] Change ENCRYPTION_KEY to 32-byte random string
- [ ] Use MongoDB Atlas or production database
- [ ] Enable HTTPS
- [ ] Set ENVIRONMENT=production
- [ ] Configure proper CORS origins
- [ ] Set up monitoring/logging
