package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/juju/errors"
	nutil "github.com/numtide/nits/pkg/nats"
	"golang.org/x/crypto/ssh"
)

type NKey struct {
	PublicKey     string `help:"contents of an ed25519 public key" xor:"key"`
	PublicKeyFile string `type:"existingfile" help:"path to an ed25519 public key file" xor:"key"`
}

func (n *NKey) Run() (err error) {
	var pk ssh.PublicKey
	keyBytes := []byte(n.PublicKey)

	if !(n.PublicKey == "" || strings.Contains(n.PublicKey, "ssh-ed25519")) {
		keyBytes = []byte("ed25519 " + n.PublicKey)
	} else if n.PublicKeyFile != "" {
		if keyBytes, err = os.ReadFile(n.PublicKeyFile); err != nil {
			return errors.Annotate(err, "failed to read public key file")
		}
	}

	if pk, _, _, _, err = ssh.ParseAuthorizedKey(keyBytes); err != nil {
		return errors.Annotate(err, "failed to parse public key")
	}

	nkey, err := nutil.NKeyForPublicKey(pk)
	if err != nil {
		return errors.Annotate(err, "failed to determine nkey for public key")
	}

	fmt.Printf("%v\n", nkey)
	return nil
}
