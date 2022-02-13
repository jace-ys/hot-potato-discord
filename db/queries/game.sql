-- name: GetGame :one
SELECT * FROM games
WHERE namespace = $1 AND channel_id = $2
LIMIT 1;

-- name: UpsertGame :one
INSERT INTO games (
  namespace, room_id, channel_id, potato_kind, heat_level, holder_user_id, turns, finished
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8
) ON CONFLICT (namespace, room_id, channel_id)
  DO UPDATE SET potato_kind = $4, heat_level = $5, holder_user_id = $6, turns = $7, finished = $8
  WHERE games.namespace = $1 AND games.room_id = $2 AND games.channel_id = $3
RETURNING *;

-- name: UpdateHolder :one
UPDATE games
SET holder_user_id = $3
WHERE namespace = $1 AND channel_id = $2
RETURNING *;

-- name: IncrementTurns :one
UPDATE games
SET turns = turns + 1
WHERE namespace = $1 AND channel_id = $2
RETURNING *;

-- name: IncreaseHeatLevel :one
UPDATE games
SET heat_level = heat_level + sqlc.arg(amount)::int
WHERE namespace = $1 AND channel_id = $2
RETURNING *;

-- name: EndGame :one
UPDATE games
SET finished = true
WHERE namespace = $1 AND channel_id = $2
RETURNING *;