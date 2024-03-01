package cluster

import (
	"context"
	"log/slog"

	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type credentialKey struct{}

type credentialInterceptor struct {
	CredentialStore
	log *slog.Logger
}

func (c *credentialInterceptor) Unary(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	// authentication (token verification)
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "missing metadata")
	}
	username := md["vince_username"]
	password := md["vince_username"]
	if len(username) == 0 || len(password) == 0 {
		return nil, status.Errorf(codes.Unauthenticated, "invalid credentials")
	}
	if !c.AA(username[0], password[0], permissions(req)) {
		return nil, status.Errorf(codes.Unauthenticated, "invalid credentials")
	}
	return handler(ctx, req)
}

func (c *credentialInterceptor) Stream(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	md, ok := metadata.FromIncomingContext(ss.Context())
	if !ok {
		return status.Errorf(codes.InvalidArgument, "missing metadata")
	}
	username := md["vince_username"]
	password := md["vince_password"]
	if len(username) == 0 || len(password) == 0 {
		return status.Errorf(codes.Unauthenticated, "invalid credentials")
	}
	if !c.AA(username[0], password[0], permissions(srv)) {
		return status.Errorf(codes.Unauthenticated, "invalid credentials")
	}
	return handler(srv, ss)
}

func permissions(r any) v1.Credential_Permission {
	switch r.(type) {
	case *v1.Join_Request:
		return v1.Credential_JOIN
	case *v1.Load_Request:
		return v1.Credential_LOAD
	case *v1.Backup_Request:
		return v1.Credential_BACKUP
	case *v1.RemoveNode_Request:
		return v1.Credential_REMOVE
	case *v1.Notify_Request:
		return v1.Credential_NOTIFY
	case *v1.NodeAPIRequest:
		return v1.Credential_NODE_API
	case *v1.Data:
		return v1.Credential_DATA
	case *v1.Realtime_Request, *v1.Aggregate_Request, *v1.Timeseries_Request,
		*v1.BreakDown_Request:
		return v1.Credential_QUERY
	default:
		return v1.Credential_ALL
	}
}
