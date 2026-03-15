package response

import (
	"net/http"

	"github.com/edgekit/edgekit/pkg/apperror"
	"github.com/gin-gonic/gin"
)

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorBody  `json:"error,omitempty"`
}

type ErrorBody struct {
	Code    apperror.Code     `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

type PaginatedData struct {
	Items interface{} `json:"items"`
	Total int64       `json:"total"`
}

func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{Success: true, Data: data})
}

func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Response{Success: true, Data: data})
}

func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

func Paginated(c *gin.Context, items interface{}, total int64) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    PaginatedData{Items: items, Total: total},
	})
}

func Fail(c *gin.Context, err error) {
	appErr := apperror.As(err)
	httpStatus := apperror.ToHTTPStatus(appErr.Code)
	c.JSON(httpStatus, Response{
		Success: false,
		Error: &ErrorBody{
			Code:    appErr.Code,
			Message: appErr.Message,
			Details: appErr.Details,
		},
	})
}

func AbortWithError(c *gin.Context, err error) {
	appErr := apperror.As(err)
	httpStatus := apperror.ToHTTPStatus(appErr.Code)
	c.AbortWithStatusJSON(httpStatus, Response{
		Success: false,
		Error: &ErrorBody{
			Code:    appErr.Code,
			Message: appErr.Message,
			Details: appErr.Details,
		},
	})
}
