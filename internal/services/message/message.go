package message

import (
	"context"
	"errors"
	"log/slog"
	"simple-chat/internal/domain/dto"
	"simple-chat/internal/domain/models"
	"simple-chat/internal/lib/logger/sl"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MessageService struct {
	log        *slog.Logger
	messagesDB MessagesDB
	pool       *pgxpool.Pool
}

type MessagesDB interface {
	CreateMessage(ctx context.Context, tx pgx.Tx, message dto.Message) (int64, error)
	GetMessageByID(ctx context.Context, tx pgx.Tx, messageID int64) (models.Message, error)
	GetMessagesByChatID(ctx context.Context, tx pgx.Tx, chatID int64, limit int, offset int) ([]models.Message, error)
	GetListMessagesByID(ctx context.Context, tx pgx.Tx, messagesID []int64) ([]models.Message, error)
}

func NewMessageServices(log *slog.Logger, messagesDB MessagesDB, pool *pgxpool.Pool) *MessageService {
	return &MessageService{
		log:        log,
		messagesDB: messagesDB,
		pool:       pool,
	}
}

func (s *MessageService) CreateMessage(ctx context.Context, message dto.Message) (messageID int64, err error) {
	const op = "message.service.CreateMessage"

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		s.log.Error("failed to start transaction", sl.OpErr(op, err))
		return 0, err
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
			return
		}
		if err := tx.Commit(ctx); err != nil {
			s.log.Error("failed to commit transaction", sl.OpErr(op, err))
			return
		}
	}()

	messageID, err = s.messagesDB.CreateMessage(ctx, tx, message)
	if err != nil {
		s.log.Error("failed to create message", sl.OpErr(op, err))
		return 0, err
	}
	if messageID == 0 {
		err = errors.New("message id is empty")
		s.log.Error("failed to create message", sl.OpErr(op, err))
		return 0, err
	}

	return messageID, nil
}

func (s *MessageService) GetMessageByID(ctx context.Context, messageID int64) (models.Message, error) {
	const op = "message.service.GetMessageByID"

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		s.log.Error("failed to start transaction", sl.OpErr(op, err))
		return models.Message{}, err
	}
	defer tx.Rollback(ctx)

	message, err := s.messagesDB.GetMessageByID(ctx, tx, messageID)
	if err != nil {
		s.log.Error("failed to get message", sl.OpErr(op, err))
		return models.Message{}, err
	}
	if message == (models.Message{}) {
		err = errors.New("message is empty")
		s.log.Error("failed to get message", sl.OpErr(op, err))
		return models.Message{}, err
	}

	return message, nil
}

func (s *MessageService) GetMessagesByChatID(ctx context.Context, chatID int64, limit int, offset int) ([]models.Message, error) {
	const op = "message.service.GetMessagesByChatID"

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		s.log.Error("failed to start transaction", sl.OpErr(op, err))
		return nil, err
	}
	defer tx.Rollback(ctx)

	messages, err := s.messagesDB.GetMessagesByChatID(ctx, tx, chatID, limit, offset)
	if err != nil {
		s.log.Error("failed to get messages", sl.OpErr(op, err))
		return nil, err
	}
	if len(messages) == 0 {
		s.log.Error("messages not found", slog.String("op", op))
		return nil, nil
	}

	return messages, nil
}

func (s *MessageService) GetListMessagesByID(ctx context.Context, chatID []int64) ([]models.Message, error) {
	const op = "message.service.GetListMessagesByID"

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		s.log.Error("failed to start transaction", sl.OpErr(op, err))
		return nil, err
	}
	defer tx.Rollback(ctx)

	messages, err := s.messagesDB.GetListMessagesByID(ctx, tx, chatID)
	if err != nil {
		s.log.Error("failed to get messages", sl.OpErr(op, err))
		return nil, err
	}
	if len(messages) == 0 {
		s.log.Info("messages list is empty")
		return nil, nil
	}

	return messages, nil
}
