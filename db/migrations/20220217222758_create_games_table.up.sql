CREATE TABLE IF NOT EXISTS games (
  namespace TEXT NOT NULL,
  room_id TEXT NOT NULL,
  channel_id TEXT NOT NULL,
  potato_kind TEXT NOT NULL,
  heat_level INT NOT NULL,
  holder_user_id TEXT NOT NULL,
  turns INT NOT NULL,
  finished BOOLEAN NOT NULL,
  created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (namespace, room_id, channel_id),
  FOREIGN KEY (namespace, room_id) REFERENCES rooms (namespace, id) ON DELETE CASCADE
)