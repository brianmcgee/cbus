package send

import (
	"context"
	"fmt"
	"strings"
)

func Property(ctx context.Context, dest string, path string, name string) ([]byte, error) {
	subject := fmt.Sprintf(
		"dbus.bus.%s%s.%s",
		strings.ReplaceAll(strings.ReplaceAll(dest, ".", "_"), ":", "_"),
		strings.ReplaceAll(path, "/", "."),
		name,
	)

	msg, err := conn.RequestWithContext(ctx, subject, nil)
	return msg.Data, err
}
