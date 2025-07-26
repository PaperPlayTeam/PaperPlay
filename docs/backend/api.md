## ✅ Level System API 端点

以下Level System端点已实现并可正常使用：

### ✅ 获取所有学科

**端点**: `GET /api/v1/subjects`

**请求头**:

```
Authorization: Bearer <access_token>
```

**响应** (200 OK):

```json
{
  "success": true,
  "message": "Subjects retrieved successfully",
  "data": [
    {
      "id": "uuid-string",
      "name": "Computer Vision",
      "description": "AI and computer vision research",
      "created_at": "2025-07-25T10:00:00Z",
      "updated_at": "2025-07-25T10:00:00Z"
    }
  ]
}
```

### ✅ 获取单个学科

**端点**: `GET /api/v1/subjects/{subject_id}`

**响应** (200 OK):

```json
{
  "success": true,
  "message": "Subject retrieved successfully",
  "data": {
    "id": "uuid-string",
    "name": "Computer Vision",
    "description": "AI and computer vision research",
    "created_at": "2025-07-25T10:00:00Z",
    "updated_at": "2025-07-25T10:00:00Z"
  }
}
```

### ✅ 获取学科下的论文

**端点**: `GET /api/v1/subjects/{subject_id}/papers`

### ✅ 获取单个论文

**端点**: `GET /api/v1/papers/{paper_id}`

### ✅ 获取论文对应的关卡

**端点**: `GET /api/v1/papers/{paper_id}/level`

### ✅ 获取单个关卡

**端点**: `GET /api/v1/levels/{level_id}`

### ✅ 获取关卡中的题目

**端点**: `GET /api/v1/levels/{level_id}/questions`

### ✅ 获取单个题目

**端点**: `GET /api/v1/questions/{question_id}`

### ✅ 获取学科路线图

**端点**: `GET /api/v1/subjects/{subject_id}/roadmap`

### ✅ 开始关卡

**端点**: `POST /api/v1/levels/{level_id}/start`

### ✅ 提交答案

**端点**: `POST /api/v1/levels/{level_id}/submit`

### ✅ 完成关卡

**端点**: `POST /api/v1/levels/{level_id}/complete`

---

## ✅ 成就系统 API 端点

以下成就系统端点已实现并可正常使用：

### ✅ 获取所有成就

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
      "name": "首战告捷",
      "description": "第一次作答即答对任意一道题目",
      "icon_url": "",
      "level": 1,
      "category": "learning",
      "is_active": true,
      "rules": {
        "type": "first_try",
        "conditions": [
          {
            "field": "attempts_first_try_correct",
            "operator": ">=",
            "value": 1
          }
        ]
      },
      "nft_metadata": {
        "name": "First Victory Badge",
        "description": "Awarded for getting first question right on first try",
        "image": "",
        "attributes": [
          {
            "trait_type": "Achievement Type",
            "value": "Learning"
          }
        ]
      }
    }
  ]
}
```

### ✅ 获取用户的成就

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
        "name": "首战告捷",
        "description": "第一次作答即答对任意一道题目",
        "level": 1,
        "icon_url": ""
      }
    }
  ]
}
```

### ✅ 触发成就评估

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

## ✅ WebSocket 接口

### WebSocket 连接

连接到 WebSocket 服务器以获取实时通知。

**端点**: `GET /ws`

**查询参数** (可选):

- `token`: 用于认证连接的 JWT 访问令牌

**连接 URL**: `ws://localhost:8080/ws`

### ✅ WebSocket 统计信息

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
    "total_clients": 0,
    "connected_users": 0,
    "broadcast_backlog": 0
  }
}
```

## ✅ 系统端点

### ✅ 健康检查

检查所有系统组件的健康状态。

**端点**: `GET /health`

**响应** (200 OK):

```json
{
  "status": "healthy",
  "timestamp": "2025-07-26T01:50:44.630975257Z",
  "version": "1.0.0",
  "services": {
    "database": true,
    "ethereum": false,
    "websocket": true,
    "achievements": true
  }
}
```

### ✅ Prometheus 指标

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

### ✅ 系统统计

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
      "total_clients": 0,
      "connected_users": 0,
      "broadcast_backlog": 0
    },
    "database": {
      "healthy": true
    },
    "metrics": {
      "status": "healthy",
      "metrics_enabled": true,
      "registry_status": "active"
    }
  }
}
```

## ✅ 错误响应

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
  "error": "Authorization header is required"
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

---

## 📋 实现总结

以下API端点已成功实现并经过测试验证：

### Level System (12个端点)
- ✅ `GET /api/v1/subjects` - 获取所有学科
- ✅ `GET /api/v1/subjects/:id` - 获取学科详情  
- ✅ `GET /api/v1/subjects/:id/papers` - 获取学科下的论文
- ✅ `GET /api/v1/subjects/:id/roadmap` - 获取学科路线图
- ✅ `GET /api/v1/papers/:id` - 获取论文详情
- ✅ `GET /api/v1/papers/:id/level` - 获取论文对应的关卡
- ✅ `GET /api/v1/levels/:id` - 获取关卡详情
- ✅ `GET /api/v1/levels/:id/questions` - 获取关卡中的题目
- ✅ `GET /api/v1/questions/:id` - 获取题目详情
- ✅ `POST /api/v1/levels/:id/start` - 开始关卡
- ✅ `POST /api/v1/levels/:id/submit` - 提交答案
- ✅ `POST /api/v1/levels/:id/complete` - 完成关卡

### Achievement System (3个端点)
- ✅ `GET /api/v1/achievements` - 获取所有成就
- ✅ `GET /api/v1/achievements/user` - 获取用户的成就
- ✅ `POST /api/v1/achievements/evaluate` - 触发成就评估

### System & Monitoring (4个端点)
- ✅ `GET /health` - 系统健康检查
- ✅ `GET /metrics` - Prometheus指标
- ✅ `GET /api/v1/ws/stats` - WebSocket统计
- ✅ `GET /api/v1/system/stats` - 详细系统统计

### 🔮 计划实现的端点

- `GET /api/v1/stats/dashboard` - 用户仪表板
- `GET /api/v1/stats/progress` - 详细进度统计

**总计：已实现19个API端点，覆盖了学习系统、成就系统和系统监控的核心功能。** 