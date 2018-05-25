package main

import (
	"context"
	"flag"
	"io/ioutil"
	"time"

	"io"

	"os"

	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/google/subcommands"
)

const (
	DICTNAME     = `phone.dat`
	HISTORY      = `phone.history`
	RESOURCE_URL = `https://raw.githubusercontent.com/xluohome/phonedata/master/phone.dat`
)

func download(urlpath, name string) error {
	resp, err := http.Head(urlpath)
	if err != nil {
		return nil
	}
	newSize := resp.ContentLength
	resp.Body.Close()

	//head to check size
	f, err := os.Open(name)
	if err == nil {
		defer f.Close()
		st, e := f.Stat()
		if e == nil && st.Size() == newSize {
			return nil
		}
		f.Close()
	} else {
		logrus.Errorf("download: %v", err)
	}

	resp, err = http.Get(urlpath)
	if err != nil {
		logrus.Error("[updateCmd] update error: ", err)
		return err
	}
	defer resp.Body.Close()
	nname := DICTNAME + ".new"
	df, err := os.Create(nname)
	if err != nil {
		logrus.Errorf("[updateCmd] can't create file %v error: %v", nname, err)
		return err
	}
	defer df.Close()
	io.Copy(df, resp.Body)
	df.Close()

	os.Remove(DICTNAME)
	os.Rename(nname, DICTNAME)
	tn := time.Now()
	ts := tn.Format(time.RFC3339)
	ioutil.WriteFile(HISTORY, []byte(ts), 0644)
	return nil
}

type updateCmd struct {
	c subcommands.Command
	u string
}

func (p *updateCmd) Name() string {
	return "update"
}

func (p *updateCmd) Synopsis() string {
	return "update"
}

func (p *updateCmd) Usage() string {
	return "update"
}

func (p *updateCmd) SetFlags(f *flag.FlagSet) {
	flag.StringVar(&p.u, "u", RESOURCE_URL, "set update resource address")
}

func (p *updateCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	err := download(RESOURCE_URL, DICTNAME)
	if err == nil {
		logrus.Println("[updateCmd] success")
		return subcommands.ExitSuccess
	}
	logrus.Println("[updateCmd] failed ", err)
	return subcommands.ExitFailure
}
