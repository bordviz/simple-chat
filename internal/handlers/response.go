package handlers

import (
	"net/http"

	"github.com/go-chi/render"
)

const (
	StatusError = "error"
	StatusOk    = "ok"
)

type Response struct {
	Status string      `json:"status"`
	Detail interface{} `json:"detail"`
}

func ErrorResponse(w http.ResponseWriter, r *http.Request, status int, detail interface{}) {
	render.Status(r, status)
	render.JSON(w, r, Response{Status: StatusError, Detail: detail})
}

func SuccessResponse(w http.ResponseWriter, r *http.Request, status int, data interface{}) {
	render.Status(r, status)
	render.JSON(w, r, data)
}
