package clickhouse_logger

import (
	"context"
	"fmt"
	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	userpb "github.com/VerTox/hezzl_test/user"
	"log"
)

type Logger struct {
	Connection driver.Conn
	Context    context.Context
}

const initTable = `
CREATE TABLE IF NOT EXISTS userLog
(
userID      Int64,
name        String,
message 	String,
action_time DateTime
) engine Memory
`

func GetLogger() (*Logger, error) {
	ctx := context.Background()
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{"127.0.0.1:9000"},
	})
	if err != nil {
		return nil, err
	}
	err = conn.Ping(ctx)
	if err != nil {
		return nil, err
	}
	if err := conn.Exec(ctx, initTable); err != nil {
		return nil, err
	}
	return &Logger{
		Connection: conn,
		Context:    ctx,
	}, nil
}

func (l *Logger) UserCreatedLog(user *userpb.User) {
	err := l.Connection.AsyncInsert(l.Context, fmt.Sprintf(`INSERT INTO userLog VALUES (
			%d, '%s' , 'User with Name = %s and id = %d created', now())`, user.Id, user.Name, user.Name, user.Id), false)
	if err != nil {
		log.Fatal(err)
	}
}
