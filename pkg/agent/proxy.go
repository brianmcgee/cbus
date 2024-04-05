package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
	"github.com/juju/errors"
	"github.com/nats-io/nats.go/micro"
)

var (
	endpointRegex = regexp.MustCompile(`^dbus\.(broadcast|agent\..*?)\.(.*?)\.(.*?)\.(method|prop).(.*)$`)

	serviceMap map[string]micro.Service
)

type invocation struct {
	dest    string
	path    dbus.ObjectPath
	variant string // "prop" or "method"
	target  string
}

func parseInvocation(subject string) (*invocation, error) {
	result := endpointRegex.FindStringSubmatch(subject)
	if len(result) != 6 {
		return nil, errors.Errorf("invalid request subject: %v", subject)
	}
	return &invocation{
		dest:    strings.ReplaceAll(result[2], "_", "."),
		path:    dbus.ObjectPath("/" + strings.ReplaceAll(result[3], ".", "/")),
		variant: result[4],
		target:  result[5],
	}, nil
}

func proxy(ctx context.Context) (err error) {
	// channel for receiving signals
	signals := make(chan *dbus.Signal, 32)
	dbusConn.Signal(signals)

	// This signal indicates that the owner of a name has changed. It's also the signal to use to detect the appearance
	// of new names on the bus.
	if err = dbusConn.AddMatchSignalContext(
		ctx,
		dbus.WithMatchInterface("org.freedesktop.DBus"),
		dbus.WithMatchMember("NameOwnerChanged"),
	); err != nil {
		return errors.Annotate(err, "failed to subscribe for NameOwnerChanged signals")
	}

	// do an initial fetch
	var names []string
	if err = dbusConn.BusObject().Call("org.freedesktop.DBus.ListNames", 0).Store(&names); err != nil {
		log.Fatal(err)
	}

	for _, name := range names {
		if err = addDestination(name); err != nil {
			log.Errorf("failed to add destination: %v", err)
		}
	}

	// process signal updates
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case signal, ok := <-signals:
			if !ok {
				log.Debugf("Signals channel has been closed")
				return nil
			}

			log.Debugf("Received signal: %s", signal.Name)

			name := signal.Body[0].(string)
			// oldOwner := signal.Body[1].(string)
			newOwner := signal.Body[2].(string)

			if newOwner != "" {
				if err = addDestination(name); err != nil {
					log.Errorf("failed to add destination: %v", err)
				}
			}
		}
	}
}

func addDestination(dest string) error {
	if serviceMap == nil {
		serviceMap = make(map[string]micro.Service)
	}

	cfg := micro.Config{
		Name:       NormalizeDestination(dest),
		Version:    "0.0.1", // todo is this required, does it have any meaning in this context?
		QueueGroup: nkey,
	}

	srv, err := micro.AddService(natsConn, cfg)
	if err != nil {
		return errors.Annotate(err, "failed to register service")
	}

	for _, subject := range busSubjects(dest) {

		if err = srv.AddEndpoint(
			NormalizeDestination(dest),
			micro.HandlerFunc(busHandler),
			micro.WithEndpointSubject(subject+".>"),
		); err != nil {
			return errors.Annotatef(err, "failed to register bus handler for %s", subject)
		}

		log.Info("registered bus handler", "subject", subject, "dest", dest)
	}

	serviceMap[dest] = srv

	log.Info("added destination", "dest", dest)
	return nil
}

func removeDestination(dest string) error {
	srv, ok := serviceMap[dest]
	if !ok {
		// todo create a const error
		return errors.New("destination not found")
	}
	defer func() {
		delete(serviceMap, dest)
	}()

	return srv.Stop()
}

func busHandler(req micro.Request) {
	inv, err := parseInvocation(req.Subject())
	if err != nil {
		_ = req.Error("100", err.Error(), nil)
		return
	}

	obj := dbusConn.Object(inv.dest, inv.path)

	// todo cache these lookups?
	node, err := introspect.Call(obj)
	if err != nil {
		_ = req.Error("100", "failed to introspect object", []byte(err.Error()))
		return
	}

	switch inv.variant {
	case "prop":

		// find the property
		// todo cache these lookups?
		var name string
		for _, iface := range node.Interfaces {
			for _, p := range iface.Properties {
				if p.Name == inv.target {
					name = iface.Name + "." + p.Name
				}
			}
		}

		if name == "" {
			_ = req.Error("100", fmt.Sprintf("invalid property: %v", inv.target), nil)
			return
		}

		prop, err := obj.GetProperty(name)
		if err != nil {
			_ = req.Error("100", err.Error(), nil)
			return
		}

		headers := micro.WithHeaders(
			micro.Headers{
				"NKey":      []string{nkey},
				"Signature": []string{prop.Signature().String()},
			},
		)
		_ = req.Respond([]byte(prop.String()), headers)

	case "method":

		// find the method
		var method string
		for _, iface := range node.Interfaces {
			for _, m := range iface.Methods {
				if m.Name == inv.target {
					method = iface.Name + "." + m.Name
				}
			}
		}

		if method == "" {
			_ = req.Error("100", fmt.Sprintf("invalid method: %v", inv.target), nil)
			return
		}

		flagHeader := req.Headers().Get("Method-Flag")
		if flagHeader == "" {
			flagHeader = "0"
		}

		flag, err := strconv.Atoi(flagHeader)
		if err != nil {
			_ = req.Error("100", "invalid 'Method-Flag' header", nil)
			return
		}

		// todo improve validation of args

		var invokeArgs []interface{}
		if len(req.Data()) > 0 {
			if err = json.Unmarshal(req.Data(), &invokeArgs); err != nil {
				_ = req.Error("100", "failed to unmarshal req.Data", nil)
				return
			}
		}

		call := obj.Call(method, dbus.Flags(flag), invokeArgs...)
		if call.Err != nil {
			_ = req.Error("100", "failed to invoke method: "+err.Error(), nil)
			return
		}

		bytes, err := json.Marshal(call.Body)
		if err != nil {
			_ = req.Error("100", "failed to marshal result to JSON", nil)
			return
		}

		headers := micro.WithHeaders(
			micro.Headers{
				"NKey": []string{nkey},
			},
		)
		_ = req.Respond(bytes, headers)

	default:
		log.Error("unexpected invocation variant", "subject", req.Subject(), "variant", inv.variant)
		_ = req.Error("500", "Internal Error", nil)
		return
	}
}
