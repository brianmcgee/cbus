package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/brianmcgee/cbus/pkg/rpc"
	nutil "github.com/numtide/nits/pkg/nats"
)

type responseList []*rpc.Response

func (rl responseList) Print() {
	for _, resp := range rl {
		for k, v := range resp.Msg.Header {
			fmt.Printf("%s: %s\n", k, strings.Join(v, ","))
		}
		var value interface{}
		if resp.Value != nil {
			if err := json.Unmarshal(resp.Value, &value); err != nil {
				value = err.Error()
			}
		}
		fmt.Printf("\n%v\n\n", value)
	}
}

type result struct {
	Destination string       `json:"destination"`
	Path        string       `json:"path"`
	Responses   responseList `json:"responses"`
}

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

	respCh := make(chan *rpc.Response, 16)

	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		return rpc.GetProperty(ctx, g.Destination, g.Path, g.Property, respCh, g.Nkeys...)
	})

	var responses responseList
	for resp := range respCh {
		responses = append(responses, resp)
	}

	if err = eg.Wait(); err != nil {
		return err
	}

	if CLI.Json {
		return printResponsesAsJson(g.Destination, g.Path, responses)
	} else {
		responses.Print()
		return nil
	}
}

func printResponsesAsJson(dest string, path string, responses responseList) error {
	result := result{}
	result.Destination = dest
	result.Path = path
	result.Responses = responses

	bytes, err := json.Marshal(result)
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", string(bytes))
	return nil
}
