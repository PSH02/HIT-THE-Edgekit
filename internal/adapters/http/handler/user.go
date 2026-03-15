package handler

import (
	"github.com/edgekit/edgekit/internal/adapters/http/response"
	"github.com/edgekit/edgekit/internal/core/auth"
	"github.com/edgekit/edgekit/internal/core/user"
	"github.com/edgekit/edgekit/pkg/apperror"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	svc *user.Service
}

func NewUserHandler(svc *user.Service) *UserHandler {
	return &UserHandler{svc: svc}
}

func (h *UserHandler) Register(c *gin.Context) {
	var input user.CreateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, apperror.New(apperror.CodeBadRequest, err.Error()))
		return
	}

	u, err := h.svc.Register(c.Request.Context(), input)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Created(c, u.ToProfile())
}

func (h *UserHandler) Login(c *gin.Context) {
	var input user.LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, apperror.New(apperror.CodeBadRequest, err.Error()))
		return
	}

	result, err := h.svc.Login(c.Request.Context(), input)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.OK(c, result)
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	ac, ok := auth.FromContext(c.Request.Context())
	if !ok {
		response.Fail(c, apperror.New(apperror.CodeUnauthorized, "unauthenticated"))
		return
	}

	profile, err := h.svc.GetProfile(c.Request.Context(), ac.UserID)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.OK(c, profile)
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	ac, ok := auth.FromContext(c.Request.Context())
	if !ok {
		response.Fail(c, apperror.New(apperror.CodeUnauthorized, "unauthenticated"))
		return
	}

	var input user.UpdateProfileInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, apperror.New(apperror.CodeBadRequest, err.Error()))
		return
	}

	profile, err := h.svc.UpdateProfile(c.Request.Context(), ac.UserID, input)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.OK(c, profile)
}
