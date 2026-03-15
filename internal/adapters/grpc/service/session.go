package service

import (
	"context"

	"github.com/edgekit/edgekit/internal/core/auth"
	"github.com/edgekit/edgekit/internal/core/session"
	"github.com/edgekit/edgekit/pkg/apperror"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CreateSessionRequest struct {
	Name       string
	MaxPlayers int32
	Metadata   map[string]string
}

type GetSessionRequest struct {
	SessionId string
}

type JoinSessionRequest struct {
	SessionId string
}

type LeaveSessionRequest struct {
	SessionId string
}

type SessionResponse struct {
	Id         string
	Name       string
	HostUserId string
	MaxPlayers int32
	Status     string
	Players    []string
	Metadata   map[string]string
	CreatedAt  string
}

type SessionServiceServer struct {
	svc *session.Service
}

func NewSessionServiceServer(svc *session.Service) *SessionServiceServer {
	return &SessionServiceServer{svc: svc}
}

func (s *SessionServiceServer) Register(srv *grpc.Server) {
	sd := &grpc.ServiceDesc{
		ServiceName: "session.v1.SessionService",
		HandlerType: (*SessionServiceServer)(nil),
		Methods: []grpc.MethodDesc{
			{
				MethodName: "CreateSession",
				Handler:    s.createSessionHandler,
			},
			{
				MethodName: "GetSession",
				Handler:    s.getSessionHandler,
			},
			{
				MethodName: "JoinSession",
				Handler:    s.joinSessionHandler,
			},
			{
				MethodName: "LeaveSession",
				Handler:    s.leaveSessionHandler,
			},
		},
		Streams:  []grpc.StreamDesc{},
		Metadata: "session/v1/session.proto",
	}
	srv.RegisterService(sd, s)
}

func (s *SessionServiceServer) createSessionHandler(
	_ interface{},
	ctx context.Context,
	dec func(interface{}) error,
	interceptor grpc.UnaryServerInterceptor,
) (interface{}, error) {
	req := &CreateSessionRequest{}
	if err := dec(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if interceptor != nil {
		return interceptor(ctx, req, &grpc.UnaryServerInfo{FullMethod: "/session.v1.SessionService/CreateSession"}, func(ctx context.Context, req interface{}) (interface{}, error) {
			return s.doCreateSession(ctx, req.(*CreateSessionRequest))
		})
	}
	return s.doCreateSession(ctx, req)
}

func (s *SessionServiceServer) doCreateSession(ctx context.Context, req *CreateSessionRequest) (*SessionResponse, error) {
	ac, ok := auth.FromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}
	view, err := s.svc.Create(ctx, ac.UserID, session.CreateSessionInput{
		Name:       req.Name,
		MaxPlayers: int(req.MaxPlayers),
		Metadata:   req.Metadata,
	})
	if err != nil {
		return nil, apperror.ToGRPCError(apperror.As(err))
	}
	return sessionViewToResponse(view), nil
}

func (s *SessionServiceServer) getSessionHandler(
	_ interface{},
	ctx context.Context,
	dec func(interface{}) error,
	interceptor grpc.UnaryServerInterceptor,
) (interface{}, error) {
	req := &GetSessionRequest{}
	if err := dec(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if interceptor != nil {
		return interceptor(ctx, req, &grpc.UnaryServerInfo{FullMethod: "/session.v1.SessionService/GetSession"}, func(ctx context.Context, req interface{}) (interface{}, error) {
			return s.doGetSession(ctx, req.(*GetSessionRequest))
		})
	}
	return s.doGetSession(ctx, req)
}

func (s *SessionServiceServer) doGetSession(ctx context.Context, req *GetSessionRequest) (*SessionResponse, error) {
	view, err := s.svc.Get(ctx, req.SessionId)
	if err != nil {
		return nil, apperror.ToGRPCError(apperror.As(err))
	}
	return sessionViewToResponse(view), nil
}

func (s *SessionServiceServer) joinSessionHandler(
	_ interface{},
	ctx context.Context,
	dec func(interface{}) error,
	interceptor grpc.UnaryServerInterceptor,
) (interface{}, error) {
	req := &JoinSessionRequest{}
	if err := dec(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if interceptor != nil {
		return interceptor(ctx, req, &grpc.UnaryServerInfo{FullMethod: "/session.v1.SessionService/JoinSession"}, func(ctx context.Context, req interface{}) (interface{}, error) {
			return s.doJoinSession(ctx, req.(*JoinSessionRequest))
		})
	}
	return s.doJoinSession(ctx, req)
}

func (s *SessionServiceServer) doJoinSession(ctx context.Context, req *JoinSessionRequest) (*SessionResponse, error) {
	ac, ok := auth.FromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}
	view, err := s.svc.Join(ctx, req.SessionId, ac.UserID)
	if err != nil {
		return nil, apperror.ToGRPCError(apperror.As(err))
	}
	return sessionViewToResponse(view), nil
}

func (s *SessionServiceServer) leaveSessionHandler(
	_ interface{},
	ctx context.Context,
	dec func(interface{}) error,
	interceptor grpc.UnaryServerInterceptor,
) (interface{}, error) {
	req := &LeaveSessionRequest{}
	if err := dec(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if interceptor != nil {
		return interceptor(ctx, req, &grpc.UnaryServerInfo{FullMethod: "/session.v1.SessionService/LeaveSession"}, func(ctx context.Context, req interface{}) (interface{}, error) {
			return s.doLeaveSession(ctx, req.(*LeaveSessionRequest))
		})
	}
	return s.doLeaveSession(ctx, req)
}

func (s *SessionServiceServer) doLeaveSession(ctx context.Context, req *LeaveSessionRequest) (*SessionResponse, error) {
	ac, ok := auth.FromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}
	view, err := s.svc.Leave(ctx, req.SessionId, ac.UserID)
	if err != nil {
		return nil, apperror.ToGRPCError(apperror.As(err))
	}
	return sessionViewToResponse(view), nil
}

func sessionViewToResponse(v *session.SessionView) *SessionResponse {
	return &SessionResponse{
		Id:         v.ID,
		Name:       v.Name,
		HostUserId: v.HostUserID,
		MaxPlayers: int32(v.MaxPlayers),
		Status:     string(v.Status),
		Players:    v.Players,
		Metadata:   v.Metadata,
		CreatedAt:  v.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
