-- name: GetRoom :one
SELECT * FROM rooms
WHERE namespace = $1 AND id = $2
LIMIT 1;

-- name: InsertRoom :one
INSERT INTO rooms (
  namespace, id 
) VALUES (
  $1, $2
)
RETURNING *;

-- name: ListDeathCount :many
SELECT * FROM deaths
WHERE namespace = $1 AND room_id = $2;

-- name: IncrementDeathCount :exec
INSERT INTO deaths (
  namespace, room_id, user_id, count
) VALUES (
  $1, $2, $3, 1
) ON CONFLICT (namespace, room_id, user_id)
  DO UPDATE SET count = deaths.count + 1
  WHERE deaths.namespace = $1 AND deaths.room_id = $2 AND deaths.user_id = $3;