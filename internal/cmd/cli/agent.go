package cli

import (
	"context"
	"syscall"

	"github.com/brianmcgee/cbus/pkg/agent"
	nutil "github.com/numtide/nits/pkg/nats"
	"github.com/ztrue/shutdown"
)

type Agent struct {
	Nats nutil.CliOptions `embed:"" prefix:"nats-"`
}

func (a *Agent) Run() (err error) {
	if err = agent.Init(a.Nats); err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	shutdown.Add(cancel)
	go shutdown.Listen(syscall.SIGINT, syscall.SIGTERM)

	return agent.Listen(ctx)
}
