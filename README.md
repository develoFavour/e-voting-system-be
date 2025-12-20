# eVote Backend - Go + Gin + MongoDB

## Project Structure

```
backend/
├── cmd/
│   └── server/
│       └── main.go           # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go         # Configuration management
│   ├── models/
│   │   ├── user.go           # User model
│   │   ├── candidate.go      # Candidate model
│   │   └── vote.go           # Vote model
│   ├── handlers/
│   │   ├── auth.go           # Authentication handlers
│   │   ├── user.go           # User management handlers
│   │   ├── vote.go           # Voting handlers
│   │   └── admin.go          # Admin handlers
│   ├── middleware/
│   │   ├── auth.go           # JWT authentication
│   │   ├── cors.go           # CORS configuration
│   │   └── ratelimit.go      # Rate limiting
│   ├── repository/
│   │   ├── user.go           # User database operations
│   │   ├── candidate.go      # Candidate database operations
│   │   └── vote.go           # Vote database operations
│   ├── services/
│   │   ├── auth.go           # Authentication business logic
│   │   ├── vote.go           # Voting business logic
│   │   └── encryption.go     # Vote encryption/decryption
│   └── database/
│       └── mongodb.go        # MongoDB connection
├── pkg/
│   └── utils/
│       ├── response.go       # Standard API responses
│       └── validator.go      # Input validation
├── .env.example
├── .gitignore
├── go.mod
├── go.sum
└── README.md
```

## Tech Stack

- **Framework**: Gin Gonic
- **Database**: MongoDB
- **Authentication**: JWT
- **Encryption**: AES-256 for vote data
- **Password Hashing**: bcrypt

## Environment Variables

```env
PORT=8080
MONGODB_URI=mongodb://localhost:27017
DB_NAME=evote
JWT_SECRET=your-super-secret-key
ENCRYPTION_KEY=your-32-byte-encryption-key
FRONTEND_URL=http://localhost:3000
```

## API Endpoints

### Authentication
- `POST /api/auth/register` - Student accreditation registration
- `POST /api/auth/login` - Student login
- `POST /api/auth/admin/login` - Admin login

### User Management
- `GET /api/users/me` - Get current user profile
- `GET /api/users/status` - Check accreditation status

### Admin
- `GET /api/admin/accreditation/pending` - Get pending accreditation requests
- `PUT /api/admin/accreditation/:id/approve` - Approve voter
- `PUT /api/admin/accreditation/:id/reject` - Reject voter
- `POST /api/admin/candidates` - Add candidate
- `GET /api/admin/results` - Get election results

### Voting
- `GET /api/vote/candidates` - Get all candidates
- `POST /api/vote/cast` - Cast vote (atomic operation)
- `GET /api/vote/results` - Get live results

## Security Features

1. **JWT Authentication**: Stateless token-based auth
2. **Password Hashing**: bcrypt with salt
3. **Vote Encryption**: AES-256 encryption for vote data
4. **Atomic Voting**: MongoDB transactions to prevent double voting
5. **Rate Limiting**: Prevent brute force attacks
6. **CORS**: Strict origin control

## Installation

```bash
# Install dependencies
go mod download

# Run the server
go run cmd/server/main.go
```

## Development

```bash
# Hot reload (install air first)
go install github.com/cosmtrek/air@latest
air
```
