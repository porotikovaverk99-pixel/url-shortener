package grpc

import (
	"context"
	"fmt"
	"net"

	"github.com/porotikovaverk99-pixel/url-shortener/api/proto"
	"github.com/porotikovaverk99-pixel/url-shortener/internal/middleware"
	"github.com/porotikovaverk99-pixel/url-shortener/internal/model"
	"github.com/porotikovaverk99-pixel/url-shortener/internal/service"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Server struct {
	proto.UnimplementedShortenerServiceServer
	service   *service.URLService
	log       *zap.Logger
	addr      string
	grpcSrv   *grpc.Server
	secretKey string
}

func NewServer(addr string, service *service.URLService, log *zap.Logger, enableTLS bool, certFile, keyFile, secretKey string) (*Server, error) {
	var opts []grpc.ServerOption

	if enableTLS {
		creds, err := credentials.NewServerTLSFromFile(certFile, keyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load TLS credentials: %w", err)
		}
		opts = append(opts, grpc.Creds(creds))
	}

	grpcSrv := grpc.NewServer(opts...)

	s := &Server{
		service:   service,
		log:       log,
		addr:      addr,
		grpcSrv:   grpcSrv,
		secretKey: secretKey,
	}
	proto.RegisterShortenerServiceServer(grpcSrv, s)
	return s, nil
}

func (s *Server) Run() error {
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}
	s.log.Info("Starting gRPC server", zap.String("address", s.addr))
	return s.grpcSrv.Serve(lis)
}

func (s *Server) Shutdown() {
	s.log.Info("Shutting down gRPC server")
	s.grpcSrv.GracefulStop()
}

func (s *Server) getUserID(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "missing metadata")
	}

	authHeaders := md.Get("authorization")
	if len(authHeaders) == 0 {
		authHeaders = md.Get("Authorization")
	}
	if len(authHeaders) == 0 {
		return "", status.Error(codes.Unauthenticated, "missing authorization header")
	}

	token := authHeaders[0]
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}
	userID, err := middleware.ValidateToken(token, s.secretKey)
	if err != nil {
		return "", status.Error(codes.Unauthenticated, "invalid token")
	}
	return userID, nil
}

func (s *Server) ShortenURL(ctx context.Context, req *proto.URLShortenRequest) (*proto.URLShortenResponse, error) {
	if req.Url == "" {
		return nil, status.Error(codes.InvalidArgument, "url is required")
	}

	userID, err := s.getUserID(ctx)
	if err != nil {
		return nil, err
	}

	result, err := s.service.Shorten(ctx, model.RequestShorten{URL: req.Url}, userID)
	if err != nil {
		switch err {
		case service.ErrInvalidURL:
			return nil, status.Error(codes.InvalidArgument, "invalid URL")
		case service.ErrURLAlreadyExists:
			return &proto.URLShortenResponse{Result: result.Result}, status.Error(codes.AlreadyExists, "URL already exists")
		default:
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	return &proto.URLShortenResponse{Result: result.Result}, nil
}

func (s *Server) ExpandURL(ctx context.Context, req *proto.URLExpandRequest) (*proto.URLExpandResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	original, err := s.service.BaseGet(ctx, req.Id)
	if err != nil {
		switch err {
		case service.ErrURLNotFound:
			return nil, status.Error(codes.NotFound, "URL not found")
		case service.ErrURLDeleted:
			return nil, status.Error(codes.NotFound, "URL deleted")
		default:
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	return &proto.URLExpandResponse{Result: original}, nil
}

func (s *Server) ListUserURLs(ctx context.Context, req *emptypb.Empty) (*proto.UserURLsResponse, error) {
	userID, err := s.getUserID(ctx)
	if err != nil {
		return nil, err
	}

	urls, err := s.service.GetUserUrls(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get user URLs")
	}

	response := &proto.UserURLsResponse{
		Urls: make([]*proto.URLData, len(urls)),
	}
	for i, u := range urls {
		response.Urls[i] = &proto.URLData{
			ShortUrl:    u.ShortURL,
			OriginalUrl: u.OriginalURL,
		}
	}

	return response, nil
}
