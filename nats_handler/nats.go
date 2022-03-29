package nats_handler

import (
	"encoding/json"
	"github.com/VerTox/hezzl_test/clickhouse_logger"
	userpb "github.com/VerTox/hezzl_test/user"
	"github.com/nats-io/nats.go"
	"log"
	"time"
)

type Nats struct {
	Connection *nats.Conn
	Logger     *clickhouse_logger.Logger
}

func NewNatsConn() (*Nats, error) {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		return nil, err
	}
	logger, err := clickhouse_logger.GetLogger()
	return &Nats{Connection: nc, Logger: logger}, nil
}

func (n *Nats) ListenUserCreation() {
	sub, err := n.Connection.SubscribeSync("userCreation")
	if err != nil {
		log.Fatal(err)
	}

	// looks terrible, need to find a better solution
	for {
		m, err := sub.NextMsg(10 * time.Second)
		if err != nil {
			if err != nats.ErrTimeout {
				log.Fatal(err)
			}
		} else {
			user := userpb.User{}
			err := json.Unmarshal(m.Data, &user)
			if err != nil {
				log.Fatal(err)
			}
			n.Logger.UserCreatedLog(&user)
		}

	}
}

func (n *Nats) SendToQueue(v interface{}, subject string) {
	ec, err := nats.NewEncodedConn(n.Connection, nats.JSON_ENCODER)
	if err != nil {
		log.Fatal(err)
	}

	if err := ec.Publish(subject, &v); err != nil {
		log.Fatal(err)
	}
}
