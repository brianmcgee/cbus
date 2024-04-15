package agent

import (
	"context"

	"github.com/godbus/dbus/v5"
	"github.com/juju/errors"
	"github.com/nats-io/nats.go"
	nutil "github.com/numtide/nits/pkg/nats"
	"golang.org/x/sync/errgroup"
)

var (
	natsConn    *nats.Conn
	dbusConn    *dbus.Conn
	monitorConn *dbus.Conn

	nkey string
)

func Init(natsConfig nutil.CliOptions) (err error) {
	// connect to nats
	natsConn, err = natsConfig.Connect()
	if err != nil {
		return err
	}

	// create a main system bus connection
	dbusConn, err = dbus.ConnectSystemBus()
	if err != nil {
		return errors.Annotate(err, "failed to connect to system dbus")
	}

	// create a separate monitor connection, it can only be used to listen for messages, not rpc
	monitorConn, err = dbus.ConnectSystemBus()
	if err != nil {
		return errors.Annotate(err, "failed to connect to system dbus")
	}

	_, nkey, _, _ = natsConfig.ToNatsOptions()

	return nil
}

func Listen(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error { return monitorSignals(ctx) })
	eg.Go(func() error { return proxy(ctx) })

	return eg.Wait()
}
