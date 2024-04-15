package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/brianmcgee/cbus/pkg/rpc"
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
	if err = rpc.Init(i.Nats); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	respCh := make(chan *rpc.Response, 16)

	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		return rpc.InvokeMethod(ctx, i.Destination, i.Path, i.Property, respCh, i.Nkeys...)
	})

	var responses responseList
	for resp := range respCh {
		responses = append(responses, resp)
	}

	for resp := range respCh {
		for k, v := range resp.Msg.Header {
			fmt.Printf("%s: %s\n", k, strings.Join(v, ","))
		}
		fmt.Println("\n" + string(resp.Value) + "\n")
	}

	if err = eg.Wait(); err != nil {
		return err
	}

	if CLI.Json {
		return printResponsesAsJson(i.Destination, i.Path, responses)
	} else {
		responses.Print()
		return nil
	}
}
