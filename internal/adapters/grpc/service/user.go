package service

import (
	"context"

	"github.com/edgekit/edgekit/internal/core/user"
	"github.com/edgekit/edgekit/pkg/apperror"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type RegisterRequest struct {
	Username string
	Email    string
	Password string
}

type RegisterResponse struct {
	Id       string
	Username string
	Email    string
}

type LoginRequest struct {
	Email    string
	Password string
}

type LoginResponse struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
}

type GetProfileRequest struct {
	UserId string
}

type GetProfileResponse struct {
	Id       string
	Username string
	Email    string
	Role     string
}

type UserServiceServer struct {
	svc *user.Service
}

func NewUserServiceServer(svc *user.Service) *UserServiceServer {
	return &UserServiceServer{svc: svc}
}

func (s *UserServiceServer) Register(srv *grpc.Server) {
	sd := &grpc.ServiceDesc{
		ServiceName: "user.v1.UserService",
		HandlerType: (*UserServiceServer)(nil),
		Methods: []grpc.MethodDesc{
			{
				MethodName: "Register",
				Handler:    s.registerHandler,
			},
			{
				MethodName: "Login",
				Handler:    s.loginHandler,
			},
			{
				MethodName: "GetProfile",
				Handler:    s.getProfileHandler,
			},
		},
		Streams:  []grpc.StreamDesc{},
		Metadata: "user/v1/user.proto",
	}
	srv.RegisterService(sd, s)
}

func (s *UserServiceServer) registerHandler(
	_ interface{},
	ctx context.Context,
	dec func(interface{}) error,
	interceptor grpc.UnaryServerInterceptor,
) (interface{}, error) {
	req := &RegisterRequest{}
	if err := dec(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if interceptor != nil {
		return interceptor(ctx, req, &grpc.UnaryServerInfo{FullMethod: "/user.v1.UserService/Register"}, func(ctx context.Context, req interface{}) (interface{}, error) {
			return s.doRegister(ctx, req.(*RegisterRequest))
		})
	}
	return s.doRegister(ctx, req)
}

func (s *UserServiceServer) doRegister(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error) {
	u, err := s.svc.Register(ctx, user.CreateUserInput{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return nil, apperror.ToGRPCError(apperror.As(err))
	}
	return &RegisterResponse{
		Id:       u.ID,
		Username: u.Username,
		Email:    u.Email,
	}, nil
}

func (s *UserServiceServer) loginHandler(
	_ interface{},
	ctx context.Context,
	dec func(interface{}) error,
	interceptor grpc.UnaryServerInterceptor,
) (interface{}, error) {
	req := &LoginRequest{}
	if err := dec(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if interceptor != nil {
		return interceptor(ctx, req, &grpc.UnaryServerInfo{FullMethod: "/user.v1.UserService/Login"}, func(ctx context.Context, req interface{}) (interface{}, error) {
			return s.doLogin(ctx, req.(*LoginRequest))
		})
	}
	return s.doLogin(ctx, req)
}

func (s *UserServiceServer) doLogin(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	result, err := s.svc.Login(ctx, user.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return nil, apperror.ToGRPCError(apperror.As(err))
	}
	return &LoginResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresIn:    result.ExpiresIn,
	}, nil
}

func (s *UserServiceServer) getProfileHandler(
	_ interface{},
	ctx context.Context,
	dec func(interface{}) error,
	interceptor grpc.UnaryServerInterceptor,
) (interface{}, error) {
	req := &GetProfileRequest{}
	if err := dec(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if interceptor != nil {
		return interceptor(ctx, req, &grpc.UnaryServerInfo{FullMethod: "/user.v1.UserService/GetProfile"}, func(ctx context.Context, req interface{}) (interface{}, error) {
			return s.doGetProfile(ctx, req.(*GetProfileRequest))
		})
	}
	return s.doGetProfile(ctx, req)
}

func (s *UserServiceServer) doGetProfile(ctx context.Context, req *GetProfileRequest) (*GetProfileResponse, error) {
	profile, err := s.svc.GetProfile(ctx, req.UserId)
	if err != nil {
		return nil, apperror.ToGRPCError(apperror.As(err))
	}
	return &GetProfileResponse{
		Id:       profile.ID,
		Username: profile.Username,
		Email:    profile.Email,
		Role:     string(profile.Role),
	}, nil
}
