package rpc

import (
	"context"
	"fmt"

	"github.com/brianmcgee/cbus/pkg/agent"
	"github.com/juju/errors"
	"github.com/nats-io/nats.go"
	nutil "github.com/numtide/nits/pkg/nats"
)

var conn *nats.Conn

func Init(natsConfig nutil.CliOptions) (err error) {
	// connect to nats
	conn, err = natsConfig.Connect()
	return err
}

func request(
	ctx context.Context,
	dest string,
	path string,
	variant string,
	name string,
	respCh chan *Response,
	nkeys ...string,
) error {
	defer close(respCh)

	// create an inbox subscription before sending the request
	inbox := nats.NewInbox()
	sub, err := conn.SubscribeSync(inbox)
	if err != nil {
		return err
	}

	var subject string

	if len(nkeys) == 0 {
		subject = fmt.Sprintf(
			"dbus.broadcast.%s%s.%s.%s",
			agent.NormalizeDestination(dest),
			agent.NormalizeObjectPath(path),
			variant,
			name,
		)

		// rpc the request
		msg := nats.NewMsg(subject)
		msg.Reply = inbox

		if err = conn.PublishMsg(msg); err != nil {
			return err
		}
	} else {
		for _, nkey := range nkeys {

			subject = fmt.Sprintf(
				"dbus.agent.%s.%s%s.%s.%s",
				nkey,
				agent.NormalizeDestination(dest),
				agent.NormalizeObjectPath(path),
				variant,
				name,
			)

			// rpc the request
			msg := nats.NewMsg(subject)
			msg.Reply = inbox

			if err = conn.PublishMsg(msg); err != nil {
				return err
			}
		}

		// since we know the max number responses we can set this
		if err = sub.AutoUnsubscribe(len(nkeys)); err != nil {
			return err
		}
	}

	// gather responses
	for {
		msg, err := sub.NextMsgWithContext(ctx)
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, nats.ErrMaxMessages) {
			return nil
		} else if err != nil {
			return err
		}

		resp := Response{}
		parseResponse(msg, &resp)
		respCh <- &resp
	}
}

func GetProperty(
	ctx context.Context,
	dest string,
	path string,
	name string,
	respCh chan *Response,
	nkeys ...string,
) error {
	return request(ctx, dest, path, "prop", name, respCh, nkeys...)
}

func InvokeMethod(
	ctx context.Context,
	dest string,
	path string,
	name string,
	respCh chan *Response,
	nkeys ...string,
) error {
	return request(ctx, dest, path, "method", name, respCh, nkeys...)
}
