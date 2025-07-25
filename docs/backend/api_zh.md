# PaperPlay 后端 API 文档

## 概览

PaperPlay 是一个基于学术论文的游戏化学习平台。本文档描述了后端系统的 REST API 端点和 WebSocket 接口。

**基础 URL**: `http://localhost:8080`
**API 版本**: v1
**认证方式**: JWT Bearer Token
**内容类型**: `application/json`

## 认证

### 注册用户

注册一个新用户账户。

**端点**: `POST /api/v1/auth/register`

**请求体**:

```json
{
  "email": "user@example.com",
  "password": "password123",
  "display_name": "John Doe"
}
```

**响应** (201 Created):

```json
{
  "success": true,
  "message": "用户注册成功",
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

### 用户登录

认证一个已存在的用户。

**端点**: `POST /api/v1/auth/login`

**请求体**:

```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**响应** (200 OK):

```json
{
  "success": true,
  "message": "登录成功",
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

### 刷新令牌

刷新一个已过期的访问令牌。

**端点**: `POST /api/v1/auth/refresh`

**请求体**:

```json
{
  "refresh_token": "refresh-token-string"
}
```

**响应** (200 OK):

```json
{
  "success": true,
  "message": "令牌刷新成功",
  "data": {
    "access_token": "new-jwt-token",
    "refresh_token": "new-refresh-token",
    "expires_in": 900
  }
}
```

## 用户管理

所有用户端点都需要通过 `Authorization: Bearer <access_token>` 请求头进行认证。

### 获取用户资料

获取当前用户的个人资料信息。

**端点**: `GET /api/v1/users/profile`

**请求头**:

```
Authorization: Bearer <access_token>
```

**响应** (200 OK):

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

### 更新用户资料

更新当前用户的个人资料信息。

**端点**: `PUT /api/v1/users/profile`

**请求头**:

```
Authorization: Bearer <access_token>
```

**请求体**:

```json
{
  "display_name": "Jane Doe",
  "avatar_url": "https://example.com/new-avatar.jpg"
}
```

**响应** (200 OK):

```json
{
  "success": true,
  "message": "资料更新成功",
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

### 获取用户进度

获取用户在所有学科上的学习进度。

**端点**: `GET /api/v1/users/progress`

**请求头**:

```
Authorization: Bearer <access_token>
```

**响应** (200 OK):

```json
{
  "success": true,
  "data": {
    "subjects": [
      {
        "subject": {
          "id": "uuid",
          "name": "Computer Science",
          "description": "与计算机科学相关的论文"
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
        "description": "已完成关卡：机器学习入门",
        "level_id": "uuid",
        "subject_id": "uuid"
      }
    ]
  }
}
```

### 获取用户成就

获取用户已获得的成就。

**端点**: `GET /api/v1/users/achievements`

**请求头**:

```
Authorization: Bearer <access_token>
```

**响应** (200 OK):

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
        "name": "第一步",
        "description": "完成你的第一个关卡",
        "icon_url": "https://example.com/icon.png",
        "level": 1,
        "is_active": true,
        "nft_metadata": {
          "name": "“第一步” NFT",
          "description": "为纪念您完成的第一个关卡",
          "image": "https://example.com/nft.png"
        }
      }
    }
  ]
}
```

### 用户登出

使用户当前的令牌失效。

**端点**: `POST /api/v1/users/logout`

**请求头**:

```
Authorization: Bearer <access_token>
```

**响应** (200 OK):

```json
{
  "success": true,
  "message": "登出成功"
}
```

## 关卡系统

### 列出所有学科

**端点**: `GET /api/v1/subjects`

**请求头**

```
Authorization: Bearer <access_token>
```

**响应** (200 OK)

```json
{
  "success": true,
  "message": "学科检索成功",
  "data": [
    {
      "id": "uuid-string",
      "name": "Computer Vision",
      "description": "使计算机能够理解图像的研究",
      "created_at": "2025-07-25T10:00:00Z",
      "updated_at": "2025-07-25T10:00:00Z"
    },
    {
      "id": "uuid-string",
      "name": "Natural Language Processing",
      "description": "处理和生成人类语言的技术",
      "created_at": "2025-07-20T14:30:00Z",
      "updated_at": "2025-07-20T14:30:00Z"
    }
  ]
}
```

---

### 获取单个学科

**端点**: `GET /api/v1/subjects/{subject_id}`

**路径参数**

* `subject_id` (UUID)

**请求头**

```
Authorization: Bearer <access_token>
```

**响应** (200 OK)

```json
{
  "success": true,
  "message": "学科检索成功",
  "data": {
    "id": "uuid-string",
    "name": "Computer Vision",
    "description": "使计算机能够理解图像的研究",
    "created_at": "2025-07-25T10:00:00Z",
    "updated_at": "2025-07-25T10:00:00Z"
  }
}
```

---

### 列出学科下的论文

**端点**: `GET /api/v1/subjects/{subject_id}/papers`

**路径参数**

* `subject_id` (UUID)

**请求头**

```
Authorization: Bearer <access_token>
```

**响应** (200 OK)

```json
{
  "success": true,
  "message": "论文检索成功",
  "data": [
    {
      "id": "uuid-string",
      "subject_id": "uuid-string",
      "title": "Deep Learning for Image Recognition",
      "citation": "Goodfellow et al., 2016",
      "created_at": "2025-06-01T08:00:00Z",
      "updated_at": "2025-06-01T08:00:00Z"
    },
    {
      "id": "uuid-string",
      "subject_id": "uuid-string",
      "title": "Convolutional Neural Networks",
      "citation": "LeCun et al., 1998",
      "created_at": "2025-05-15T16:20:00Z",
      "updated_at": "2025-05-15T16:20:00Z"
    }
  ]
}
```

---

### 获取单篇论文

**端点**: `GET /api/v1/papers/{paper_id}`

**路径参数**

* `paper_id` (UUID)

**请求头**

```
Authorization: Bearer <access_token>
```

**响应** (200 OK)

```json
{
  "success": true,
  "message": "论文检索成功",
  "data": {
    "id": "uuid-string",
    "subject_id": "uuid-string",
    "title": "Deep Learning for Image Recognition",
    "citation": "Goodfellow et al., 2016",
    "created_at": "2025-06-01T08:00:00Z",
    "updated_at": "2025-06-01T08:00:00Z"
  }
}
```

---

### 获取论文对应的关卡

**端点**: `GET /api/v1/papers/{paper_id}/level`

**路径参数**

* `paper_id` (UUID)

**请求头**

```
Authorization: Bearer <access_token>
```

**响应** (200 OK)

```json
{
  "success": true,
  "message": "关卡检索成功",
  "data": {
    "id": "uuid-string",
    "paper_id": "uuid-string",
    "paper_author": "Goodfellow; Bengio; Courville",
    "paper_pub_ym": "2016-11",
    "citation_count": 5240,
    "name": "深度学习入门",
    "pass_condition": "{\"min_score\":80}",
    "meta_json": "{}",
    "created_at": "2025-06-01T08:00:00Z",
    "updated_at": "2025-06-01T08:00:00Z"
  }
}
```

---

### 获取单个关卡

**端点**: `GET /api/v1/levels/{level_id}`

**路径参数**

* `level_id` (UUID)

**请求头**

```
Authorization: Bearer <access_token>
```

**响应** (200 OK)

```json
{
  "success": true,
  "message": "关卡检索成功",
  "data": {
    "id": "uuid-string",
    "paper_id": "uuid-string",
    "paper_author": "Goodfellow; Bengio; Courville",
    "paper_pub_ym": "2016-11",
    "citation_count": 5240,
    "name": "深度学习入门",
    "pass_condition": "{\"min_score\":80}",
    "meta_json": "{}",
    "created_at": "2025-06-01T08:00:00Z",
    "updated_at": "2025-06-01T08:00:00Z"
  }
}
```

---

### 列出关卡中的问题

**端点**: `GET /api/v1/levels/{level_id}/questions`

**路径参数**

* `level_id` (UUID)

**请求头**

```
Authorization: Bearer <access_token>
```

**响应** (200 OK)

```json
{
  "success": true,
  "message": "问题检索成功",
  "data": [
    {
      "id": "uuid-string",
      "level_id": "uuid-string",
      "stem": "什么是反向传播？",
      "content_json": "{\"type\":\"mcq\",\"options\":[\"...\",\"...\"]}",
      "answer_json": "{\"correct\":0}",
      "score": 10,
      "difficulty": 3,
      "created_by": "AI‑Agent‑v1",
      "created_at": "2025-06-01T08:00:00Z"
    },
    {
      "id": "uuid-string",
      "level_id": "uuid-string",
      "stem": "解释卷积层。",
      "content_json": "{\"type\":\"essay\"}",
      "answer_json": "{\"keywords\":[\"filter\",\"stride\"]}",
      "score": 20,
      "difficulty": 4,
      "created_by": "Author‑X",
      "created_at": "2025-06-01T08:00:00Z"
    }
  ]
}
```

---

### 获取单个问题

**端点**: `GET /api/v1/questions/{question_id}`

**路径参数**

* `question_id` (UUID)

**请求头**

```
Authorization: Bearer <access_token>
```

**响应** (200 OK)

```json
{
  "success": true,
  "message": "问题检索成功",
  "data": {
    "id": "uuid-string",
    "level_id": "uuid-string",
    "stem": "什么是反向传播？",
    "content_json": "{\"type\":\"mcq\",\"options\":[\"...\",\"...\"]}",
    "answer_json": "{\"correct\":0}",
    "score": 10,
    "difficulty": 3,
    "created_by": "AI‑Agent‑v1",
    "created_at": "2025-06-01T08:00:00Z"
  }
}
```

### 获取学科路线图

**端点**: `GET /api/v1/subjects/{subject_id}/roadmap`

**路径参数**

* `subject_id` (UUID)

**请求头**

```
Authorization: Bearer <access_token>
```

**响应** (200 OK)

```json
{
  "success": true,
  "message": "路线图检索成功",
  "data": [
    {
      "id": "uuid-string",
      "subject_id": "uuid-string",
      "level_id": "uuid-string",
      "parent_id": null,
      "sort": 1
    },
    {
      "id": "uuid-string",
      "subject_id": "uuid-string",
      "level_id": "uuid-string",
      "parent_id": "uuid-string",
      "sort": 2
    }
  ]
}
```

### 开始一个关卡

**端点**: `POST /api/v1/levels/{level_id}/start`

**路径参数**

* `level_id` (UUID)

**请求头**

```
Authorization: Bearer <access_token>
```

**响应** (200 OK)

```json
{
  "success": true,
  "message": "关卡已开始",
  "data": {
    "user_id": "uuid-string",
    "level_id": "uuid-string",
    "status": 1,
    "started_at": "2025-07-25T11:00:00Z"
  }
}
```

### 提交答案

**端点**: `POST /api/v1/levels/{level_id}/submit`

**路径参数**

* `level_id` (UUID)

**请求头**

```
Authorization: Bearer <access_token>
```

**请求体**

```json
{
  "question_id": "uuid-string",
  "answer_json": { /* 用户的答案 */ },
  "duration_ms": 120000
}
```

**响应** (200 OK)

```json
{
  "success": true,
  "message": "答案已提交",
  "data": {
    "question_id": "uuid-string",
    "is_correct": true,
    "score": 10,
    "total_score": 30
  }
}
```

### 完成关卡

**端点**: `POST /api/v1/levels/{level_id}/complete`

**路径参数**

* `level_id` (UUID)

**请求头**

```
Authorization: Bearer <access_token>
```

**响应** (200 OK)

```json
{
  "success": true,
  "message": "关卡已完成",
  "data": {
    "user_id": "uuid-string",
    "level_id": "uuid-string",
    "score": 80,
    "stars": 3,
    "completed_at": "2025-07-25T11:15:00Z"
  }
}
```

---

> **错误码**
>
> * `401 Unauthorized`: 缺少或无效的 JWT
> * `404 Not Found`: 资源未找到 (例如，无效的 `subject_id`, `paper_id`, `level_id`, 或 `question_id`)
> * `500 Internal Server Error`: 服务器端错误

## 成就系统

### 获取所有成就

获取系统中所有可用的成就。

**端点**: `GET /api/v1/achievements`

**请求头**:

```
Authorization: Bearer <access_token>
```

**响应** (200 OK):

```json
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "name": "第一步",
      "description": "完成你的第一个关卡",
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
        "name": "“第一步” NFT",
        "description": "为纪念您完成的第一个关卡",
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

### 获取用户的成就

获取当前用户获得的成就。

**端点**: `GET /api/v1/achievements/user`

**请求头**:

```
Authorization: Bearer <access_token>
```

**响应** (200 OK):

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
        "name": "第一步",
        "description": "完成你的第一个关卡",
        "level": 1
      }
    }
  ]
}
```

### 触发成就评估

为当前用户手动触发成就评估（主要用于测试）。

**端点**: `POST /api/v1/achievements/evaluate`

**请求头**:

```
Authorization: Bearer <access_token>
```

**响应** (200 OK):

```json
{
  "success": true,
  "message": "成就评估完成"
}
```

## WebSocket 接口

### WebSocket 连接

连接到 WebSocket 服务器以获取实时通知。

**端点**: `GET /ws`

**查询参数** (可选):

- `token`: 用于认证连接的 JWT 访问令牌

**连接 URL**: `ws://localhost:8080/ws`

### WebSocket 消息类型

#### 客户端到服务器消息

**Ping 消息**:

```json
{
  "type": "ping"
}
```

**订阅频道**:

```json
{
  "type": "subscribe",
  "channel": "achievements"
}
```

#### 服务器到客户端消息

**连接确认**:

```json
{
  "type": "connected",
  "data": {
    "status": "connected"
  },
  "timestamp": "2025-01-01T10:00:00Z"
}
```

**Pong 响应**:

```json
{
  "type": "pong",
  "data": {
    "status": "ok"
  },
  "timestamp": "2025-01-01T10:00:00Z"
}
```

**成就通知**:

```json
{
  "type": "notification",
  "user_id": "uuid",
  "data": {
    "id": "notification-uuid",
    "type": "achievement",
    "title": "新成就！",
    "message": "你获得了“第一步”成就！",
    "icon": "https://example.com/icon.png",
    "achievement": {
      "id": "uuid",
      "name": "第一步",
      "description": "完成你的第一个关卡",
      "level": 1,
      "icon_url": "https://example.com/icon.png"
    }
  },
  "timestamp": "2025-01-01T10:00:00Z"
}
```

**周报通知**:

```json
{
  "type": "notification",
  "user_id": "uuid",
  "data": {
    "id": "notification-uuid",
    "type": "weekly_report",
    "title": "学习周报",
    "message": "本周你学习了5天，正确回答了50个问题！"
  },
  "timestamp": "2025-01-01T10:00:00Z"
}
```

### WebSocket 统计信息

获取 WebSocket 连接统计信息。

**端点**: `GET /api/v1/ws/stats`

**请求头**:

```
Authorization: Bearer <access_token>
```

**响应** (200 OK):

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

## 系统端点

### 健康检查

检查所有系统组件的健康状态。

**端点**: `GET /health`

**响应** (200 OK):

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

### Prometheus 指标

以 Prometheus 格式获取应用程序指标。

**端点**: `GET /metrics`

**响应** (200 OK):

```
# HELP paperplay_http_requests_total HTTP 请求总数
# TYPE paperplay_http_requests_total counter
paperplay_http_requests_total{endpoint="/health",method="GET",status="200"} 1

# HELP paperplay_active_connections 活跃连接数
# TYPE paperplay_active_connections gauge
paperplay_active_connections 5

# HELP paperplay_user_logins_total 用户登录尝试总数
# TYPE paperplay_user_logins_total counter
paperplay_user_logins_total{status="success"} 10
paperplay_user_logins_total{status="failed"} 2
```

### 系统统计

获取详细的系统统计信息（需要认证）。

**端点**: `GET /api/v1/system/stats`

**请求头**:

```
Authorization: Bearer <access_token>
```

**响应** (200 OK):

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

## 错误响应

所有 API 端点都遵循一致的错误响应格式：

**验证错误** (400 Bad Request):

```json
{
  "error": "validation_error",
  "message": "验证失败",
  "details": "email: 必填字段缺失"
}
```

**未经授权** (401 Unauthorized):

```json
{
  "error": "unauthorized",
  "message": "无效或过期的令牌"
}
```

**未找到** (404 Not Found):

```json
{
  "error": "not_found",
  "message": "资源未找到"
}
```

**冲突** (409 Conflict):

```json
{
  "error": "user_exists",
  "message": "该邮箱已存在用户"
}
```

**内部服务器错误** (500 Internal Server Error):

```json
{
  "error": "internal_error",
  "message": "发生内部服务器错误",
  "details": "数据库连接失败"
}
```

## 速率限制

目前沒有实施速率限制。在生产环境中，应考虑根据以下条件实施速率限制：

- 公共端点的 IP 地址
- 认证端点的用户 ID
- 用于系统保护的全局速率限制

## 安全注意事项

1. **HTTPS**: 在生产环境中使用 HTTPS
2. **JWT 密钥**: 更改默认的 JWT 密钥
3. **CORS**: 配置适当的 CORS 源
4. **输入验证**: 所有输入都经过验证
5. **SQL 注入**: 由 GORM ORM 提供保护
6. **密码安全**: 密码使用 bcrypt 进行哈希处理

## 定时任务

系统运行以下后台任务：

1. **每日统计更新** (每天凌晨 2:00)

   - 更新用户学习连续记录
   - 计算复习建议
   - 更新间隔重复时间表
2. **每周报告生成** (每周日凌晨 3:00)

   - 生成每周学习报告
   - 向活跃用户发送通知
3. **成就检查** (每 5 分钟)

   - 评估用户成就
   - 授予新成就
   - 触发 NFT 铸造 (如果启用)

## 环境变量

使用环境变量配置应用程序：

```bash
# 服务器配置
SERVER_PORT=8080
SERVER_MODE=debug

# 数据库
DATABASE_DSN=./data/paperplay.db

# JWT
JWT_SECRET_KEY=your-super-secret-key
JWT_ACCESS_TOKEN_DURATION=15
JWT_REFRESH_TOKEN_DURATION=7

# 以太坊 (可选)
ETHEREUM_ENABLED=false
ETHEREUM_NETWORK_URL=https://sepolia.infura.io/v3/your-project-id
ETHEREUM_CHAIN_ID=11155111

# 日志
LOG_LEVEL=info
LOG_OUTPUT_PATH=./logs/app.log

# 定时任务
CRON_ENABLED=true
```

## 未来 API 端点

计划在未来实现以下端点：

- `GET /api/v1/subjects` - 获取所有学科
- `GET /api/v1/subjects/:id` - 获取学科详情
- `GET /api/v1/levels/:id` - 获取关卡详情
- `GET /api/v1/stats/dashboard` - 获取用户仪表盘
- `GET /api/v1/stats/progress` - 获取详细进度统计
