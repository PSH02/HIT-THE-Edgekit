package apperror

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var grpcCodeMap = map[Code]codes.Code{
	CodeBadRequest:   codes.InvalidArgument,
	CodeUnauthorized: codes.Unauthenticated,
	CodeForbidden:    codes.PermissionDenied,
	CodeNotFound:     codes.NotFound,
	CodeConflict:     codes.AlreadyExists,
	CodeRateLimited:  codes.ResourceExhausted,
	CodeInternal:     codes.Internal,
	CodeUnavailable:  codes.Unavailable,
}

var httpStatusMap = map[Code]int{
	CodeOK:           200,
	CodeBadRequest:   400,
	CodeUnauthorized: 401,
	CodeForbidden:    403,
	CodeNotFound:     404,
	CodeConflict:     409,
	CodeRateLimited:  429,
	CodeInternal:     500,
	CodeUnavailable:  503,
}

func ToGRPCError(err *AppError) error {
	c, ok := grpcCodeMap[err.Code]
	if !ok {
		c = codes.Unknown
	}
	return status.Error(c, err.Message)
}

func ToHTTPStatus(code Code) int {
	if s, ok := httpStatusMap[code]; ok {
		return s
	}
	return 500
}
