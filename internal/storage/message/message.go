package message

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"simple-chat/internal/domain/dto"
	"simple-chat/internal/domain/models"
	"simple-chat/internal/lib/logger/sl"
	"simple-chat/internal/lib/storage/query"

	"github.com/jackc/pgx/v5"
)

type MessageDB struct {
	log *slog.Logger
}

func NewMessageDB(log *slog.Logger) *MessageDB {
	return &MessageDB{
		log: log,
	}
}

const (
	messageTable = "message"
)

var (
	ErrMessageNotFound  = errors.New("message not found")
	ErrMessagesNotFound = errors.New("messages not found")
)

func (m *MessageDB) CreateMessage(ctx context.Context, tx pgx.Tx, message dto.Message) (int64, error) {
	const op = "storage.message.CreateMessage"

	q := fmt.Sprintf(`
        INSERT INTO %s 
            (chat_id, sender, text, created_at)
        VALUES 
            ($1, $2, $3, $4)
		RETURNING id;
	`, messageTable)

	m.log.Debug("create message query:", slog.String("query", query.QueryToString(q)))

	var messageID int64
	err := tx.QueryRow(ctx, q, message.ChatID, message.Sender, message.Text, message.CreatedAt).Scan(&messageID)
	if err != nil {
		m.log.Error("faield to create message", sl.OpErr(op, err))
		return 0, err
	}

	return messageID, nil
}

func (m *MessageDB) GetMessageByID(ctx context.Context, tx pgx.Tx, messageID int64) (models.Message, error) {
	const op = "storage.message.GetMessageByID"

	q := fmt.Sprintf(`
        SELECT 
            id, chat_id, sender, text, created_at 
        FROM %s 
        WHERE id = $1;
	`, messageTable)

	m.log.Debug("get message by id query:", slog.String("query", query.QueryToString(q)))

	var message models.Message
	err := tx.QueryRow(ctx, q, messageID).Scan(&message.ID, &message.ChatID, &message.Sender, &message.Text, &message.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return models.Message{}, ErrMessageNotFound
		}
		m.log.Error("faield to get message by id", sl.OpErr(op, err))
		return models.Message{}, err
	}

	return message, nil
}

func (m *MessageDB) GetMessagesByChatID(ctx context.Context, tx pgx.Tx, chatID int64, limit int, offset int) ([]models.Message, error) {
	const op = "storage.message.GetMessagesByChatID"

	q := fmt.Sprintf(`
        SELECT 
            id, chat_id, sender, text, created_at 
        FROM %s 
        WHERE chat_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3;
	`, messageTable)

	m.log.Debug("get messages by chat id query:", slog.String("query", query.QueryToString(q)))

	var messages []models.Message

	rows, err := tx.Query(ctx, q, chatID, limit, offset)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrMessagesNotFound
		}
		m.log.Error("faield to get messages by chat id", sl.OpErr(op, err))
		return nil, err
	}

	for rows.Next() {
		var message models.Message
		err := rows.Scan(&message.ID, &message.ChatID, &message.Sender, &message.Text, &message.CreatedAt)
		if err != nil {
			m.log.Error("faield to scan message", sl.OpErr(op, err))
			return nil, err
		}

		messages = append(messages, message)
	}

	if rows.Err() != nil {
		m.log.Error("faield to get messages by chat id", sl.OpErr(op, err))
		return nil, err
	}

	return messages, nil
}

func (m *MessageDB) GetListMessagesByID(ctx context.Context, tx pgx.Tx, messagesID []int64) ([]models.Message, error) {
	const op = "storage.message.GetListMessagesByID"

	q := fmt.Sprintf(`
        SELECT 
            id, chat_id, sender, text, created_at 
        FROM %s 
        WHERE chat_id ON $1;
	`, messageTable)

	m.log.Debug("get list messages by id query:", slog.String("query", query.QueryToString(q)))

	var messages []models.Message
	rows, err := tx.Query(ctx, q, messagesID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrMessagesNotFound
		}
		m.log.Error("faield to get list messages by id", sl.OpErr(op, err))
		return nil, err
	}

	for rows.Next() {
		var message models.Message
		err := rows.Scan(&message.ID, &message.ChatID, &message.Sender, &message.Text, &message.CreatedAt)
		if err != nil {
			m.log.Error("faield to scan message", sl.OpErr(op, err))
			return nil, err
		}

		messages = append(messages, message)
	}

	if rows.Err() != nil {
		m.log.Error("faield to get list messages by id", sl.OpErr(op, err))
		return nil, err
	}

	return messages, nil
}
