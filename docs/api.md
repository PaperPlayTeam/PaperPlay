# PaperPlay Backend API Documentation

## Overview

PaperPlay is a gamified learning platform based on academic papers. This document describes the REST API endpoints and WebSocket interfaces for the backend system.

**Base URL**: `http://localhost:8080`  
**API Version**: v1  
**Authentication**: JWT Bearer Token  
**Content-Type**: `application/json`

## Authentication

### Register User

Register a new user account.

**Endpoint**: `POST /api/v1/auth/register`

**Request Body**:
```json
{
  "email": "user@example.com",
  "password": "password123",
  "display_name": "John Doe"
}
```

**Response** (201 Created):
```json
{
  "success": true,
  "message": "User registered successfully",
  "data": {
    "user": {
      "id": "uuid-string",
      "email": "user@example.com",
      "display_name": "John Doe",
      "eth_address": "0x...",
      "created_at": "2025-01-01T00:00:00Z"
    },
    "access_token": "jwt-token",
    "refresh_token": "refresh-token",
    "expires_in": 900
  }
}
```

### Login User

Authenticate an existing user.

**Endpoint**: `POST /api/v1/auth/login`

**Request Body**:
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Response** (200 OK):
```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "user": {
      "id": "uuid-string",
      "email": "user@example.com",
      "display_name": "John Doe",
      "eth_address": "0x..."
    },
    "access_token": "jwt-token",
    "refresh_token": "refresh-token",
    "expires_in": 900
  }
}
```

### Refresh Token

Refresh an expired access token.

**Endpoint**: `POST /api/v1/auth/refresh`

**Request Body**:
```json
{
  "refresh_token": "refresh-token-string"
}
```

**Response** (200 OK):
```json
{
  "success": true,
  "message": "Token refreshed successfully",
  "data": {
    "access_token": "new-jwt-token",
    "refresh_token": "new-refresh-token",
    "expires_in": 900
  }
}
```

## User Management

All user endpoints require authentication via `Authorization: Bearer <access_token>` header.

### Get User Profile

Get the current user's profile information.

**Endpoint**: `GET /api/v1/users/profile`

**Headers**:
```
Authorization: Bearer <access_token>
```

**Response** (200 OK):
```json
{
  "success": true,
  "data": {
    "user": {
      "id": "uuid-string",
      "email": "user@example.com",
      "display_name": "John Doe",
      "avatar_url": "https://example.com/avatar.jpg",
      "eth_address": "0x...",
      "created_at": "2025-01-01T00:00:00Z",
      "updated_at": "2025-01-01T00:00:00Z"
    },
    "stats": {
      "total_questions_answered": 150,
      "total_correct_answers": 120,
      "overall_correct_rate": 0.8,
      "current_streak_days": 5,
      "longest_streak_days": 12,
      "total_study_time_ms": 3600000,
      "levels_completed": 25,
      "levels_in_progress": 3,
      "achievements_earned": 8,
      "today_questions_answered": 10,
      "today_correct_answers": 8,
      "today_study_time_ms": 300000,
      "weekly_activity_days": 6,
      "favorite_subjects": [
        {
          "subject_id": "uuid",
          "subject_name": "Computer Science",
          "completed": 15,
          "in_progress": 2,
          "correct_rate": 0.85
        }
      ]
    }
  }
}
```

### Update User Profile

Update the current user's profile information.

**Endpoint**: `PUT /api/v1/users/profile`

**Headers**:
```
Authorization: Bearer <access_token>
```

**Request Body**:
```json
{
  "display_name": "Jane Doe",
  "avatar_url": "https://example.com/new-avatar.jpg"
}
```

**Response** (200 OK):
```json
{
  "success": true,
  "message": "Profile updated successfully",
  "data": {
    "user": {
      "id": "uuid-string",
      "email": "user@example.com",
      "display_name": "Jane Doe",
      "avatar_url": "https://example.com/new-avatar.jpg",
      "updated_at": "2025-01-01T12:00:00Z"
    }
  }
}
```

### Get User Progress

Get the user's learning progress across all subjects.

**Endpoint**: `GET /api/v1/users/progress`

**Headers**:
```
Authorization: Bearer <access_token>
```

**Response** (200 OK):
```json
{
  "success": true,
  "data": {
    "subjects": [
      {
        "subject": {
          "id": "uuid",
          "name": "Computer Science",
          "description": "Computer Science related papers"
        },
        "roadmap": [
          {
            "id": "uuid",
            "level_id": "uuid",
            "path": "1",
            "order_index": 1,
            "is_unlocked": true,
            "children": []
          }
        ],
        "progress": [
          {
            "id": "uuid",
            "user_id": "uuid",
            "level_id": "uuid",
            "status": 2,
            "score": 85,
            "stars": 3,
            "last_attempt_at": "2025-01-01T10:00:00Z"
          }
        ],
        "completion_rate": 0.75
      }
    ],
    "recent_activity": [
      {
        "date": "2025-01-01",
        "activity": "level_completed",
        "description": "Completed level: Introduction to ML",
        "level_id": "uuid",
        "subject_id": "uuid"
      }
    ]
  }
}
```

### Get User Achievements

Get the user's earned achievements.

**Endpoint**: `GET /api/v1/users/achievements`

**Headers**:
```
Authorization: Bearer <access_token>
```

**Response** (200 OK):
```json
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "user_id": "uuid",
      "achievement_id": "uuid",
      "earned_at": "2025-01-01T10:00:00Z",
      "event_data": {
        "trigger_event": "level_completed",
        "level_id": "uuid"
      },
      "achievement": {
        "id": "uuid",
        "name": "First Steps",
        "description": "Complete your first level",
        "icon_url": "https://example.com/icon.png",
        "level": 1,
        "is_active": true,
        "nft_metadata": {
          "name": "First Steps NFT",
          "description": "Commemorating your first completed level",
          "image": "https://example.com/nft.png"
        }
      }
    }
  ]
}
```

### Logout User

Invalidate the current user's tokens.

**Endpoint**: `POST /api/v1/users/logout`

**Headers**:
```
Authorization: Bearer <access_token>
```

**Response** (200 OK):
```json
{
  "success": true,
  "message": "Logout successful"
}
```

## Achievement System

### Get All Achievements

Get all available achievements in the system.

**Endpoint**: `GET /api/v1/achievements`

**Headers**:
```
Authorization: Bearer <access_token>
```

**Response** (200 OK):
```json
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "name": "First Steps",
      "description": "Complete your first level",
      "icon_url": "https://example.com/icon.png",
      "level": 1,
      "category": "progression",
      "is_active": true,
      "rules": {
        "type": "first_try",
        "conditions": [
          {
            "field": "levels_completed",
            "operator": ">=",
            "value": 1
          }
        ]
      },
      "nft_metadata": {
        "name": "First Steps NFT",
        "description": "Commemorating your first completed level",
        "image": "https://example.com/nft.png",
        "attributes": [
          {
            "trait_type": "Category",
            "value": "Progression"
          }
        ]
      }
    }
  ]
}
```

### Get User's Achievements

Get achievements earned by the current user.

**Endpoint**: `GET /api/v1/achievements/user`

**Headers**:
```
Authorization: Bearer <access_token>
```

**Response** (200 OK):
```json
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "user_id": "uuid",
      "achievement_id": "uuid",
      "earned_at": "2025-01-01T10:00:00Z",
      "event_data": {
        "trigger_event": "level_completed"
      },
      "achievement": {
        "id": "uuid",
        "name": "First Steps",
        "description": "Complete your first level",
        "level": 1
      }
    }
  ]
}
```

### Trigger Achievement Evaluation

Manually trigger achievement evaluation for the current user (primarily for testing).

**Endpoint**: `POST /api/v1/achievements/evaluate`

**Headers**:
```
Authorization: Bearer <access_token>
```

**Response** (200 OK):
```json
{
  "success": true,
  "message": "Achievement evaluation completed"
}
```

## WebSocket Interface

### WebSocket Connection

Connect to the WebSocket server for real-time notifications.

**Endpoint**: `GET /ws`

**Query Parameters** (optional):
- `token`: JWT access token for authenticated connections

**Connection URL**: `ws://localhost:8080/ws`

### WebSocket Message Types

#### Client to Server Messages

**Ping Message**:
```json
{
  "type": "ping"
}
```

**Subscribe to Channel**:
```json
{
  "type": "subscribe",
  "channel": "achievements"
}
```

#### Server to Client Messages

**Connection Confirmation**:
```json
{
  "type": "connected",
  "data": {
    "status": "connected"
  },
  "timestamp": "2025-01-01T10:00:00Z"
}
```

**Pong Response**:
```json
{
  "type": "pong",
  "data": {
    "status": "ok"
  },
  "timestamp": "2025-01-01T10:00:00Z"
}
```

**Achievement Notification**:
```json
{
  "type": "notification",
  "user_id": "uuid",
  "data": {
    "id": "notification-uuid",
    "type": "achievement",
    "title": "New Achievement!",
    "message": "You earned the 'First Steps' achievement!",
    "icon": "https://example.com/icon.png",
    "achievement": {
      "id": "uuid",
      "name": "First Steps",
      "description": "Complete your first level",
      "level": 1,
      "icon_url": "https://example.com/icon.png"
    }
  },
  "timestamp": "2025-01-01T10:00:00Z"
}
```

**Weekly Report Notification**:
```json
{
  "type": "notification",
  "user_id": "uuid",
  "data": {
    "id": "notification-uuid",
    "type": "weekly_report",
    "title": "Weekly Learning Report",
    "message": "This week you studied 5 days and answered 50 questions correctly!"
  },
  "timestamp": "2025-01-01T10:00:00Z"
}
```

### WebSocket Stats

Get WebSocket connection statistics.

**Endpoint**: `GET /api/v1/ws/stats`

**Headers**:
```
Authorization: Bearer <access_token>
```

**Response** (200 OK):
```json
{
  "success": true,
  "data": {
    "total_clients": 15,
    "connected_users": 12,
    "broadcast_backlog": 0
  }
}
```

## System Endpoints

### Health Check

Check the health status of all system components.

**Endpoint**: `GET /health`

**Response** (200 OK):
```json
{
  "status": "healthy",
  "timestamp": "2025-01-01T10:00:00Z",
  "version": "1.0.0",
  "services": {
    "database": true,
    "ethereum": false,
    "websocket": true,
    "cron_jobs": {
      "enabled": true,
      "total_jobs": 3,
      "jobs": [
        {
          "id": 1,
          "next_run": "2025-01-02T02:00:00Z",
          "prev_run": "2025-01-01T02:00:00Z"
        }
      ]
    },
    "achievements": true
  }
}
```

### Prometheus Metrics

Get application metrics in Prometheus format.

**Endpoint**: `GET /metrics`

**Response** (200 OK):
```
# HELP paperplay_http_requests_total Total number of HTTP requests
# TYPE paperplay_http_requests_total counter
paperplay_http_requests_total{endpoint="/health",method="GET",status="200"} 1

# HELP paperplay_active_connections Number of active connections
# TYPE paperplay_active_connections gauge
paperplay_active_connections 5

# HELP paperplay_user_logins_total Total number of user login attempts
# TYPE paperplay_user_logins_total counter
paperplay_user_logins_total{status="success"} 10
paperplay_user_logins_total{status="failed"} 2
```

### System Statistics

Get detailed system statistics (requires authentication).

**Endpoint**: `GET /api/v1/system/stats`

**Headers**:
```
Authorization: Bearer <access_token>
```

**Response** (200 OK):
```json
{
  "success": true,
  "data": {
    "websocket": {
      "total_clients": 15,
      "connected_users": 12,
      "broadcast_backlog": 0
    },
    "database": {
      "healthy": true
    }
  }
}
```

## Error Responses

All API endpoints follow a consistent error response format:

**Validation Error** (400 Bad Request):
```json
{
  "error": "validation_error",
  "message": "Validation failed",
  "details": "email: required field is missing"
}
```

**Unauthorized** (401 Unauthorized):
```json
{
  "error": "unauthorized",
  "message": "Invalid or expired token"
}
```

**Not Found** (404 Not Found):
```json
{
  "error": "not_found",
  "message": "Resource not found"
}
```

**Conflict** (409 Conflict):
```json
{
  "error": "user_exists",
  "message": "User with this email already exists"
}
```

**Internal Server Error** (500 Internal Server Error):
```json
{
  "error": "internal_error",
  "message": "An internal server error occurred",
  "details": "Database connection failed"
}
```

## Rate Limiting

Currently, no rate limiting is implemented. In production, consider implementing rate limiting based on:
- IP address for public endpoints
- User ID for authenticated endpoints
- Global rate limits for system protection

## Security Considerations

1. **HTTPS**: Use HTTPS in production
2. **JWT Secrets**: Change default JWT secret key
3. **CORS**: Configure appropriate CORS origins
4. **Input Validation**: All inputs are validated
5. **SQL Injection**: Protected by GORM ORM
6. **Password Security**: Passwords are hashed with bcrypt

## Cron Jobs

The system runs background jobs for:

1. **Daily Stats Update** (2:00 AM daily)
   - Updates user learning streaks
   - Calculates review recommendations
   - Updates spaced repetition schedules

2. **Weekly Report Generation** (3:00 AM Sundays)
   - Generates weekly learning reports
   - Sends notifications to active users

3. **Achievement Check** (Every 5 minutes)
   - Evaluates user achievements
   - Awards new achievements
   - Triggers NFT minting (if enabled)

## Environment Variables

Configure the application using environment variables:

```bash
# Server Configuration
SERVER_PORT=8080
SERVER_MODE=debug

# Database
DATABASE_DSN=./data/paperplay.db

# JWT
JWT_SECRET_KEY=your-super-secret-key
JWT_ACCESS_TOKEN_DURATION=15
JWT_REFRESH_TOKEN_DURATION=7

# Ethereum (optional)
ETHEREUM_ENABLED=false
ETHEREUM_NETWORK_URL=https://sepolia.infura.io/v3/your-project-id
ETHEREUM_CHAIN_ID=11155111

# Logging
LOG_LEVEL=info
LOG_OUTPUT_PATH=./logs/app.log

# Cron Jobs
CRON_ENABLED=true
```

## Future API Endpoints

The following endpoints are planned for future implementation:

- `GET /api/v1/subjects` - Get all subjects
- `GET /api/v1/subjects/:id` - Get subject details
- `GET /api/v1/subjects/:id/roadmap` - Get subject roadmap
- `GET /api/v1/levels/:id` - Get level details
- `POST /api/v1/levels/:id/start` - Start a level
- `POST /api/v1/levels/:id/submit` - Submit answer
- `POST /api/v1/levels/:id/complete` - Complete level
- `GET /api/v1/stats/dashboard` - Get user dashboard
- `GET /api/v1/stats/progress` - Get detailed progress stats 