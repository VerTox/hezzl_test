package main

import (
	"context"
	"encoding/json"
	"github.com/VerTox/hezzl_test/nats_handler"
	userpb "github.com/VerTox/hezzl_test/user"
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"time"
)

const cacheKey = "user_list"

type UserServer struct {
	Nats   *nats_handler.Nats
	DBConn *pgx.Conn
	Redis  *redis.Client
	userpb.UnimplementedUserServiceServer
}

func NewUserServer(n *nats_handler.Nats, p *pgx.Conn, r *redis.Client) (*UserServer, error) {
	return &UserServer{
		Nats:                           n,
		DBConn:                         p,
		Redis:                          r,
		UnimplementedUserServiceServer: userpb.UnimplementedUserServiceServer{},
	}, nil
}

func (u UserServer) CreateUser(_ context.Context, request *userpb.CreateUserRequest) (*userpb.User, error) {
	userName := request.GetName()
	if userName == "" {
		log.Print("name is not valid")
		return nil, status.Error(codes.InvalidArgument, "name is not valid")
	}
	var userId int64
	err := u.DBConn.QueryRow(context.Background(), "INSERT INTO users (name) VALUES($1) returning id", userName).Scan(&userId)
	if err != nil {
		return nil, err
	}
	user := userpb.User{
		Id:   userId,
		Name: userName,
	}
	u.Redis.Del(context.Background(), cacheKey)
	u.Nats.SendToQueue(&user, "userCreation")
	return &user, nil
}

func (u UserServer) DeleteUser(_ context.Context, request *userpb.DeleteUserRequest) (*userpb.User, error) {
	userId := request.GetId()
	var user userpb.User

	err := u.DBConn.QueryRow(context.Background(), "DELETE FROM users WHERE id = $1 returning *", userId).Scan(&user.Id, &user.Name)
	if err != nil {
		return nil, err
	}
	u.Redis.Del(context.Background(), cacheKey)

	return &user, err
}

func (u UserServer) ListUsers(context.Context, *userpb.ListUserRequest) (*userpb.ListUser, error) {
	var users []*userpb.User
	result, err := u.Redis.Get(context.Background(), cacheKey).Result()
	if err == nil {
		err := json.Unmarshal([]byte(result), &users)
		if err != nil {
			return nil, err
		}
	} else {
		rows, err := u.DBConn.Query(context.Background(), "SELECT * FROM users")
		if err != nil {
			return nil, err
		}

		for rows.Next() {
			user := userpb.User{}
			err := rows.Scan(&user.Id, &user.Name)
			if err != nil {
				return nil, err
			}
			users = append(users, &user)
		}
		userBytes, err := json.Marshal(users)
		if err != nil {
			return nil, err
		}

		err = u.Redis.SetEX(context.Background(), cacheKey, userBytes, time.Minute).Err()
		if err != nil {
			return nil, err
		}
	}

	return &userpb.ListUser{Users: users}, nil
}
