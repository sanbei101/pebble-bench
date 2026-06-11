package store

import (
	"context"
	"encoding/json/jsontext"
	"errors"

	"github.com/google/uuid"
)

var (
	ErrMessageNotFound = errors.New("message not found")
	ErrInvalidCursor   = errors.New("invalid message cursor")
)

type Message struct {
	ID        uuid.UUID      `json:"id"`
	RoomID    uuid.UUID      `json:"room_id"`
	SenderID  uuid.UUID      `json:"sender_id"`
	Content   jsontext.Value `json:"content"`
	CreatedAt int64          `json:"timestamp"`
}

type Cursor struct {
	CreatedAt int64
	MessageID uuid.UUID
}

type ListOptions struct {
	Limit  int
	Before *Cursor
}

type MessagePage struct {
	Messages   []Message
	NextCursor *Cursor
	HasMore    bool
}

type WriteOptions struct {
	Durable bool
}

type Repository interface {
	PutMessage(
		ctx context.Context,
		message *Message,
		options WriteOptions,
	) error

	PutMessages(
		ctx context.Context,
		messages []Message,
		options WriteOptions,
	) error

	GetMessage(
		ctx context.Context,
		messageID uuid.UUID,
	) (Message, error)

	MessageExists(
		ctx context.Context,
		messageID uuid.UUID,
	) (bool, error)

	ListRoomMessages(
		ctx context.Context,
		roomID uuid.UUID,
		options ListOptions,
	) (MessagePage, error)

	ListMessagesBySender(
		ctx context.Context,
		senderID uuid.UUID,
		options ListOptions,
	) (MessagePage, error)

	DeleteMessage(
		ctx context.Context,
		messageID uuid.UUID,
		options WriteOptions,
	) error

	DeleteMessages(
		ctx context.Context,
		messageIDs []uuid.UUID,
		options WriteOptions,
	) error

	DeleteRoomMessages(
		ctx context.Context,
		roomID uuid.UUID,
		options WriteOptions,
	) error

	SetReadCursor(
		ctx context.Context,
		userID uuid.UUID,
		roomID uuid.UUID,
		cursor Cursor,
		options WriteOptions,
	) error

	GetReadCursor(
		ctx context.Context,
		userID uuid.UUID,
		roomID uuid.UUID,
	) (Cursor, error)

	CountRoomMessages(
		ctx context.Context,
		roomID uuid.UUID,
	) (uint64, error)

	Close() error
}
