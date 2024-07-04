package chat

import (
	"context"
	"errors"
	"log/slog"
	"simple-chat/internal/domain/dto"
	"simple-chat/internal/domain/models"
	"simple-chat/internal/lib/logger/sl"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ChatService struct {
	log    *slog.Logger
	chatDB ChatDB
	pool   *pgxpool.Pool
}

type ChatDB interface {
	CreateChat(ctx context.Context, tx pgx.Tx, chat *dto.Chat) (int64, error)
	GetChatByID(ctx context.Context, tx pgx.Tx, chatID int64) (models.Chat, error)
	GetUserChats(ctx context.Context, tx pgx.Tx, userID int64, limit int, offset int) ([]models.Chat, error)
	UpdateChatMessage(ctx context.Context, tx pgx.Tx, chatID int64, message string, updatedAt time.Time) error
}

func NewChatService(log *slog.Logger, chatDB ChatDB, pool *pgxpool.Pool) *ChatService {
	return &ChatService{
		log:    log,
		chatDB: chatDB,
		pool:   pool,
	}
}

func (s *ChatService) CreateChat(ctx context.Context, chat *dto.Chat) (chatID int64, err error) {
	const op = "chat.service.CreateChat"

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

	chatID, err = s.chatDB.CreateChat(ctx, tx, chat)
	if err != nil {
		s.log.Error("failed to create chat", sl.OpErr(op, err))
		return 0, err
	}

	if chatID == 0 {
		s.log.Error("failed to create chat", sl.OpErr(op, errors.New("chat id is empty")))
		err = errors.New("chat id is empty")
		return 0, err
	}

	return chatID, nil
}

func (s *ChatService) GetChatByID(ctx context.Context, chatID int64) (models.Chat, error) {
	const op = "chat.service.GetChatByID"

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		s.log.Error("failed to start transaction", sl.OpErr(op, err))
		return models.Chat{}, err
	}
	defer tx.Rollback(ctx)

	chat, err := s.chatDB.GetChatByID(ctx, tx, chatID)
	if err != nil {
		s.log.Error("failed to get chat", sl.OpErr(op, err))
		return models.Chat{}, err
	}
	if chat == (models.Chat{}) {
		s.log.Error("failed to get chat", sl.OpErr(op, errors.New("chat model is empty")))
		err = errors.New("chat model is empty")
		return models.Chat{}, err
	}

	return chat, nil
}

func (s *ChatService) GetUserChats(ctx context.Context, userID int64, limit int, offset int) ([]models.Chat, error) {
	const op = "chat.service.GetUserChats"

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		s.log.Error("failed to start transaction", sl.OpErr(op, err))
		return nil, err
	}
	defer tx.Rollback(ctx)

	chats, err := s.chatDB.GetUserChats(ctx, tx, userID, limit, offset)
	if err != nil {
		s.log.Error("failed to get user chats", sl.OpErr(op, err))
		return nil, err
	}
	if len(chats) == 0 {
		s.log.Error("chats not found", slog.String("op", op))
		return nil, nil
	}

	return chats, nil
}

func (s *ChatService) UpdateChatMessage(ctx context.Context, chatID int64, message string, updatedAt time.Time) (err error) {
	const op = "chat.service.UpdateChatMessage"

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		s.log.Error("failed to start transaction", sl.OpErr(op, err))
		return err
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

	err = s.chatDB.UpdateChatMessage(ctx, tx, chatID, message, updatedAt)
	if err != nil {
		s.log.Error("failed to update chat message", sl.OpErr(op, err))
		return err
	}

	return nil
}
