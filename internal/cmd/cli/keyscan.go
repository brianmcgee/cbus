package cli

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
	"sync/atomic"
	"syscall"

	"github.com/juju/errors"
	nutil "github.com/numtide/nits/pkg/nats"
	"github.com/ztrue/shutdown"
	"golang.org/x/crypto/ssh"
	sshagent "golang.org/x/crypto/ssh/agent"
	"golang.org/x/sync/errgroup"
)

type nkeyResolution struct {
	Hostname string
	NKey     string `json:",omitempty"`
	Err      error  `json:",omitempty"`
}

type Keyscan struct {
	Port  int      `default:"22" help:"ssh port to connect on"`
	Hosts []string `arg:"" help:"list of hosts to connect to" xor:"hosts"`
	Stdin bool     `help:"read list of hosts from stdin" xor:"hosts"`

	waitCount    atomic.Int32
	waitDelta    int32
	authMethods  []ssh.AuthMethod
	resolutionCh chan *nkeyResolution
}

func (k *Keyscan) Run() error {
	// connect to ssh auth socket
	authSocket := os.Getenv("SSH_AUTH_SOCK")
	if authSocket == "" {
		return errors.New("SSH_AUTH_SOCK is not defined")
	}

	conn, err := net.Dial("unix", authSocket)
	if err != nil {
		return errors.Annotatef(err, "failed to connect to %s", authSocket)
	}
	defer conn.Close()

	sshAgent := sshagent.NewClient(conn)
	k.authMethods = []ssh.AuthMethod{ssh.PublicKeysCallback(sshAgent.Signers)}

	// construct an app context and listen for shutdown signals
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	shutdown.Add(cancel)
	go shutdown.Listen(syscall.SIGINT, syscall.SIGTERM)

	//

	if k.Stdin {
		scanner := bufio.NewScanner(bufio.NewReader(os.Stdin))
		for scanner.Scan() {
			k.Hosts = append(k.Hosts, scanner.Text())
		}
	}

	// process hosts list
	eg, ctx := errgroup.WithContext(ctx)

	k.waitCount = atomic.Int32{}
	k.waitCount.Store(int32(len(k.Hosts)))
	k.resolutionCh = make(chan *nkeyResolution, 16)

	// create a worker for each host
	for _, host := range k.Hosts {
		eg.Go(k.dial(host))
	}

	// process resolutions
	eg.Go(func() error {
		if CLI.Json {
			return k.jsonOut()
		} else {
			k.textOut()
		}
		return nil
	})

	return eg.Wait()
}

func (k *Keyscan) dial(host string) func() error {
	return func() error {
		defer func() {
			// close the resolution channel if we have processed all hosts
			if k.waitCount.Add(-1) == 0 {
				close(k.resolutionCh)
			}
		}()

		config := &ssh.ClientConfig{
			Auth:              k.authMethods,
			HostKeyAlgorithms: []string{ssh.KeyAlgoED25519},
			HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
				// attempt to calculate the nkey
				resp := nkeyResolution{
					Hostname: host,
				}
				resp.NKey, resp.Err = nutil.NKeyForPublicKey(key)
				k.resolutionCh <- &resp
				return nil
			},
		}
		_, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", host, k.Port), config)
		// todo is there a better way of checking the error?
		if err != nil && strings.Contains(err.Error(), "ssh: handshake failed") {
			// we don't care about failing to connect, the host key check is all we're interested in
			err = nil
		}
		return err
	}
}

func (k *Keyscan) textOut() {
	for res := range k.resolutionCh {
		if res.Err == nil {
			if len(k.Hosts) == 1 {
				// simple output for only one host
				fmt.Println(res.NKey)
			} else {
				// more complex output for multiple hosts
				fmt.Printf("%s %s\n", res.Hostname, res.NKey)
			}
		} else {
			_, _ = fmt.Fprintf(os.Stderr, "%s\n", res.Err)
		}
	}
}

func (k *Keyscan) jsonOut() error {
	var list []*nkeyResolution
	for res := range k.resolutionCh {
		list = append(list, res)
	}
	bytes, err := json.Marshal(list)
	if err != nil {
		return errors.Annotatef(err, "failed to marshal to json")
	}
	fmt.Printf("%s\n", string(bytes))
	return nil
}
