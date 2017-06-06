package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"./src/janus"
	"github.com/Sirupsen/logrus"
	"github.com/jessevdk/go-flags"
)

var opts struct {
	Number   int     `short:"n" long:"number" default:"1" description:"loop number"`
	Url      string  `short:"u" long:"url" description:"url"`
	ClientID int64   `short:"c" long:"clientid" default:"2000" despcription:"clientid"`
	Log      string  `short:"l" long:"log" default:"./output.log" description:"the log file to tail -f"`
	Ptype    int     `short:"p" long:"ptype" default:"1" description:"1 pulisher 2 listener"`
	Step     int     `short:"s" long:"step" default:"1" despcription:"step chrome dir"`
	Qps      float64 `short:"q" long:"qps" default:"512" despcription:"qps kbps"`
}

type startParam struct {
	clientid uint64
	ptype    int
	url      string
	qps      float64
}

var log *logrus.Logger

var listJanus []*janus.JanusData
var index = 0

func init() {
	log = logrus.New()
	log.Level = logrus.InfoLevel
	f := new(logrus.TextFormatter)
	f.TimestampFormat = "2006-01-02 15:04:05"
	f.FullTimestamp = true
	log.Formatter = f
}
func Run() {
	var p startParam
	var clientid int64
	clientid = opts.ClientID
	if opts.ClientID == 0 {
		rand.Seed(time.Now().UnixNano())
		clientid = rand.Int63n(10000)

	}
	p.clientid = uint64(clientid)
	for i := 0; i < opts.Number; i++ {
		p.clientid = p.clientid + uint64(i)*uint64(opts.Step)
		p.url = opts.Url
		p.ptype = opts.Ptype
		p.qps = opts.Qps
		startOne(p)
		time.Sleep(500)
	}
}

func startOne(param startParam) {
	fmt.Println("param :%v", param)

	fmt.Println("start one %s over", param.url)
	/*
		var jsonStr = []byte(`{"uprtc":"create","transaction":"UKzUEYCgKuLn"}`)
		client := http.Client{}
		req, err := http.NewRequest("POST", param.url, bytes.NewBuffer(jsonStr))
		req.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("unable ro reach the serve")
		} else {
			body, _ := ioutil.ReadAll(resp.Body)
			fmt.Println("body=", string(body))
		}
	*/
	// httpReq := transports.NewhttpHandle(param.url)
	// resp, _ := httpReq.CreateSession(`{"uprtc":"create","transaction":"UKzUEYCgKuLn"}`)
	janusObj := janus.NewJanusObject(param.url, param.clientid, param.ptype, param.qps)
	janusObj.CreateSession()
	// listJanus[index] = janusObj
	// index += 1
}

func main() {
	_, err := flags.Parse(&opts)
	if err != nil {
		if !strings.Contains(err.Error(), "Usage") {
			log.Fatalf("error: %v", err)
		} else {
			return
		}
	}
	if opts.Number == 0 || opts.Url == "" {
		log.Println("error no input cmd")
		return
	}
	out, err := os.Create(opts.Log)
	if err != nil {
		log.Errorf("creat %s file error", opts.Log)
		return
	}
	defer out.Close()
	log.Out = out
	Run()

	select {}
	// quit delete profile

}
