package send

import (
	"github.com/nats-io/nats.go"
	nutil "github.com/numtide/nits/pkg/nats"
)

var (
	conn *nats.Conn
)

func Init(natsConfig nutil.CliOptions) (err error) {
	// connect to nats
	conn, err = natsConfig.Connect()
	return err
}
