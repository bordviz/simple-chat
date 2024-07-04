package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	ssogrpc "simple-chat/internal/clients/sso/grpc"
	"simple-chat/internal/domain/models"
	"simple-chat/internal/handlers"
	"simple-chat/internal/lib/logger/sl"
	"strings"

	ssov1 "github.com/bordviz/sso-protos/gen/go/sso"
)

type ContextKey struct {
	Name string
}

var UserContextKey = &ContextKey{"user"}

func Auth(log *slog.Logger, client *ssogrpc.Client, appID int32) func(next http.Handler) http.Handler {
	const op = "middleware.auth.Auth"

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					log.Debug("recovered", slog.String("op", op), slog.String("detail", fmt.Sprintf("%s", rec)))
					log.Error("unauthorized", slog.String("op", op))
					handlers.ErrorResponse(w, r, 401, "unauthorized")
				}
			}()

			token := strings.ReplaceAll(r.Header.Get("Authorization"), "Bearer ", "")
			user, err := client.Api.CurrentUser(
				context.Background(),
				&ssov1.CurrentUserRequest{
					Token: token,
					AppId: appID,
				},
			)
			if err != nil {
				log.Error("unauthorized", sl.OpErr(op, err))
				handlers.ErrorResponse(w, r, 401, "unauthorized")
				return
			}

			userModel := models.User{
				UserID: user.GetUserId(),
				Email:  user.GetEmail(),
				Name:   user.GetName(),
			}

			log.Debug("user from sso",
				slog.String("op", op),
				slog.Any("user", user),
				slog.Any("user_model", userModel),
			)

			ctx := context.WithValue(r.Context(), UserContextKey, userModel)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}
