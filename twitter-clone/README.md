# Twitter Clone Backend (X Clone)

A modern Twitter clone backend implemented in Go following Clean Architecture principles and OOP concepts.

## 🏗️ Architecture

This project follows **Clean Architecture** with clear separation of concerns:

```
twitter-clone/
├── cmd/server/main.go              # Application entry point
├── internal/
│   ├── domain/                     # Domain layer (Entities, DTOs, Business Rules)
│   │   ├── user.go                 # User entities (Account, User, NormalUser, BlueUser, GoldUser, Admin)
│   │   ├── post.go                 # Post entities (Post, File, Image, Video, Hashtag, Report)
│   │   └── dto.go                  # Data Transfer Objects & validation
│   ├── usecase/                    # Use Case layer (Business Logic)
│   │   ├── auth_usecase.go         # Authentication & user management
│   │   └── post_usecase.go         # Post operations & interactions
│   ├── interface/
│   │   └── http/                   # Interface layer (HTTP Handlers)
│   │       ├── auth_handler.go     # Auth endpoints
│   │       └── post_handler.go     # Post endpoints
│   └── infrastructure/
│       └── repository/             # Infrastructure layer (Data Access)
│           ├── interfaces.go       # Repository interfaces
│           └── in_memory_repository.go  # In-memory implementation
├── pkg/
│   ├── hash/                       # Password hashing (bcrypt)
│   ├── token/                      # JWT token generation/validation
│   └── middleware/                 # HTTP middleware (auth, CORS)
└── configs/                        # Configuration files
```

## ✨ Features Implemented

### Authentication & User Management
- ✅ User registration with validation (email, phone, password)
- ✅ User login with JWT token generation
- ✅ Profile management
- ✅ Follow/Unfollow users
- ✅ Search users by username
- ✅ View user profiles (limited info)

### Subscription System
- ✅ Normal User (default)
- ✅ Blue Premium User ($9/month)
- ✅ Gold Premium User ($19/month)
- ✅ Token system (1000 tokens = $1)
- ✅ Different post costs based on user type:
  - Normal: text length + 10 for media
  - Blue: half text length + 5 for media
  - Gold: fixed 5 tokens per post

### Posts & Content
- ✅ Create posts with text, images, videos
- ✅ Automatic hashtag extraction from text
- ✅ Like/Unlike posts
- ✅ Reply to posts (threads)
- ✅ Delete posts
- ✅ Edit posts (premium users only)
- ✅ Search posts by text
- ✅ Get popular posts (sorted by likes)
- ✅ Sort posts (by likes, views, date)
- ✅ Filter posts by date range
- ✅ Generate share links
- ✅ Post recommendations based on interests

### Moderation
- ✅ Report posts/users
- ✅ Block/unblock users (admin)
- ✅ Lock/unlock posts (admin)
- ✅ Manage reports (admin)

### Advanced Features
- ✅ JWT authentication
- ✅ Password hashing with bcrypt
- ✅ CORS support
- ✅ In-memory database (easy to replace with real DB)
- ✅ Thread/reply system (up to 10,000 replies per post)
- ✅ View count tracking
- ✅ Recommendation algorithm

## 🚀 Quick Start

### Prerequisites
- Go 1.19 or higher

### Installation

```bash
cd twitter-clone
go mod tidy
go build ./cmd/server
```

### Running the Server

```bash
./server
```

The server will start on `http://localhost:8080`

## 📡 API Endpoints

### Authentication

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/api/register` | Register new user | No |
| POST | `/api/login` | Login user | No |

### User Operations

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/api/user/profile?username=` | Get user profile | Yes |
| POST | `/api/user/follow` | Follow a user | Yes |
| POST | `/api/user/unfollow` | Unfollow a user | Yes |
| POST | `/api/user/upgrade` | Upgrade subscription | Yes |
| POST | `/api/user/tokens` | Add tokens | Yes |
| GET | `/api/users/search?q=` | Search users | Yes |

### Post Operations

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/api/posts` | Create post | Yes |
| GET | `/api/posts/all` | Get all posts | No |
| GET | `/api/posts/popular?limit=` | Get popular posts | No |
| GET | `/api/posts/user?user_id=` | Get user's posts | No |
| GET | `/api/posts/get?id=` | Get single post | No |
| POST | `/api/posts/like` | Like/unlike post | Yes |
| DELETE | `/api/posts/delete?id=` | Delete post | Yes |
| PUT | `/api/posts/edit?id=` | Edit post | Yes |
| POST | `/api/posts/reply?post_id=` | Reply to post | Yes |
| GET | `/api/posts/search?q=` | Search posts | No |
| GET | `/api/posts/replies?post_id=` | Get replies | No |
| GET | `/api/posts/share?post_id=` | Generate share link | No |
| GET | `/api/posts/recommend?limit=` | Get recommendations | Yes |
| POST | `/api/posts/sort?sort_by=` | Sort posts | No |
| POST | `/api/posts/report` | Report post | Yes |

## 🧪 Testing

### Register a new user

```bash
curl -X POST http://localhost:8080/api/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "Test1234",
    "full_name": "Test User",
    "birth_date": "2000-01-01T00:00:00Z",
    "email": "test@example.com",
    "phone": "+1234567890"
  }'
```

### Login

```bash
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "Test1234"
  }'
```

### Create a post

```bash
curl -X POST http://localhost:8080/api/posts \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "text": "Hello Twitter Clone! #golang #backend",
    "media_type": "",
    "media_path": ""
  }'
```

### Add tokens

```bash
curl -X POST http://localhost:8080/api/user/tokens \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{"amount": 1000}'
```

### Get popular posts

```bash
curl -X GET http://localhost:8080/api/posts/popular?limit=10
```

## 🎯 OOP Principles Applied

### Encapsulation
- Private fields with public methods
- Business logic encapsulated in use cases
- Data access abstracted through repositories

### Inheritance
- `Account` → `User` → `NormalUser`/`PremiumUser`
- `PremiumUser` → `BlueUser`/`GoldUser`
- `File` → `Image`/`Video`

### Polymorphism
- Different `CalculatePostCost()` implementations for each user type
- Repository interfaces with multiple implementations
- Handler methods working with interface types

### Singleton Pattern
- `Admin` instance
- `Database` instance

## 🔄 Clean Architecture Layers

1. **Domain Layer**: Core business entities and rules
2. **Use Case Layer**: Application business logic
3. **Interface Layer**: HTTP handlers and request/response handling
4. **Infrastructure Layer**: Data persistence and external services

## 📝 Notes

- This implementation uses an in-memory database for simplicity
- Replace `InMemoryDatabase` with a real database (PostgreSQL, MongoDB, etc.) for production
- JWT secret key should be stored in environment variables
- Password validation can be enhanced with more complex rules
- The recommendation algorithm is basic and can be improved with ML

## 🛠️ Technologies Used

- **Go 1.19+**: Programming language
- **bcrypt**: Password hashing
- **JWT**: Authentication tokens
- **net/http**: HTTP server
- **Clean Architecture**: Software design pattern

## 📄 License

This project is for educational purposes.
