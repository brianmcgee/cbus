package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/godbus/dbus/v5"
	"github.com/juju/errors"
	"github.com/nats-io/nats.go"
)

func monitorSignals(ctx context.Context) error {
	// empty rules indicates we want to monitor everything
	var rules []string
	call := monitorConn.BusObject().Call("org.freedesktop.DBus.Monitoring.BecomeMonitor", 0, rules, uint(0))
	if call.Err != nil {
		return errors.Annotate(call.Err, "failed to become monitor")
	}

	// channel for receiving signals
	busMsgs := make(chan *dbus.Message, 32)
	monitorConn.Eavesdrop(busMsgs)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case busMsg, ok := <-busMsgs:

			if !ok {
				return nil
			} else if busMsg.Type != dbus.TypeSignal {
				// we're only interested in signals
				continue
			}

			var sender, iface, member, path string
			// todo see if there's a helper in dbus somewhere for parsing this
			if err := busMsg.Headers[dbus.FieldSender].Store(&sender); err != nil {
				log.Errorf("failed to get sender from signal msg: %v", err)
				continue
			} else if err = busMsg.Headers[dbus.FieldInterface].Store(&iface); err != nil {
				log.Errorf("failed to get interface from signal msg: %v", err)
				continue
			} else if err = busMsg.Headers[dbus.FieldMember].Store(&member); err != nil {
				log.Errorf("failed to get member from signal msg: %v", err)
				continue
			} else if err = busMsg.Headers[dbus.FieldPath].Store(&path); err != nil {
				log.Errorf("failed to get path from signal msg: %v", err)
				continue
			}

			subject := fmt.Sprintf(
				"dbus.signals.%s.%s%s",
				nkey,
				strings.ReplaceAll(strings.ReplaceAll(sender, ".", "_"), ":", "_"),
				strings.ReplaceAll(path, "/", "."),
			)

			msg := nats.NewMsg(subject)
			msg.Header.Set("NKey", nkey)
			msg.Header.Set("Sender", sender)
			msg.Header.Set("Interface", iface)
			msg.Header.Set("Member", member)
			msg.Header.Set("Path", path)

			bytes, err := json.Marshal(busMsg.Body)
			if err != nil {
				log.Errorf("failed to marshal msg body: %v", err)
			}

			msg.Data = bytes
			if err = natsConn.PublishMsg(msg); err != nil {
				log.Errorf("failed to publish signal msg: %v", err)
			}
		}
	}
}
