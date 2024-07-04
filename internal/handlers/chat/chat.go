package chat

import (
	"context"
	"errors"
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
	"github.com/go-chi/render"
	"github.com/gorilla/websocket"
)

type ChatHandler struct {
	log            *slog.Logger
	chatService    ChatService
	messageService MessageService
	appID          int32
}

type ChatService interface {
	CreateChat(ctx context.Context, chat *dto.Chat) (chatID int64, err error)
	GetChatByID(ctx context.Context, chatID int64) (models.Chat, error)
	GetUserChats(ctx context.Context, userID int64, limit int, offset int) ([]models.Chat, error)
	UpdateChatMessage(ctx context.Context, chatID int64, message string, updatedAt time.Time) (err error)
}

type MessageService interface {
	CreateMessage(ctx context.Context, message dto.Message) (int64, error)
}

func NewChatHandler(log *slog.Logger, chatService ChatService, messageService MessageService, appID int32) *ChatHandler {
	return &ChatHandler{
		log:            log,
		chatService:    chatService,
		messageService: messageService,
		appID:          appID,
	}
}

func AddChatHandler(log *slog.Logger, chatService ChatService, messageService MessageService, ssoClient *ssogrpc.Client, appID int32) func(r chi.Router) {
	chatHandler := NewChatHandler(log, chatService, messageService, appID)

	return func(r chi.Router) {
		r.Use(authMiddleware.Auth(log, ssoClient, chatHandler.appID))

		r.Post("/create", chatHandler.CreateChat(context.Background()))
		r.Get("/{chat_id}", chatHandler.GetChatByID(context.Background()))
		r.Get("/list", chatHandler.GetUserChats(context.Background()))

		r.Get("/ws/{chat_id}", chatHandler.ChatWebsocket(context.Background()))
	}
}

func (h *ChatHandler) CreateChat(ctx context.Context) http.HandlerFunc {
	const op = "handlers.chat.CreateChat"

	return func(w http.ResponseWriter, r *http.Request) {
		h.log = h.log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req dto.CreateChatRequest
		if err := render.Decode(r, &req); err != nil {
			h.log.Error("failed to decode request", sl.Err(err))
			handlers.ErrorResponse(w, r, 400, "bad request")
			return
		}
		if err := req.Validate(); err != nil {
			h.log.Error("failed to validate request", sl.Err(err))
			handlers.ErrorResponse(w, r, 422, err.Error())
			return
		}

		user, ok := r.Context().Value(authMiddleware.UserContextKey).(models.User)
		h.log.Debug("user from context",
			slog.Any("user", r.Context().Value(authMiddleware.UserContextKey)),
			slog.Any("user_model", user),
		)

		if !ok {
			h.log.Error("failed to get user")
			handlers.ErrorResponse(w, r, 401, "unauthorized")
			return
		}

		if req.FirstUserID != user.UserID && req.SecondUserID != user.UserID {
			h.log.Error("you cannot create a chat room that you are not a member of")
			handlers.ErrorResponse(w, r, 400, "you cannot create a chat room that you are not a member of")
			return
		}

		chatModel := &dto.Chat{
			FirstUserID:  req.FirstUserID,
			SecondUserID: req.SecondUserID,
			UpdatedAt:    time.Now().UTC(),
		}

		chatID, err := h.chatService.CreateChat(ctx, chatModel)
		if err != nil {
			h.log.Error("failed to create chat", sl.Err(err))
			handlers.ErrorResponse(w, r, 500, "failed to create chat")
			return
		}

		handlers.SuccessResponse(w, r, 200, map[string]any{
			"message": "new chat successfully created",
			"chat_id": chatID,
		})
	}
}

func (h *ChatHandler) GetChatByID(ctx context.Context) http.HandlerFunc {
	const op = "handlers.chat.GetChatByID"

	return func(w http.ResponseWriter, r *http.Request) {
		h.log = h.log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		chatID, err := strconv.ParseInt(chi.URLParam(r, "chat_id"), 10, 64)
		h.log.Debug("chat id from url params", slog.Any("chat_id", chatID))
		if err != nil {
			h.log.Error("failed to parse chat_id", sl.Err(err))
			handlers.ErrorResponse(w, r, 400, "bad request")
			return
		}

		chat, err := h.chatService.GetChatByID(ctx, chatID)
		if err != nil {
			h.log.Error("failed to get chat", sl.Err(err))
			handlers.ErrorResponse(w, r, 404, "chat not found")
			return
		}

		handlers.SuccessResponse(w, r, 200, chat)
	}
}

func (h *ChatHandler) GetUserChats(ctx context.Context) http.HandlerFunc {
	const op = "handlers.chat.GetUserChats"

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

		user, ok := r.Context().Value(authMiddleware.UserContextKey).(models.User)
		if !ok {
			h.log.Error("failed to get user")
			handlers.ErrorResponse(w, r, 401, "unauthorized")
			return
		}

		chats, err := h.chatService.GetUserChats(ctx, user.UserID, limit, offset)
		if err != nil {
			h.log.Error("failed to get user chats")
			handlers.ErrorResponse(w, r, 500, "failed to get user chats")
			return
		}
		if chats == nil {
			handlers.ErrorResponse(w, r, 404, "chats not found")
			return
		}

		handlers.SuccessResponse(w, r, 200, chats)
	}
}

type ChatRoom struct {
	chatID    int32
	users     []*websocket.Conn
	broadcast chan models.Message
	done      chan bool
}

var rooms = make(map[int32]*ChatRoom)

func connectToChatRoom(log *slog.Logger, chatID int32, ws *websocket.Conn) *ChatRoom {
	room, ok := rooms[chatID]
	if !ok {
		room = &ChatRoom{
			chatID:    chatID,
			users:     make([]*websocket.Conn, 0),
			broadcast: make(chan models.Message),
			done:      make(chan bool),
		}
		rooms[chatID] = room
	}
	go broadcast(log, room)
	room.users = append(room.users, ws)
	return room
}

func broadcast(log *slog.Logger, room *ChatRoom) {
	for {
		select {
		case msg := <-room.broadcast:
			for _, user := range room.users {
				go func(ws *websocket.Conn) {
					err := ws.WriteJSON(msg)
					if err != nil {
						log.Error("failed to write message to websocket", sl.Err(err))
					}
				}(user)
			}
		case <-room.done:
			return
		}
	}
}

func findWebsocketIndex(log *slog.Logger, ws *websocket.Conn, users []*websocket.Conn) (int, error) {
	const op = "handlers.chat.findWebsocketIndex"

	for idx, conn := range users {
		if conn == ws {
			return idx, nil
		}
	}
	log.Error("failed to find websocket in chat room", slog.String("op", op))
	return 0, errors.New("websocket not found in chat room")
}

func disconnectFromChatRoom(log *slog.Logger, chatID int32, ws *websocket.Conn) {
	const op = "handlers.chat.disconnectFromChatRoom"

	defer ws.Close()

	room, ok := rooms[chatID]
	if ok {
		if len(room.users) == 1 {
			room.done <- true
			delete(rooms, chatID)
			return
		}
		wsIdx, err := findWebsocketIndex(log, ws, room.users)
		if err != nil {
			log.Error("failed to find websocket in chat room", sl.OpErr(op, err))
			return
		}
		room.users = append(room.users[:wsIdx], room.users[wsIdx+1:]...)
		return
	}
}

func (h *ChatHandler) ChatWebsocket(ctx context.Context) http.HandlerFunc {
	const op = "handlers.chat.ChatWebsocket"

	upgrader := websocket.Upgrader{
		ReadBufferSize:    1024,
		WriteBufferSize:   1024,
		EnableCompression: true,
	}

	return func(w http.ResponseWriter, r *http.Request) {
		h.log = h.log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		chatID, err := strconv.ParseInt(chi.URLParam(r, "chat_id"), 10, 32)
		if err != nil {
			h.log.Error("failed to parse chat_id", sl.Err(err))
			handlers.ErrorResponse(w, r, 400, "bad request")
			return
		}

		user, ok := r.Context().Value(authMiddleware.UserContextKey).(models.User)
		if !ok {
			h.log.Error("failed to get user")
			handlers.ErrorResponse(w, r, 401, "unauthorized")
			return
		}

		h.log.Debug("new connection to chat websocket")
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			h.log.Error("failed to upgrade connection", sl.Err(err))
			return
		}

		room := connectToChatRoom(h.log, int32(chatID), ws)

		for {
			var mes dto.MessageRequest
			if err := ws.ReadJSON(&mes); err != nil {
				h.log.Error("failed to read message", sl.Err(err))
				disconnectFromChatRoom(h.log, int32(chatID), ws)
				h.log.Debug("chat room users", slog.Int64("chat_id", chatID), "rooms", rooms)
				break
			}

			mes.ChatID = chatID

			h.log.Debug("message received", slog.Any("message", mes))
			h.log.Debug("chat room users", slog.Int64("chat_id", chatID), "rooms", rooms)

			mesModel, err := h.sendMessageAndUpdateChat(ctx, mes, user.UserID)
			if err != nil {
				h.log.Error("failed to send message and update chat", sl.OpErr(op, err))
				continue
			}

			room.broadcast <- mesModel
		}

		h.log.Debug("connection closed")
	}
}

func (h *ChatHandler) sendMessageAndUpdateChat(ctx context.Context, mes dto.MessageRequest, sender int64) (models.Message, error) {
	const op = "handlers.chat.sendMessageAndUpdateChat"

	if err := mes.Validate(); err != nil {
		h.log.Error("failed to validate message", sl.OpErr(op, err))
		return models.Message{}, err
	}

	createdAt := time.Now().UTC()
	messageModel := dto.Message{
		ChatID:    mes.ChatID,
		Sender:    sender,
		Text:      mes.Text,
		CreatedAt: createdAt,
	}
	if err := messageModel.Validate(); err != nil {
		h.log.Error("failed to validate message", sl.OpErr(op, err))
		return models.Message{}, err
	}

	messadeID, err := h.messageService.CreateMessage(ctx, messageModel)
	if err != nil || messadeID <= 0 {
		h.log.Error("failed to create message", sl.OpErr(op, err))
		return models.Message{}, err
	}

	if err := h.chatService.UpdateChatMessage(ctx, mes.ChatID, mes.Text, createdAt); err != nil {
		h.log.Error("failed to update chat message", sl.OpErr(op, err))
		return models.Message{}, err
	}

	return models.Message{
		ID:        messadeID,
		ChatID:    mes.ChatID,
		Sender:    sender,
		Text:      mes.Text,
		CreatedAt: createdAt,
	}, nil
}
