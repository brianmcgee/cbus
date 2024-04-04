package cli

import (
	"context"
	"fmt"
	"github.com/brianmcgee/cbus/pkg/send"
	nutil "github.com/numtide/nits/pkg/nats"
	"time"
)

type Property struct {
	Nats        nutil.CliOptions `embed:"" prefix:"nats-"`
	Destination string           `arg:"" help:"The bus to target"`
	Path        string           `arg:"" help:"The object path to target"`
	Property    string           `arg:"" help:"The property to retrieve"`
}

func (s *Property) Run() (err error) {
	if err = send.Init(s.Nats); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	value, err := send.Property(ctx, s.Destination, s.Path, s.Property)
	if err != nil {
		return err
	}

	fmt.Println(string(value))
	return nil
}
