package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/chennqqi/phonedata"
	"github.com/gin-gonic/gin"
)

var phoneExp = regexp.MustCompile(`1\d{10}`)

type WebServer struct {
	pd       *phonedata.PhoneDict
	lastUp   time.Time
	reLoadTo time.Duration
	upTo     time.Duration
}

func (s *WebServer) Run(port int) error {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.GET("/location/:phone", s.location)
	r.GET("/location", s.locationQuery)

	r.LoadHTMLGlob("assets/*.html")
	r.Static("/css", "assets/css")
	r.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", nil)
	})

	pd, err := phonedata.Parse(DICTNAME)
	if err != nil {
		return err
	}
	s.pd = pd

	go s.liveReload()
	return r.Run(fmt.Sprintf(":%d", port))
}

func (s *WebServer) location(c *gin.Context) {
	pd := s.pd

	phoneNumber := c.Param("phone")
	if !phoneExp.MatchString(phoneNumber) {
		var rs = struct {
			Msg    string `json:"message"`
			Status int    `json:"status"`
		}{"无效的手机号", 200}

		c.JSON(200, &rs)
		return
	}

	resp, err := pd.Find(phoneNumber)
	if err == nil {
		c.JSON(200, resp)
		return
	}
	var rs = struct {
		Msg    string `json:"message"`
		Status int    `json:"status"`
	}{"not found", 200}

	c.JSON(200, &rs)
}

func (s *WebServer) liveReload() {
	reTicker := time.NewTicker(s.reLoadTo)
	defer reTicker.Stop()
	upTicker := time.NewTimer(s.upTo)
	defer upTicker.Stop()

	//for init
	_, err := ioutil.ReadFile(HISTORY)
	if !os.IsExist(err) {
		f, err := os.Open(DICTNAME)
		if err != nil {
			logrus.Panic("[web] open dict error %v", err)
		}
		st, _ := f.Stat()
		lt := st.ModTime()
		ioutil.WriteFile(HISTORY, []byte(lt.Format(time.RFC3339)), 0644)
	} else if err != nil {
		logrus.Errorf("[web] open history error %v", err)
	}

	for {
		select {
		case <-reTicker.C:
			txt, err := ioutil.ReadFile(HISTORY)
			if err != nil {
				logrus.Errorf("[web] open history error %v", err)
				continue
			}
			last, err := time.Parse(time.RFC3339, string(txt))
			if err != nil {
				logrus.Errorf("[web] time.Parse error %v", err)
				continue

			}
			if last.After(s.lastUp) {
				npd, err := phonedata.Parse(DICTNAME)
				if err != nil {
					logrus.Errorf("[web] phonedata.Parse error %v", err)
					continue
				}
				s.pd = npd
			}

		case <-upTicker.C:
			download(RESOURCE_URL, DICTNAME)
		}
	}
}

func (s *WebServer) locationQuery(c *gin.Context) {
	pd := s.pd

	phoneNumber, _ := c.GetQuery("phone")
	if !phoneExp.MatchString(phoneNumber) {
		var rs = struct {
			Msg    string `json:"message"`
			Status int    `json:"status"`
		}{"无效的手机号", 200}

		c.JSON(200, &rs)
		return
	}

	resp, err := pd.Find(phoneNumber)
	if err == nil {
		c.JSON(200, resp)
		return
	}
	var rs = struct {
		Msg    string `json:"message"`
		Status int    `json:"status"`
	}{"not found", 200}

	c.JSON(200, &rs)
}
