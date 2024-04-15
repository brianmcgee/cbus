package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
	"golang.org/x/sync/errgroup"

	"github.com/brianmcgee/cbus/pkg/rpc"
	nutil "github.com/numtide/nits/pkg/nats"
)

type Get struct {
	Nats        nutil.CliOptions `embed:"" prefix:"nats-"`
	Destination string           `arg:"" help:"The bus to target"`
	Path        string           `arg:"" help:"The object path to target"`
	Property    string           `arg:"" help:"The property to retrieve"`

	Nkeys []string `help:"A list of agent nkeys to target"`
}

func (g *Get) Run() (err error) {
	if err = rpc.Init(g.Nats); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	respCh := make(chan *nats.Msg, 16)

	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		return rpc.GetProperty(ctx, g.Destination, g.Path, g.Property, respCh, g.Nkeys...)
	})

	for msg := range respCh {
		for k, v := range msg.Header {
			fmt.Printf("%s: %s\n", k, strings.Join(v, ","))
		}
		fmt.Println("\n" + string(msg.Data) + "\n")
	}

	return eg.Wait()
}
