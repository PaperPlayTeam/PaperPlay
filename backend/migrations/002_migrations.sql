-- Enable FK constraints
PRAGMA foreign_keys = ON;

-- ========== UP ==========
BEGIN TRANSACTION;

/* ---------- user_achievements ---------- */
CREATE TABLE IF NOT EXISTS user_achievements (
  id            TEXT    PRIMARY KEY,
  user_id       TEXT    NOT NULL,
  achievement_id TEXT   NOT NULL,
  earned_at     DATETIME NOT NULL,
  progress      REAL    NOT NULL DEFAULT 1.0 CHECK (progress BETWEEN 0.0 AND 1.0),
  meta_json     TEXT,
  notified_at   DATETIME,
  viewed_at     DATETIME,
  FOREIGN KEY (user_id)        REFERENCES users(id)        ON DELETE CASCADE,
  FOREIGN KEY (achievement_id) REFERENCES achievements(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_user_achievements_user   ON user_achievements(user_id);
CREATE INDEX IF NOT EXISTS idx_user_achievements_achv   ON user_achievements(achievement_id);

/* ---------- events ---------- */
CREATE TABLE IF NOT EXISTS events (
  id          TEXT     PRIMARY KEY,
  user_id     TEXT     NOT NULL,
  event_type  TEXT     NOT NULL,
  level_id    TEXT,
  question_id TEXT,
  data_json   TEXT,
  created_at  DATETIME NOT NULL,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_events_user_id     ON events(user_id);
CREATE INDEX IF NOT EXISTS idx_events_event_type  ON events(event_type);
CREATE INDEX IF NOT EXISTS idx_events_level_id    ON events(level_id);
CREATE INDEX IF NOT EXISTS idx_events_question_id ON events(question_id);
CREATE INDEX IF NOT EXISTS idx_events_created_at  ON events(created_at);

/* ---------- nft_assets ---------- */
CREATE TABLE IF NOT EXISTS nft_assets (
  id               TEXT    PRIMARY KEY,
  user_id          TEXT    NOT NULL,
  achievement_id   TEXT,
  contract_address TEXT    NOT NULL,
  token_id         TEXT    NOT NULL,
  metadata_uri     TEXT    NOT NULL,
  mint_tx_hash     TEXT,
  status           TEXT    NOT NULL DEFAULT 'pending'
                   CHECK (status IN ('pending','minted','failed')),
  created_at       INTEGER NOT NULL,           -- Unix ms
  updated_at       INTEGER NOT NULL,
  FOREIGN KEY (user_id)        REFERENCES users(id)        ON DELETE CASCADE,
  FOREIGN KEY (achievement_id) REFERENCES achievements(id)
);
CREATE INDEX IF NOT EXISTS idx_nft_assets_user        ON nft_assets(user_id);
CREATE INDEX IF NOT EXISTS idx_nft_assets_achievement ON nft_assets(achievement_id);

COMMIT;

-- ========== DOWN ==========
BEGIN TRANSACTION;
DROP TABLE IF EXISTS nft_assets;
DROP TABLE IF EXISTS events;
DROP TABLE IF EXISTS user_achievements;
COMMIT;
