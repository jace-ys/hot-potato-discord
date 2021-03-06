// Code generated by sqlc. DO NOT EDIT.
// source: room.sql

package store

import (
	"context"
)

const getRoom = `-- name: GetRoom :one
SELECT namespace, id, created_at FROM rooms
WHERE namespace = $1 AND id = $2
LIMIT 1
`

type GetRoomParams struct {
	Namespace string
	ID        string
}

func (q *Queries) GetRoom(ctx context.Context, arg GetRoomParams) (Room, error) {
	row := q.db.QueryRowContext(ctx, getRoom, arg.Namespace, arg.ID)
	var i Room
	err := row.Scan(&i.Namespace, &i.ID, &i.CreatedAt)
	return i, err
}

const incrementDeathCount = `-- name: IncrementDeathCount :exec
INSERT INTO deaths (
  namespace, room_id, user_id, count
) VALUES (
  $1, $2, $3, 1
) ON CONFLICT (namespace, room_id, user_id)
  DO UPDATE SET count = deaths.count + 1
  WHERE deaths.namespace = $1 AND deaths.room_id = $2 AND deaths.user_id = $3
`

type IncrementDeathCountParams struct {
	Namespace string
	RoomID    string
	UserID    string
}

func (q *Queries) IncrementDeathCount(ctx context.Context, arg IncrementDeathCountParams) error {
	_, err := q.db.ExecContext(ctx, incrementDeathCount, arg.Namespace, arg.RoomID, arg.UserID)
	return err
}

const insertRoom = `-- name: InsertRoom :one
INSERT INTO rooms (
  namespace, id 
) VALUES (
  $1, $2
)
RETURNING namespace, id, created_at
`

type InsertRoomParams struct {
	Namespace string
	ID        string
}

func (q *Queries) InsertRoom(ctx context.Context, arg InsertRoomParams) (Room, error) {
	row := q.db.QueryRowContext(ctx, insertRoom, arg.Namespace, arg.ID)
	var i Room
	err := row.Scan(&i.Namespace, &i.ID, &i.CreatedAt)
	return i, err
}

const listDeathCount = `-- name: ListDeathCount :many
SELECT namespace, room_id, user_id, count FROM deaths
WHERE namespace = $1 AND room_id = $2
`

type ListDeathCountParams struct {
	Namespace string
	RoomID    string
}

func (q *Queries) ListDeathCount(ctx context.Context, arg ListDeathCountParams) ([]Death, error) {
	rows, err := q.db.QueryContext(ctx, listDeathCount, arg.Namespace, arg.RoomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Death
	for rows.Next() {
		var i Death
		if err := rows.Scan(
			&i.Namespace,
			&i.RoomID,
			&i.UserID,
			&i.Count,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
