# 字段设计

## 关卡系统

关键是学科和论文/关卡的一对多关系，论文和关卡的一对一关系，关卡和题目的一对多关系；以及路线图的多叉树结构怎么定位

**subjects**

| 列名                      | 类型      | 约束       | 说明   |
| ----------------------- | ------- | -------- | ---- |
| id                      | TEXT    | PK UUID  | 学科主键 |
| name                    | TEXT    | NOT NULL | 学科名称 |
| description             | TEXT    |          | 说明   |
| created_at / updated_at | INTEGER | NOT NULL | 时间戳  |

**papers**

| 列名                      | 类型      | 约束                                  | 说明                 |
| ----------------------- | ------- | ----------------------------------- | ------------------ |
| id                      | TEXT    | PK UUID                             | 论文主键               |
| subject_id              | TEXT    | FK → subjects(id) ON DELETE CASCADE | 所属学科               |
| title                   | TEXT    | NOT NULL                            | 论文题目               |
| paper_author            | TEXT    | NOT NULL                            | 对应论文作者 (多作者用`;`分隔) |
| paper_pub_ym            | TEXT    | NOT NULL                            | 发表年月               |
| paper_citation_count    | TEXT    | NOT NULL                            | 引用信息               |
| created_at / updated_at | INTEGER |                                     |                    |

**levels**

| 列名                      | 类型      | 约束                     | 说明        |
| ----------------------- | ------- | ---------------------- | --------- |
| id                      | TEXT    | PK UUID                | 关卡主键      |
| paper_id                | TEXT    | UNIQUE FK → papers(id) | 对应论文      |
| name                    | TEXT    | NOT NULL               | 关卡标题      |
| pass_condition          | TEXT    | NOT NULL               | 通关条件 JSON |
| meta_json               | TEXT    |                        | 其他元数据     |
| x                       | INTEGER | NOT NULL               | UI上的横坐标   |
| y                       | INTEGER | NOT NULL               | UI上的纵坐标   |
| created_at / updated_at | INTEGER |                        |           |


**questions**

| 列名           | 类型      | 约束                                | 说明         |
| ------------ | ------- | --------------------------------- | ---------- |
| id           | TEXT    | PK UUID                           | 题目主键       |
| level_id     | TEXT    | FK → levels(id) ON DELETE CASCADE | 所属关卡       |
| subtitle     | TEXT    |                                   | 小标题        |
| stem         | TEXT    | NOT NULL                          | 题干         |
| content_json | TEXT    | NOT NULL                          | 题面         |
| answer_json  | TEXT    | NOT NULL                          | 标准答案       |
| score        | INTEGER | NOT NULL                          | 分值         |
| created_by   | TEXT    |                                   | 出题者 / 生成引擎 |
| created_at   | INTEGER |                                   |            |

**roadmap_nodes**

多叉树直接用绝对路径+同级排序进行定位了，`path`就是绝对路径，其值类似于`001.001.001`，`depth`维护深度数据，`sort_order`维护同级nodes顺序

| 列名         | 类型      | 约束                          | 说明                   |
| ---------- | ------- | --------------------------- | -------------------- |
| id         | TEXT    | PK UUID                     | 节点主键                 |
| subject_id | TEXT    | FK → subjects(id)           | 所属学科                 |
| level_id   | TEXT    | FK → levels(id)             | 对应关卡                 |
| parent_id  | TEXT    | FK → roadmap_nodes(id) NULL | 父节点，NULL 表示根         |
| sort_order | INTEGER | NOT NULL DEFAULT 1          | 同层级内序号，从 1 开始        |
| path       | TEXT    | NOT NULL                    | 物化路径，如 `001.002.005` |
| depth      | INTEGER | NOT NULL                    | 深度，从 1（根）开始          |

## 用户系统

用户系统主要解决的问题有两个，一个是JWT鉴权，一个是用户进度维护，一个是对用户信息进行同级，以用于学习报告生成/主题推荐之类（这些也不一定包括在MVP里）

**users**

| 字段                      | 类型      | 约束              | 说明            |
| ----------------------- | ------- | --------------- | ------------- |
| id                      | TEXT    | PK UUID         | 主键            |
| email                   | TEXT    | UNIQUE NOT NULL | 登录账户          |
| password_hash           | TEXT    | NOT NULL        | PBKDF2/Bcrypt |
| display_name            | TEXT    |                 | 昵称            |
| avatar_url              | TEXT    |                 | 头像            |
| eth_address             | TEXT    |                 | 以太坊钱包地址       |
| eth_private_key         | TEXT    |                 | 以太坊私钥         |
| created_at / updated_at | INTEGER | NOT NULL        | UNIX ms       |

**refresh_tokens**

JWT就用 refresh tokens 解决了

|字段|类型|约束|说明|
|---|---|---|---|
|token|TEXT|PK|随机 UUID|
|user_id|TEXT|FK → users.id ON DELETE CASCADE||
|expires_at|INTEGER|NOT NULL|过期时间戳|

**user_progress**

维护用户学习进度

| 字段              | 类型      | 约束       | 说明                |
| --------------- | ------- | -------- | ----------------- |
| user_id         | TEXT    | PK,FK    |                   |
| level_id        | TEXT    | PK,FK    |                   |
| status          | INTEGER | NOT NULL | 0=未开始 1=进行中 2=已通过 |
| score           | INTEGER |          | 最新得分              |
| stars           | INTEGER |          | 星级 0‑3            |
| last_attempt_at | INTEGER |          | 最近一次作答            |

**user_attempts**

统计用户答题信息，这里的字段也可以用于触发成就，与成就系统互联

| 字段                         | 类型      | 约束                | 说明                                          |
| -------------------------- | ------- | ----------------- | ------------------------------------------- |
| stat_date                  | TEXT    | PK `YYYY‑MM‑DD`   | 统计日期                                        |
| user_id                    | TEXT    | PK FK → users(id) | 用户                                          |
| attempts_total             | INTEGER | DEFAULT 0         | 当日作答总题数                                     |
| attempts_correct           | INTEGER | DEFAULT 0         | 当日正确题数                                      |
| attempts_first_try_correct | INTEGER | DEFAULT 0         | 当日首次作答即正确的题数                                |
| correct_rate               | REAL    |                   | `attempts_correct / attempts_total`（方便直接查询） |
| first_try_correct_rate     | REAL    |                   | 当日首次正确率                                     |
| giveup_count               | INTEGER | DEFAULT 0         | 放弃/跳过题数                                     |
| skip_rate                  | REAL    |                   | `giveup_count / attempts_total`             |
| avg_duration_ms            | INTEGER |                   | 平均单题作答时长                                    |
| total_time_ms              | INTEGER |                   | 当日学习总时长                                     |
| sessions_count             | INTEGER |                   | 当日独立学习会话次数                                  |
| streak_days                | INTEGER |                   | 当前连续学习天数（由作业脚本更新）                           |
| review_due_count           | INTEGER |                   | 根据遗忘曲线计算当天应复习的关卡数量                          |
| retention_score            | REAL    |                   | 记忆保留度 0‑1（基于复习间隔模型）                         |
| peak_hour                  | INTEGER |                   | 学习活跃峰值小时（0‑23）                              |
| updated_at                 | INTEGER | NOT NULL          | 更新时间戳                                       |

“当日独立学习会话次数”（`sessions_count`）指的是用户在同一天内分开的、彼此之间有明显间隔的学习“块”或“段”数。具体来说：

1. **会话的定义**
    
    - **会话开始**：用户在某一时刻开始答题或浏览关卡。
        
    - **会话结束**：如果在一段时间内（通常设定一个阈值，比如 30 分钟）没有任何学习活动，则认为本次会话结束。
        
2. **如何计数**
    
    - 系统按时间顺序扫描该日所有的作答或事件日志。
        
    - 每当检测到距离上一次活动超过阈值，就认为是一个新的学习会话。
        
    - 最终累加当天所有的“新会话”数，写入 `sessions_count`

---

## 成就系统

**achievements**

| 列名           | 类型       | 约束                  | 描述                             |
| ------------ | -------- | ------------------- | ------------------------------ |
| id           | TEXT     | PRIMARY KEY         | 成就唯一标识                         |
| name         | TEXT     | NOT NULL            | 成就名称                           |
| description  | TEXT     | NOT NULL            | 成就描述                           |
| level        | INTEGER  | NOT NULL, DEFAULT 1 | 等级（1=Bronze, 2=Silver, 3=Gold） |
| icon_url     | TEXT     |                     | 图标 URL                         |
| badge_type   | TEXT     | NOT NULL            | 类别（learning、streak、speed 等）    |
| rule_json    | TEXT     | NOT NULL            | 触发规则（JSON）                     |
| nft_enabled  | BOOLEAN  | DEFAULT FALSE       | 是否生成 NFT                       |
| nft_metadata | TEXT     |                     | NFT 元数据模板（JSON）                |
| is_active    | BOOLEAN  | DEFAULT TRUE        | 是否启用                           |
| created_at   | DATETIME | NOT NULL            | 创建时间                           |
| updated_at   | DATETIME | NOT NULL            | 更新时间                           |

**user achievements**

| 字段名            | 类型       | 约束 / 默认值            | 说明            |
| -------------- | -------- | ------------------- | ------------- |
| id             | TEXT     | **PK**              | 主键            |
| user_id        | TEXT     | **NOT NULL**, INDEX | 用户 ID         |
| achievement_id | TEXT     | **NOT NULL**, INDEX | 成就 ID         |
| earned_at      | DATETIME | **NOT NULL**        | 获得成就的时间       |
| progress       | REAL     | DEFAULT 1.0         | 完成度 (0.0–1.0) |
| meta_json      | TEXT     |                     | 额外元数据         |
| notified_at    | DATETIME | NULLABLE            | 通知时间          |
| viewed_at      | DATETIME | NULLABLE            | 用户查看时间        |

**events**

| 列名          | 数据类型     | 约束              | 说明                                                |
| ----------- | -------- | --------------- | ------------------------------------------------- |
| id          | TEXT     | PRIMARY KEY     | 唯一事件标识                                            |
| user_id     | TEXT     | NOT NULL, INDEX | 触发事件的用户 ID                                        |
| event_type  | TEXT     | NOT NULL, INDEX | 事件类型，例如 `"question_answered"`、`"level_completed"` |
| level_id    | TEXT     | NULLABLE, INDEX | 关联关卡 ID（可选）                                       |
| question_id | TEXT     | NULLABLE, INDEX | 关联题目 ID（可选）                                       |
| data_json   | TEXT     | —               | 事件特定数据（JSON 字符串）                                  |
| created_at  | DATETIME | NOT NULL, INDEX | 事件创建时间（UTC）                                       |

**NET assets**

| 字段               | 类型      | 约束                               | 说明                              |
| ---------------- | ------- | -------------------------------- | ------------------------------- |
| id               | TEXT    | PK UUID                          | 资产记录主键                          |
| user_id          | TEXT    | FK → users(id) ON DELETE CASCADE | 拥有者用户                           |
| achievement_id   | TEXT    | FK → achievements(id)            | 对应成就（可选，集成时填入）                  |
| contract_address | TEXT    | NOT NULL                         | ERC‑721 合约地址                    |
| token_id         | TEXT    | NOT NULL                         | 铸造后返回的 NFT Token ID             |
| metadata_uri     | TEXT    | NOT NULL                         | NFT 元数据 URI（IPFS/CID 或 HTTP 链接） |
| mint_tx_hash     | TEXT    |                                  | 铸造交易哈希                          |
| status           | TEXT    | NOT NULL DEFAULT 'pending'       | 状态：`pending`、`minted`、`failed`  |
| created_at       | INTEGER | NOT NULL                         | 记录创建时间（Unix ms）                 |
| updated_at       | INTEGER | NOT NULL                         | 最近更新时间（Unix ms）                 |
