package handler

import (
	"strconv"

	"github.com/edgekit/edgekit/internal/adapters/http/response"
	"github.com/edgekit/edgekit/internal/core/auth"
	"github.com/edgekit/edgekit/internal/core/session"
	"github.com/edgekit/edgekit/pkg/apperror"
	"github.com/gin-gonic/gin"
)

type SessionHandler struct {
	svc *session.Service
}

func NewSessionHandler(svc *session.Service) *SessionHandler {
	return &SessionHandler{svc: svc}
}

func (h *SessionHandler) Create(c *gin.Context) {
	ac, ok := auth.FromContext(c.Request.Context())
	if !ok {
		response.Fail(c, apperror.New(apperror.CodeUnauthorized, "unauthenticated"))
		return
	}

	var input session.CreateSessionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, apperror.New(apperror.CodeBadRequest, err.Error()))
		return
	}

	view, err := h.svc.Create(c.Request.Context(), ac.UserID, input)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Created(c, view)
}

func (h *SessionHandler) Get(c *gin.Context) {
	id := c.Param("id")

	view, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.OK(c, view)
}

func (h *SessionHandler) Join(c *gin.Context) {
	ac, ok := auth.FromContext(c.Request.Context())
	if !ok {
		response.Fail(c, apperror.New(apperror.CodeUnauthorized, "unauthenticated"))
		return
	}

	id := c.Param("id")

	view, err := h.svc.Join(c.Request.Context(), id, ac.UserID)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.OK(c, view)
}

func (h *SessionHandler) Leave(c *gin.Context) {
	ac, ok := auth.FromContext(c.Request.Context())
	if !ok {
		response.Fail(c, apperror.New(apperror.CodeUnauthorized, "unauthenticated"))
		return
	}

	id := c.Param("id")

	view, err := h.svc.Leave(c.Request.Context(), id, ac.UserID)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.OK(c, view)
}

func (h *SessionHandler) ListWaiting(c *gin.Context) {
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	views, err := h.svc.ListWaiting(c.Request.Context(), offset, limit)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.OK(c, views)
}
