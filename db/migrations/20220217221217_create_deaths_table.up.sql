CREATE TABLE IF NOT EXISTS deaths (
  namespace TEXT NOT NULL,
  room_id TEXT NOT NULL,
  user_id TEXT NOT NULL,
  count INT DEFAULT 0,
  PRIMARY KEY (namespace, room_id, user_id),
  FOREIGN KEY (namespace, room_id) references rooms (namespace, id) ON DELETE CASCADE
)