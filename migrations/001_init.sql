-- 学科表
CREATE TABLE IF NOT EXISTS subjects (
  id                TEXT    PRIMARY KEY,
  name              TEXT    NOT NULL,
  description       TEXT,
  created_at        INTEGER NOT NULL,
  updated_at        INTEGER NOT NULL
);

-- 论文表
CREATE TABLE IF NOT EXISTS papers (
  id                   TEXT    PRIMARY KEY,
  subject_id           TEXT    NOT NULL,
  title                TEXT    NOT NULL,
  paper_author         TEXT    NOT NULL,
  paper_pub_ym         TEXT    NOT NULL,
  paper_citation_count TEXT    NOT NULL,
  created_at           INTEGER,
  updated_at           INTEGER,
  FOREIGN KEY(subject_id) REFERENCES subjects(id) ON DELETE CASCADE
);

-- 关卡表
CREATE TABLE IF NOT EXISTS levels (
  id             TEXT    PRIMARY KEY,
  paper_id       TEXT    NOT NULL UNIQUE,
  name           TEXT    NOT NULL,
  pass_condition TEXT    NOT NULL,
  meta_json      TEXT,
  x              INTEGER NOT NULL,
  y              INTEGER NOT NULL,
  created_at     INTEGER,
  updated_at     INTEGER,
  FOREIGN KEY(paper_id) REFERENCES papers(id) ON DELETE CASCADE
);

-- 题目表
CREATE TABLE IF NOT EXISTS questions (
  id           TEXT    PRIMARY KEY,
  level_id     TEXT    NOT NULL,
  stem         TEXT    NOT NULL,
  content_json TEXT    NOT NULL,
  answer_json  TEXT    NOT NULL,
  score        INTEGER NOT NULL,
  created_by   TEXT,
  created_at   INTEGER,
  FOREIGN KEY(level_id) REFERENCES levels(id) ON DELETE CASCADE
);

-- 路线图节点表
CREATE TABLE IF NOT EXISTS roadmap_nodes (
  id         TEXT    PRIMARY KEY,
  subject_id TEXT    NOT NULL,
  level_id   TEXT    NOT NULL,
  parent_id  TEXT,
  sort_order INTEGER NOT NULL DEFAULT 1,
  path       TEXT    NOT NULL,
  depth      INTEGER NOT NULL,
  FOREIGN KEY(subject_id) REFERENCES subjects(id),
  FOREIGN KEY(level_id)   REFERENCES levels(id),
  FOREIGN KEY(parent_id)  REFERENCES roadmap_nodes(id)
);

-- 用户表
CREATE TABLE IF NOT EXISTS users (
  id           TEXT    PRIMARY KEY,
  email        TEXT    NOT NULL UNIQUE,
  password_hash TEXT   NOT NULL,
  display_name TEXT,
  avatar_url   TEXT,
  created_at   INTEGER NOT NULL,
  updated_at   INTEGER NOT NULL
);

-- Refresh Token 表
CREATE TABLE IF NOT EXISTS refresh_tokens (
  token       TEXT    PRIMARY KEY,
  user_id     TEXT    NOT NULL,
  expires_at  INTEGER NOT NULL,
  FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- 用户关卡进度表
CREATE TABLE IF NOT EXISTS user_progress (
  user_id         TEXT    NOT NULL,
  level_id        TEXT    NOT NULL,
  status          INTEGER NOT NULL,
  score           INTEGER,
  stars           INTEGER,
  last_attempt_at INTEGER,
  PRIMARY KEY(user_id, level_id),
  FOREIGN KEY(user_id)  REFERENCES users(id),
  FOREIGN KEY(level_id) REFERENCES levels(id)
);

-- 用户每日答题统计表
CREATE TABLE IF NOT EXISTS user_attempts (
  stat_date                   TEXT    NOT NULL,    -- 格式 YYYY-MM-DD
  user_id                     TEXT    NOT NULL,
  attempts_total              INTEGER DEFAULT 0,
  attempts_correct            INTEGER DEFAULT 0,
  attempts_first_try_correct  INTEGER DEFAULT 0,
  correct_rate                REAL,
  first_try_correct_rate      REAL,
  giveup_count                INTEGER DEFAULT 0,
  skip_rate                   REAL,
  avg_duration_ms             INTEGER,
  total_time_ms               INTEGER,
  sessions_count              INTEGER,
  streak_days                 INTEGER,
  review_due_count            INTEGER,
  retention_score             REAL,
  peak_hour                   INTEGER,
  updated_at                  INTEGER NOT NULL,
  PRIMARY KEY(stat_date, user_id),
  FOREIGN KEY(user_id) REFERENCES users(id)
);
