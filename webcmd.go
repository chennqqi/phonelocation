package main

import (
	"context"
	"flag"
	"os"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/google/subcommands"
)

type webCmd struct {
	w      WebServer
	port   int
	update string
	reload string
}

func (p *webCmd) Name() string {
	return "web"
}

func (p *webCmd) Synopsis() string {
	return "web"
}

func (p *webCmd) Usage() string {
	return "web -p port"
}

func (p *webCmd) SetFlags(f *flag.FlagSet) {
	f.IntVar(&p.port, "p", 8080, "set port")
	f.StringVar(&p.update, "u", "1d", "set update timeout")
	f.StringVar(&p.reload, "r", "10m", "set reload timeout")
}

func (p *webCmd) Execute(context.Context, *flag.FlagSet, ...interface{}) subcommands.ExitStatus {
	f, err := os.Open(DICTNAME)
	if os.IsNotExist(err) {
		err = download(RESOURCE_URL, DICTNAME)
		if err != nil {
			logrus.Errorf("[webCmd] init %v error %v", DICTNAME, err)
			return subcommands.ExitFailure
		}
	}
	f.Close()

	p.w.upTo, _ = time.ParseDuration(p.update)
	p.w.reLoadTo, _ = time.ParseDuration(p.reload)
	p.w.Run(p.port)
	return subcommands.ExitSuccess
}
