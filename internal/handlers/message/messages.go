package message

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	ssogrpc "simple-chat/internal/clients/sso/grpc"
	"simple-chat/internal/domain/dto"
	"simple-chat/internal/domain/models"
	"simple-chat/internal/handlers"
	"simple-chat/internal/lib/logger/sl"
	authMiddleware "simple-chat/internal/lib/middleware"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type MessageHandler struct {
	log            *slog.Logger
	messageService MessageService
	appID          int32
}

type MessageService interface {
	CreateMessage(ctx context.Context, message dto.Message) (int64, error)
	GetMessagesByChatID(ctx context.Context, chatID int64, limit int, offset int) ([]models.Message, error)
}

func NewMessageHandler(log *slog.Logger, messageService MessageService, appID int32) *MessageHandler {
	return &MessageHandler{
		log:            log,
		messageService: messageService,
		appID:          appID,
	}
}

func AddMessageHandler(log *slog.Logger, messageService MessageService, ssoClient *ssogrpc.Client, appID int32) func(chi.Router) {
	messageHandler := NewMessageHandler(log, messageService, appID)

	return func(r chi.Router) {
		r.Use(authMiddleware.Auth(log, ssoClient, appID))

		r.Post("/create", messageHandler.Create(context.Background()))
		r.Get("/{chat_id}", messageHandler.GetMessagesByChatID(context.Background()))
	}
}

func (h *MessageHandler) Create(ctx context.Context) http.HandlerFunc {
	const op = "handlers.message.Create"

	return func(w http.ResponseWriter, r *http.Request) {
		h.log = h.log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var message dto.MessageRequest
		err := json.NewDecoder(r.Body).Decode(&message)
		if err != nil {
			h.log.Error("failed to decode request body", sl.Err(err))
			handlers.ErrorResponse(w, r, 400, "bad request")
			return
		}
		if err := message.Validate(); err != nil {
			h.log.Error("failed to validate message", sl.Err(err))
			handlers.ErrorResponse(w, r, 422, err.Error())
			return
		}

		user, ok := r.Context().Value(authMiddleware.UserContextKey).(models.User)
		if !ok {
			h.log.Error("failed to get user")
			handlers.ErrorResponse(w, r, 401, "unauthorized")
			return
		}

		messageModel := dto.Message{
			ChatID:    message.ChatID,
			Sender:    user.UserID,
			Text:      message.Text,
			CreatedAt: time.Now().UTC(),
		}

		if err := message.Validate(); err != nil {
			h.log.Error("failed to validate message", sl.Err(err))
			handlers.ErrorResponse(w, r, 422, err.Error())
			return
		}

		messageID, err := h.messageService.CreateMessage(ctx, messageModel)
		if err != nil {
			h.log.Error("failed to create message", sl.Err(err))
			handlers.ErrorResponse(w, r, 500, "failed to create message")
			return
		}

		handlers.SuccessResponse(w, r, 200, map[string]any{
			"message":    "message successfully created",
			"message_id": messageID,
		})
	}
}

func (h *MessageHandler) GetMessagesByChatID(ctx context.Context) http.HandlerFunc {
	const op = "handlers.message.GetMessagesByChatID"

	return func(w http.ResponseWriter, r *http.Request) {
		h.log = h.log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
		if err != nil || limit <= 0 {
			limit = 10
		}
		offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
		if err != nil || offset < 0 {
			offset = 0
		}
		h.log.Debug("limit and offset from query", slog.Int("limit", limit), slog.Int("offset", offset))

		chatID, err := strconv.ParseInt(chi.URLParam(r, "chat_id"), 10, 64)
		if err != nil {
			h.log.Error("failed to parse chat id from url params", sl.Err(err))
			handlers.ErrorResponse(w, r, 400, "bad request")
			return
		}

		messages, err := h.messageService.GetMessagesByChatID(ctx, chatID, limit, offset)
		if err != nil {
			h.log.Error("failed to get messages by chat id", sl.Err(err))
			handlers.ErrorResponse(w, r, 500, "failed to get messages")
			return
		}
		if messages == nil {
			h.log.Error("messages not found")
			handlers.ErrorResponse(w, r, 404, "messages not found")
			return
		}

		handlers.SuccessResponse(w, r, 200, messages)
	}
}
