# PaperPlay Test Suite

这个目录包含了PaperPlay应用的完整测试套件，包括单元测试、集成测试和性能测试。

## 测试文件结构

```
scripts/
├── README.md                 # 测试说明文档（本文件）
├── run_all_tests.sh         # 主测试运行器
├── test_all_endpoints.sh    # 完整的API集成测试
└── test_performance.sh      # 性能测试脚本

internal/
├── api/
│   ├── user_test.go         # 用户API单元测试
│   ├── level_test.go        # 关卡API单元测试
│   └── achievement_test.go  # 成就API单元测试
├── service/
│   └── user_test.go         # 用户服务单元测试
└── model/
    └── user_test.go         # 用户模型单元测试
```

## 前置要求

确保以下工具已安装：

- **Go 1.21+**: 用于运行单元测试
- **curl**: 用于HTTP请求测试
- **jq**: 用于JSON数据处理
- **Running Server**: 服务器必须在 `http://localhost:8080` 运行

### 安装依赖

```bash
# Ubuntu/Debian
sudo apt-get update
sudo apt-get install curl jq

# macOS
brew install curl jq

# 检查依赖
./run_all_tests.sh --check-deps
```

## 快速开始

### 1. 运行所有测试

```bash
# 运行所有测试（单元测试 + 集成测试 + 性能测试）
./run_all_tests.sh
```

### 2. 运行特定类型的测试

```bash
# 只运行单元测试
./run_all_tests.sh --unit

# 只运行集成测试
./run_all_tests.sh --integration

# 只运行性能测试
./run_all_tests.sh --performance
```

### 3. 运行单个测试脚本

```bash
# 运行API集成测试
./test_all_endpoints.sh

# 运行性能测试
./test_performance.sh
```

## 测试详情

### 单元测试

单元测试使用Go的标准测试框架和testify库：

```bash
# 在项目根目录运行
cd /path/to/paperplay/backend
go test -v ./...

# 运行带覆盖率的测试
go test -v -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

#### 测试覆盖的功能：

- **API层**：HTTP处理器的请求/响应逻辑
- **服务层**：业务逻辑和数据处理
- **模型层**：数据模型和数据库操作

### 集成测试

集成测试验证所有API端点的功能：

#### 测试的API端点：

**用户系统 (8个端点)**
- `POST /api/v1/auth/register` - 用户注册
- `POST /api/v1/auth/login` - 用户登录
- `POST /api/v1/auth/refresh` - 刷新令牌
- `GET /api/v1/users/profile` - 获取用户资料
- `PUT /api/v1/users/profile` - 更新用户资料
- `GET /api/v1/users/progress` - 获取用户进度
- `GET /api/v1/users/achievements` - 获取用户成就
- `POST /api/v1/users/logout` - 用户登出

**关卡系统 (12个端点)**
- `GET /api/v1/subjects` - 列出所有学科
- `GET /api/v1/subjects/{subject_id}` - 获取单个学科
- `GET /api/v1/subjects/{subject_id}/papers` - 列出学科下的论文
- `GET /api/v1/papers/{paper_id}` - 获取单篇论文
- `GET /api/v1/papers/{paper_id}/level` - 获取论文对应的关卡
- `GET /api/v1/levels/{level_id}` - 获取单个关卡
- `GET /api/v1/levels/{level_id}/questions` - 列出关卡中的问题
- `GET /api/v1/questions/{question_id}` - 获取单个问题
- `GET /api/v1/subjects/{subject_id}/roadmap` - 获取学科路线图
- `POST /api/v1/levels/{level_id}/start` - 开始关卡
- `POST /api/v1/levels/{level_id}/submit` - 提交答案
- `POST /api/v1/levels/{level_id}/complete` - 完成关卡

**成就系统 (3个端点)**
- `GET /api/v1/achievements` - 获取所有成就
- `GET /api/v1/achievements/user` - 获取用户的成就
- `POST /api/v1/achievements/evaluate` - 触发成就评估

**系统监控 (4个端点)**
- `GET /health` - 健康检查
- `GET /metrics` - Prometheus指标
- `GET /api/v1/ws/stats` - WebSocket统计
- `GET /api/v1/system/stats` - 系统统计

#### 错误场景测试：
- 未授权访问 (401)
- 无效登录凭据 (401)
- 无效注册数据 (400)
- 资源不存在 (404)

### 性能测试

性能测试评估系统在负载下的表现：

#### 测试指标：
- **响应时间**: 各端点的响应时间（< 500ms）
- **并发处理**: 同时处理多个请求的能力
- **吞吐量**: 每秒处理的请求数
- **内存使用**: 负载下的内存增长
- **稳定性**: 持续负载下的系统稳定性

#### 性能基准：
- 健康检查: < 100ms
- API端点: < 500ms
- 并发请求: 10个请求 < 2秒
- 吞吐量: > 50 请求/秒

## 测试报告

### 自动生成报告

```bash
# 生成完整的测试报告
./run_all_tests.sh --report
```

报告包含：
- 单元测试覆盖率
- 集成测试结果
- 性能测试指标
- 项目结构概览

### 手动查看覆盖率

```bash
# 生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# 在浏览器中查看
open coverage.html  # macOS
xdg-open coverage.html  # Linux
```

## CI/CD 集成

在CI/CD管道中使用：

```yaml
# GitHub Actions 示例
- name: Run Tests
  run: |
    cd backend
    scripts/run_all_tests.sh

# GitLab CI 示例
test:
  script:
    - cd backend
    - ./scripts/run_all_tests.sh
```

## 故障排除

### 常见问题

1. **服务器未运行**
   ```
   Error: Server is not running or not healthy
   Solution: 启动服务器 ./paperplay
   ```

2. **依赖缺失**
   ```
   Error: jq is required but not installed
   Solution: 安装 jq - sudo apt-get install jq
   ```

3. **权限问题**
   ```
   Error: Permission denied
   Solution: chmod +x scripts/*.sh
   ```

4. **端口冲突**
   ```
   Error: Connection refused
   Solution: 检查端口 8080 是否被占用
   ```

### 调试模式

```bash
# 运行带详细输出的测试
./test_all_endpoints.sh 2>&1 | tee test_output.log

# 检查特定端点
curl -v http://localhost:8080/health
```

## 贡献指南

### 添加新的测试

1. **单元测试**：在对应包目录下创建 `*_test.go` 文件
2. **集成测试**：在 `test_all_endpoints.sh` 中添加新的测试函数
3. **性能测试**：在 `test_performance.sh` 中添加性能基准测试

### 测试最佳实践

1. **命名约定**：
   - 单元测试：`TestFunctionName`
   - 集成测试：`test_endpoint_name`

2. **测试结构**：
   - Arrange: 设置测试数据
   - Act: 执行被测试的功能
   - Assert: 验证结果

3. **错误处理**：
   - 测试正常情况和错误情况
   - 验证错误消息和状态码

## 性能优化建议

基于测试结果的优化建议：

1. **响应时间优化**：
   - 数据库查询优化
   - 缓存策略
   - 并发处理优化

2. **内存优化**：
   - 避免内存泄漏
   - 优化数据结构
   - 及时释放资源

3. **并发优化**：
   - 连接池管理
   - 异步处理
   - 负载均衡

---

有关测试的任何问题，请查看测试输出或联系开发团队。 