package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
	"golang.org/x/sync/errgroup"

	"github.com/brianmcgee/cbus/pkg/send"
	nutil "github.com/numtide/nits/pkg/nats"
)

type Invoke struct {
	Nats        nutil.CliOptions `embed:"" prefix:"nats-"`
	Destination string           `arg:"" help:"The bus to target"`
	Path        string           `arg:"" help:"The object path to target"`
	Property    string           `arg:"" help:"The property to retrieve"`

	Nkeys []string `help:"A list of agent nkeys to target"`
}

func (i *Invoke) Run() (err error) {
	if err = send.Init(i.Nats); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	respCh := make(chan *nats.Msg, 16)

	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		return send.Method(ctx, i.Destination, i.Path, i.Property, respCh, i.Nkeys...)
	})

	for msg := range respCh {
		for k, v := range msg.Header {
			fmt.Printf("%s: %s\n", k, strings.Join(v, ","))
		}
		fmt.Println("\n" + string(msg.Data) + "\n")
	}

	return eg.Wait()
}
