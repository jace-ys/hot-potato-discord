CREATE TABLE IF NOT EXISTS rooms (
  namespace TEXT NOT NULL,
  id TEXT NOT NULL,
  created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (namespace, id)
)