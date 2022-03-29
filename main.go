package main

import (
	"github.com/VerTox/hezzl_test/nats_handler"
	"github.com/VerTox/hezzl_test/postgres_conn"
	"github.com/VerTox/hezzl_test/redis_conn"
	userpb "github.com/VerTox/hezzl_test/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	"net"
)

func main() {
	listener, err := net.Listen("tcp", ":5300")

	if err != nil {
		grpclog.Fatalf("failed to listen: %v", err)
	}

	natsHandler, err := nats_handler.NewNatsConn()
	if err != nil {
		panic(err)
	}
	defer natsHandler.Connection.Close()

	go natsHandler.ListenUserCreation()

	postgresConn, err := postgres_conn.NewPostgresConn()

	if err != nil {
		panic(err)
	}

	redisConn := redis_conn.NewRedisConn()

	opts := []grpc.ServerOption{}
	grpcServer := grpc.NewServer(opts...)

	userServer, err := NewUserServer(natsHandler, postgresConn, redisConn)

	userpb.RegisterUserServiceServer(grpcServer, userServer)
	err = grpcServer.Serve(listener)
	if err != nil {
		panic(err)
	}
}
