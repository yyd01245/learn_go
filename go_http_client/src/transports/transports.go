package transports

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
)

type HttpHandle struct {
	url          string
	method       string
	httpClient   *http.Client
	httpLongpoll *http.Client
}

type transport interface {
	createSession(json string) interface{}
}

func NewhttpHandle(url string) *HttpHandle {
	var res = new(HttpHandle)
	tr := &http.Transport{
		DisableKeepAlives: true,
	}
	res.httpClient = &http.Client{Transport: tr}
	res.httpLongpoll = &http.Client{}
	res.url = url
	return res
}

func (obj HttpHandle) CreateSession(jsonStr []byte) (string, error) {
	//var jsonStr = []byte(json)
	req, err := http.NewRequest("POST", obj.url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	resp, err := obj.httpClient.Do(req)
	if err != nil {
		fmt.Println("unable ro reach the serve")
	} else {
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Println("body=", string(body))
		return string(body), nil
	}
	return "", err
}
func (obj HttpHandle) PostRequest(url string, jsonStr []byte) (string, error) {
	//var jsonStr = []byte(json)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	resp, err := obj.httpClient.Do(req)
	if err != nil {
		fmt.Println("unable ro reach the serve")
	} else {
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Println("body=", string(body))
		return string(body), nil
	}
	return "", err
}
func (obj HttpHandle) LongPoll(url string) (string, error) {
	//var jsonStr = []byte(json)
	req, err := http.NewRequest("GET", url, nil)
	//req.Header.Set("Content-Type", "application/json")
	//	resp, err := obj.httpClient.Do(req)
	resp, err := obj.httpLongpoll.Do(req)
	if err != nil {
		fmt.Println("unable ro reach the serve")
	} else {
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Println("body=", string(body))
		return string(body), nil
	}
	return "", err
}
