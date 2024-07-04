package chat

import (
	"context"
	"fmt"
	"log/slog"
	"simple-chat/internal/domain/dto"
	"simple-chat/internal/domain/models"
	"simple-chat/internal/lib/logger/sl"
	"simple-chat/internal/lib/storage/query"
	"time"

	"github.com/jackc/pgx/v5"
)

type ChatDB struct {
	log *slog.Logger
}

func NewChatDB(log *slog.Logger) *ChatDB {
	return &ChatDB{
		log: log,
	}
}

const (
	chatTable = "chat"
)

var (
	ErrChatNotFound  = fmt.Errorf("chat not found")
	ErrChatsNotFound = fmt.Errorf("chats not found")
)

func (c *ChatDB) CreateChat(ctx context.Context, tx pgx.Tx, chat *dto.Chat) (int64, error) {
	const op = "storage.chat.CreateChat"

	q := fmt.Sprintf(`
		INSERT INTO %s 
			(first_user_id, second_user_id, updated_at) 
		VALUES ($1, $2, $3)
		RETURNING id;
	`, chatTable)

	c.log.Debug("create chat query:", slog.String("query", query.QueryToString(q)))

	var chatID int64

	err := tx.QueryRow(ctx, q, chat.FirstUserID, chat.SecondUserID, chat.UpdatedAt).Scan(&chatID)
	if err != nil {
		c.log.Error("faield to create chat", sl.OpErr(op, err))
		return 0, err
	}

	return chatID, nil
}

func (c *ChatDB) UpdateChatMessage(ctx context.Context, tx pgx.Tx, chatID int64, message string, updatedAt time.Time) error {
	const op = "storage.chat.UpdateChatMessage"

	q := fmt.Sprintf(`
        UPDATE %s 
        SET last_message = $1, updated_at = $2
        WHERE id = $3;
    `, chatTable)

	c.log.Debug("update chat message query:", slog.String("query", query.QueryToString(q)))

	if _, err := tx.Exec(ctx, q, message, updatedAt, chatID); err != nil {
		c.log.Error("faield to update chat message", sl.OpErr(op, err))
		return err
	}

	return nil
}
func (c *ChatDB) GetChatByID(ctx context.Context, tx pgx.Tx, chatID int64) (models.Chat, error) {
	const op = "storage.chat.GetChatByID"

	q := fmt.Sprintf(`
        SELECT id, first_user_id, second_user_id, COALESCE(last_message, '') AS last_message, updated_at
        FROM %s
        WHERE id = $1;
    `, chatTable)

	c.log.Debug("get chat by id query:", slog.String("query", query.QueryToString(q)))

	var chat models.Chat

	err := tx.QueryRow(ctx, q, chatID).Scan(&chat.ID, &chat.FirstUserID, &chat.SecondUserID, &chat.LastMessage, &chat.UpdatedAt)

	c.log.Debug("chat by id:",
		slog.String("op", op),
		slog.Int64("chatID", chatID),
		slog.Any("chat", chat),
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return models.Chat{}, ErrChatNotFound
		}
		c.log.Error("faield to get chat by id", sl.OpErr(op, err))
		return models.Chat{}, err
	}

	return chat, nil
}
func (c *ChatDB) GetUserChats(ctx context.Context, tx pgx.Tx, userID int64, limit int, offset int) ([]models.Chat, error) {
	const op = "storage.chat.GetUserChats"

	q := fmt.Sprintf(`
        SELECT id, first_user_id, second_user_id, COALESCE(last_message, '') AS last_message, updated_at
        FROM %s
        WHERE first_user_id = $1 OR second_user_id = $1
        LIMIT $2 OFFSET $3;
	`, chatTable)

	c.log.Debug("get user chats query:", slog.String("query", query.QueryToString(q)))

	var chats []models.Chat

	rows, err := tx.Query(ctx, q, userID, limit, offset)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrChatsNotFound
		}
		c.log.Error("faield to get user chats", sl.OpErr(op, err))
		return nil, err
	}

	for rows.Next() {
		var chat models.Chat

		err := rows.Scan(&chat.ID, &chat.FirstUserID, &chat.SecondUserID, &chat.LastMessage, &chat.UpdatedAt)
		if err != nil {
			c.log.Error("faield to get user chats", sl.OpErr(op, err))
			return nil, err
		}

		chats = append(chats, chat)
	}

	if rows.Err() != nil {
		c.log.Error("faield to get user chats", sl.OpErr(op, err))
		return nil, err
	}

	return chats, nil
}
