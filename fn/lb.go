package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/iron-io/functions/lb"
	"github.com/urfave/cli"
)

type lbFnCmd struct {
	nodes  string
	listen string
}

func balance() cli.Command {
	l := lbFnCmd{}

	return cli.Command{
		Name:        "balance",
		Usage:       "set up load-balancer for functions",
		Description: "Creates a func.yaml file in the current directory.  ",
		ArgsUsage:   "fn balance `listen` `nodes`",
		Action:      l.balance,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:        "listen",
				Usage:       "listening port for incoming connections",
				Destination: &l.listen,
				Value:       "0.0.0.0:8081",
			},
			cli.StringFlag{
				Name:        "nodes",
				Usage:       "comma separated list of IronFunction nodes",
				Destination: &l.nodes,
				Value:       "127.0.0.1:8080",
			},
		},
	}
}

func (l *lbFnCmd) balance(c *cli.Context) error {
	nodes := strings.Split(l.nodes, ",")
	p := lb.ConsistentHashReverseProxy(context.Background(), nodes)
	fmt.Printf("Starting Functions Load Balancer on %s for %s.\n", l.listen, strings.Join(nodes, ", "))
	if err := http.ListenAndServe(l.listen, p); err != nil {
		fmt.Fprintln(os.Stderr, "could not start server. error:", err)
		return err
	}
	return nil
}
