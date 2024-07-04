package auth

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	ssogrpc "simple-chat/internal/clients/sso/grpc"
	"simple-chat/internal/domain/dto"
	"simple-chat/internal/domain/models"
	"simple-chat/internal/handlers"
	"simple-chat/internal/lib/logger/sl"
	authMiddleware "simple-chat/internal/lib/middleware"
	"strings"

	ssov1 "github.com/bordviz/sso-protos/gen/go/sso"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type AuthHandler struct {
	GRPCClient *ssogrpc.Client
	log        *slog.Logger
	AppID      int32
}

func NewAuthHandler(gRPCClient *ssogrpc.Client, log *slog.Logger, appID int32) *AuthHandler {
	return &AuthHandler{
		GRPCClient: gRPCClient,
		log:        log,
		AppID:      appID,
	}
}

func AddAuthHandler(gRPCClient *ssogrpc.Client, log *slog.Logger, appID int32) func(r chi.Router) {
	authHandler := NewAuthHandler(gRPCClient, log, appID)

	return func(r chi.Router) {
		r.Post("/register", authHandler.Register(context.Background()))
		r.Post("/login", authHandler.Login(context.Background()))
		r.With(authMiddleware.Auth(log, authHandler.GRPCClient, authHandler.AppID)).
			Get("/current_user", authHandler.CurrentUser(context.Background()))
		r.Get("/refresh_token", authHandler.RefreshToken(context.Background()))
	}
}

func (h *AuthHandler) Register(ctx context.Context) http.HandlerFunc {
	const op = "handlers.auth.Register"

	return func(w http.ResponseWriter, r *http.Request) {
		h.log = h.log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req dto.RegisterRequest

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

		userID, err := h.GRPCClient.Api.Register(
			ctx,
			&ssov1.RegisterRequest{
				Email:    req.Email,
				Password: req.Password,
				Name:     req.Name,
			},
		)

		if err != nil {
			h.log.Error("failed to register user", sl.Err(err))
			handlers.ErrorResponse(w, r, 500, "failed to register user")
			return
		}

		handlers.SuccessResponse(w, r, 201, map[string]any{
			"message": "user registered successfully",
			"user_id": userID.UserId,
		})
	}
}

func (h *AuthHandler) Login(ctx context.Context) http.HandlerFunc {
	const op = "handlers.auth.Login"

	return func(w http.ResponseWriter, r *http.Request) {
		h.log = h.log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req dto.LoginRequest

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

		tokensPair, err := h.GRPCClient.Api.Login(
			ctx,
			&ssov1.LoginRequest{
				Email:    req.Email,
				Password: req.Password,
				AppId:    h.AppID,
			},
		)

		if err != nil {
			h.log.Error("failed to login user", sl.Err(err))
			handlers.ErrorResponse(w, r, 500, "failed to login user")
			return
		}

		handlers.SuccessResponse(w, r, 200, map[string]string{
			"access_token":  tokensPair.AccessToken,
			"refresh_token": tokensPair.RefreshToken,
		})
	}
}

func (h *AuthHandler) CurrentUser(ctx context.Context) http.HandlerFunc {
	const op = "handlers.auth.CurrentUser"

	return func(w http.ResponseWriter, r *http.Request) {
		h.log = h.log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

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

		handlers.SuccessResponse(w, r, 200, user)
	}
}

func (h *AuthHandler) RefreshToken(ctx context.Context) http.HandlerFunc {
	const op = "handlers.auth.RefreshToken"

	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				h.log.Debug("recovered", slog.String("detail", fmt.Sprintf("%s", rec)))
				h.log.Error("unauthorized")
				handlers.ErrorResponse(w, r, 401, "unauthorized")
			}
		}()

		h.log = h.log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		token := strings.ReplaceAll(r.Header.Get("Authorization"), "Bearer ", "")

		tokensPair, err := h.GRPCClient.Api.RefreshToken(
			ctx,
			&ssov1.RefreshTokenRequest{
				Token: token,
				AppId: h.AppID,
			},
		)
		if err != nil {
			h.log.Error("unauthorized", sl.Err(err))
			handlers.ErrorResponse(w, r, 401, "unauthorized")
			return
		}

		handlers.SuccessResponse(w, r, 200, map[string]string{
			"access_token":  tokensPair.AccessToken,
			"refresh_token": tokensPair.RefreshToken,
		})
	}
}
