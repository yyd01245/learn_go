package janus

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"../transports"
	"github.com/jmcvetta/randutil"
	webrtc "github.com/keroserene/go-webrtc"
	"github.com/tidwall/gjson"
)

/*
* janus signal business
*
*
 */

type JanusData struct {
	transMap  map[string]chan []byte
	sessionID int64
	handleID  int64
	sdp       chan string
	url       string
	ptype     int
	clientId  uint64
	httpReq   *transports.HttpHandle

	sendData    string
	sendTime    float64
	sendTextLen int
	maxev       int
	connected   bool
	retryCount  int
	status      string
	pc          *webrtc.PeerConnection
	dc          *webrtc.DataChannel

	logCount int
}

const JANUS = "uprtc"
const TRANSACTION = "transaction"
const EVENT = "event"
const SUCCESS = "success"
const TIMEOUT = "timeout"
const ACK = "ack"
const KEEPALIVE = "keepalive"
const PULGIN_BROADCAST = "uprtc.plugin.broadcast"

const CreateHandleSuccess = "createhandled"
const RegisterSuccess = "registered"
const JoinSuccess = "joined"

// go 版本 >1.8.1

// 必须第一个大写
type jsonMsgHeader struct {
	Reqeust  string `json:"request"`
	ClientId uint64 `json:"client_id"`
}
type httpRequestJson struct {
	Uprtc   string        `json:"uprtc"`
	Transac string        `json:"transaction"`
	Body    jsonMsgHeader `json:"body"`
}

type RegisterData struct {
	Plugindata struct {
		Data struct {
			Broadcast string `json:"broadcast"`
			ClientID  int64  `json:"client_id"`
		} `json:"data"`
		Plugin string `json:"plugin"`
	} `json:"plugindata"`
	Sender      int64  `json:"sender"`
	SessionID   int64  `json:"session_id"`
	Transaction string `json:"transaction"`
	Uprtc       string `json:"uprtc"`
}

type joinMsgHeader struct {
	Reqeust     string `json:"request"`
	ClientId    uint64 `json:"client_id"`
	Ptype       string `json:"ptype"`
	SecKey      string `json:"sec_key"`
	DataChannel bool   `json:"datachannel"`
}
type joinRequestJson struct {
	Uprtc   string        `json:"uprtc"`
	Transac string        `json:"transaction"`
	Body    joinMsgHeader `json:"body"`
}

type jsepMsgHeader struct {
	Type string `json:"type"`
	Sdp  string `json:"sdp"`
}
type configureMsgHeader struct {
	Reqeust string `json:"request"`
	Video   bool   `json:"video"`
	Audio   bool   `json:"audio"`
}
type offerRequestJson struct {
	Uprtc   string             `json:"uprtc"`
	Transac string             `json:"transaction"`
	Body    configureMsgHeader `json:"body"`
	Jsep    jsepMsgHeader      `json:"jsep"`
}

type candidateData struct {
	Candidate     string `json:"candidate"`
	SdpMLineIndex int64  `json:"sdpMLineIndex"`
	SdpMid        string `json:"sdpMid"`
}

type trickRequestJson struct {
	Candidate   candidateData `json:"candidate"`
	Transaction string        `json:"transaction"`
	Uprtc       string        `json:"uprtc"`
}

type candidateComplete struct {
	Complete bool `json:"completed"`
}
type trickCompleteJson struct {
	Candidate   candidateComplete `json:"candidate"`
	Transaction string            `json:"transaction"`
	Uprtc       string            `json:"uprtc"`
}

type signalTemplate struct {
	Uprtc       string `json:"uprtc"`
	SrcClientId uint64 `json:"src_client_id"`
	Action      string `json:"action"`
	Body        string `json:"body"`
}

func NewJanusObject(url string, clientid uint64, ptype int, Qps float64) *JanusData {
	var res = new(JanusData)
	res.transMap = make(map[string]chan []byte)
	res.sdp = make(chan string, 1)
	res.url = url
	res.ptype = ptype
	res.clientId = clientid
	res.connected = false
	res.retryCount = 0
	res.logCount = 0
	// qps bps
	res.sendTextLen = 802
	// ms
	res.sendTime = 1000 / float64(Qps*1000/8/800)

	var data string
	var n int
	for n = 0; len(data) < res.sendTextLen; n++ {
		data = data + strconv.Itoa(n)
	}
	//fmt.Println(data)
	fmt.Println("++++++++ send time =%f ms, len=%d n=%d ", res.sendTime, res.sendTextLen, n-1)
	res.sendData = data

	return res
}
func (obj *JanusData) CreateSession() {
	obj.httpReq = transports.NewhttpHandle(obj.url)
	createSig := make(map[string]string)
	createSig[JANUS] = "create"
	createSig[TRANSACTION], _ = randutil.AlphaString(12)
	//json.Marshal(createSig)
	sndData, _ := json.Marshal(createSig)
	fmt.Println("sndData = ", string(sndData))
	resp, err := obj.httpReq.PostRequest(obj.url, sndData)
	if nil != err {
		fmt.Println("creatsession error ", err)
		obj.connected = false
		return
	}
	obj.connected = true
	fmt.Println("creatsession get %s", resp)
	// response := make(chan []byte)

	// obj.transMap[createSig[TRANSACTION]] = response

	// bytes := <-response
	// fmt.Println("Create response: %v", string(bytes))

	// obj.sessionID = gjson.GetBytes(bytes, "data.id").Int()
	// fmt.Println("get sessionid %v", obj.sessionID)
	obj.ProcessRecv([]byte(resp))
}
func (obj *JanusData) CreateHandle() {
	//	obj.httpReq = transports.NewhttpHandle(obj.url)
	attachSig := make(map[string]interface{})
	attachSig[JANUS] = "attach"
	attachSig["session_id"] = obj.sessionID
	attachSig["plugin"] = PULGIN_BROADCAST
	attachSig[TRANSACTION], _ = randutil.AlphaString(12)

	sndData, _ := json.Marshal(attachSig)
	fmt.Println("sndData = ", string(sndData))
	resp, err := obj.httpReq.PostRequest(obj.url, sndData)
	if nil != err {
		fmt.Println("create handle error ", err)
		obj.connected = false
		return
	}
	fmt.Println("create handle get: %s", resp)
	obj.ProcessRecv([]byte(resp))

	// response := make(chan []byte)

	// //	obj.transMap[attachSig[TRANSACTION]] = response

	// bytes := <-response
	// fmt.Println("Create response: %v", string(bytes))

	// obj.sessionID = gjson.GetBytes(bytes, "data.id").Int()
	// fmt.Println("get sessionid %v", obj.sessionID)
}

func (obj *JanusData) EventHandle() {
	// http long poll
	strSession := strconv.FormatInt(obj.sessionID, 10)
	//	fmt.Println("sessionid =", strSession)
	longpoll := obj.url + "/" + strSession + "?rid=" + strconv.FormatInt(time.Now().UnixNano()%1e6/1e3, 10)
	if obj.maxev != 0 {
		longpoll = longpoll + "&maxev=" + string(obj.maxev)
	}
	fmt.Println("longpoll =", longpoll)
	resp, err := obj.httpReq.LongPoll(longpoll)
	if err != nil {
		// is down
		obj.retryCount += 1
		if obj.retryCount > 3 {
			// down
			obj.connected = false
			fmt.Println("gateway is down ", err)
		} else {
			obj.EventHandle()
		}
		return
	}
	if obj.connected == false {
		return
	}
	//fmt.Println("eventhandle get %s", resp)
	obj.ProcessRecv([]byte(resp))
	obj.EventHandle()

}

func (obj *JanusData) Register() {
	//	obj.httpReq = transports.NewhttpHandle(obj.url)
	// registerSig := make(map[string]interface{})
	// registerSig[JANUS] = "message"
	// registerSig["session_id"] = obj.sessionID
	// attachSig["plugin"] = PULGIN_BROADCAST
	// attachSig[TRANSACTION], _ = randutil.AlphaString(12)
	var registerBody jsonMsgHeader
	registerBody.Reqeust = "register"
	registerBody.ClientId = obj.clientId
	// registerBody := make(map[string]interface{})
	// registerBody["request"] = "register"
	// registerBody["client_id"] = 123123

	var registerSig httpRequestJson
	registerSig.Uprtc = "message"
	registerSig.Transac, _ = randutil.AlphaString(12)
	registerSig.Body = registerBody

	//	fmt.Println("begin register %v", registerSig)
	sndData, _ := json.Marshal(registerSig)
	fmt.Println("sndData = ", string(sndData))
	strSession := strconv.FormatInt(obj.sessionID, 10)
	strhandle := strconv.FormatInt(obj.handleID, 10)
	//	fmt.Println("register handle id =", obj.handleID)
	url := obj.url + "/" + strSession + "/" + strhandle
	//	fmt.Println("register post url =", url)
	resp, err := obj.httpReq.PostRequest(url, sndData)
	if nil != err {
		obj.connected = false
		fmt.Println("gateway is down ?", err)
	}
	//fmt.Println("register get %s", resp)
	obj.ProcessRecv([]byte(resp))

}

func (obj *JanusData) Join(role int) {

	var joinBody joinMsgHeader
	joinBody.Reqeust = "join"
	joinBody.ClientId = obj.clientId

	joinBody.SecKey = ""

	if role == 1 {
		// publisher
		joinBody.Ptype = "publisher"
	} else if role == 2 {
		// listner
		joinBody.DataChannel = true
		joinBody.Ptype = "listener"
	}
	var joinSig joinRequestJson
	joinSig.Uprtc = "message"
	joinSig.Transac, _ = randutil.AlphaString(12)
	joinSig.Body = joinBody
	sndData, _ := json.Marshal(joinSig)
	fmt.Println("sndData = ", string(sndData))
	strSession := strconv.FormatInt(obj.sessionID, 10)
	strhandle := strconv.FormatInt(obj.handleID, 10)
	//fmt.Println("sessionid =", obj.handleID)
	url := obj.url + "/" + strSession + "/" + strhandle
	//fmt.Println("join post url =", url)
	resp, err := obj.httpReq.PostRequest(url, sndData)
	if nil != err {
		obj.connected = false
		fmt.Println("gateway is down ? ", err)
	}
	//	fmt.Println("join get %s", resp)
	obj.ProcessRecv([]byte(resp))

}

func (obj *JanusData) SendOffer(sdp string) {

	var confBody configureMsgHeader
	confBody.Reqeust = "configure"
	confBody.Video = false
	confBody.Audio = false

	var jsepBody jsepMsgHeader
	jsepBody.Type = "offer"
	jsepBody.Sdp = sdp
	var offerSig offerRequestJson
	offerSig.Uprtc = "message"
	offerSig.Transac, _ = randutil.AlphaString(12)
	offerSig.Body = confBody
	offerSig.Jsep = jsepBody
	//	fmt.Println("begin join %v", offerSig)
	sndData, _ := json.Marshal(offerSig)
	fmt.Println("sndData = ", string(sndData))
	strSession := strconv.FormatInt(obj.sessionID, 10)
	strhandle := strconv.FormatInt(obj.handleID, 10)
	//	fmt.Println("sessionid =", obj.handleID)
	url := obj.url + "/" + strSession + "/" + strhandle
	//	fmt.Println("join post url =", url)
	resp, err := obj.httpReq.PostRequest(url, sndData)
	if nil != err {
		obj.connected = false
		fmt.Println("gateway is down ", err)
	}
	//	fmt.Println("join get %s", resp)
	obj.ProcessRecv([]byte(resp))
}
func (obj *JanusData) SendAnswer(sdp string) {

	var confBody configureMsgHeader
	confBody.Reqeust = "start"

	var jsepBody jsepMsgHeader
	jsepBody.Type = "answer"
	jsepBody.Sdp = sdp
	var offerSig offerRequestJson
	offerSig.Uprtc = "message"
	offerSig.Transac, _ = randutil.AlphaString(12)
	offerSig.Body = confBody
	offerSig.Jsep = jsepBody
	//	fmt.Println("begin join %v", offerSig)
	sndData, _ := json.Marshal(offerSig)
	fmt.Println("sndData = ", string(sndData))
	strSession := strconv.FormatInt(obj.sessionID, 10)
	strhandle := strconv.FormatInt(obj.handleID, 10)
	//	fmt.Println("sessionid =", obj.handleID)
	url := obj.url + "/" + strSession + "/" + strhandle
	//	fmt.Println("join post url =", url)
	resp, err := obj.httpReq.PostRequest(url, sndData)
	if nil != err {
		obj.connected = false
		fmt.Println("gateway is down ", err)
	}
	//	fmt.Println("join get %s", resp)
	obj.ProcessRecv([]byte(resp))
}

func (obj *JanusData) SendTrick(ic webrtc.IceCandidate) {
	iceCandidate, err := json.Marshal(ic)
	if nil != err {
		fmt.Println("ice candidate marshal error ")
		//todo stop http connect
		return
	}
	//fmt.Println("trickle send data: ", string(iceCandidate))
	var candidate candidateData
	candidate.Candidate = gjson.GetBytes(iceCandidate, "candidate").String()
	candidate.SdpMLineIndex = gjson.GetBytes(iceCandidate, "sdpMLineIndex").Int()
	candidate.SdpMid = gjson.GetBytes(iceCandidate, "sdpMid").String()

	var trickSig trickRequestJson
	trickSig.Uprtc = "trickle"
	trickSig.Transaction, _ = randutil.AlphaString(12)
	trickSig.Candidate = candidate

	sndData, _ := json.Marshal(trickSig)
	fmt.Println("sndData = ", string(sndData))
	strSession := strconv.FormatInt(obj.sessionID, 10)
	strhandle := strconv.FormatInt(obj.handleID, 10)
	//fmt.Println("sessionid =", obj.handleID)
	url := obj.url + "/" + strSession + "/" + strhandle
	//fmt.Println("join post url =", url)
	resp, err := obj.httpReq.PostRequest(url, sndData)
	if nil != err {
		obj.connected = false
		fmt.Println("gateway is down ", err)
	}
	//fmt.Println("join get %s", resp)
	obj.ProcessRecv([]byte(resp))
}

func (obj *JanusData) SendTrickComplete() {

	//fmt.Println("trickle send data: ", string(iceCandidate))
	var candidate candidateComplete
	candidate.Complete = true

	var trickSig trickCompleteJson
	trickSig.Uprtc = "trickle"
	trickSig.Transaction, _ = randutil.AlphaString(12)
	trickSig.Candidate = candidate

	sndData, _ := json.Marshal(trickSig)
	fmt.Println("sndData = ", string(sndData))
	strSession := strconv.FormatInt(obj.sessionID, 10)
	strhandle := strconv.FormatInt(obj.handleID, 10)
	//fmt.Println("sessionid =", obj.handleID)
	url := obj.url + "/" + strSession + "/" + strhandle
	//fmt.Println("join post url =", url)
	resp, err := obj.httpReq.PostRequest(url, sndData)
	if nil != err {
		obj.connected = false
		fmt.Println("gateway is down ", err)
	}
	//fmt.Println("join get %s", resp)
	obj.ProcessRecv([]byte(resp))
}

func (obj *JanusData) SendDataChannelData() {
	go func() {
		var total int
		total = 0
		for {
			if false == obj.connected {
				break
			}
			if total > 1000 {
				tmp := []byte(obj.sendData)
				length := len(obj.sendData)
				fmt.Println("------- send data len=%d end=%d", length, string(tmp[length-3:]))
				total = 0
			}
			obj.dc.SendText(obj.sendData)
			total += 1
			sleepTime := obj.sendTime * 1000
			time.Sleep(time.Duration(sleepTime) * time.Microsecond)
		}
	}()
}

func (obj *JanusData) SendSignalDataChannel() {

	var signalData signalTemplate
	signalData.Uprtc = "datachannel"
	signalData.SrcClientId = obj.clientId
	signalData.Action = "test"
	signalData.Body = ""

	go func() {
		//	var total int
		//total = 0
		for {
			if false == obj.connected {
				break
			}
			// micro time
			signalData.Body = strconv.FormatInt(time.Now().UnixNano()/1000, 10)
			sndData, _ := json.Marshal(signalData)
			fmt.Println("send datachannel sndData = ", string(sndData))
			obj.dc.SendText(string(sndData))
			sleepTime := obj.sendTime * 1000
			time.Sleep(time.Duration(sleepTime) * time.Microsecond)
		}
	}()
}

func (obj *JanusData) ParseSignalDataChannel(msg []byte) {
	n := len(msg)
	fmt.Println("datachannel  receive: %s", string(msg))
	typ := gjson.GetBytes(msg[:n], JANUS).String()
	if typ == "datachannel" {
		srcClientID := gjson.GetBytes(msg[:n], "src_client_id").Int()
		action := gjson.GetBytes(msg[:n], "action").String()
		body := gjson.GetBytes(msg[:n], "body").String()
		fmt.Println("datachannle parse srcclientid=%ld, action=%s,body=%s",
			srcClientID, action, body)
		if action == "test" {
			srcTime, err := strconv.ParseInt(body, 10, 64)
			if err != nil {
				fmt.Println("data channel parse body to int error ")
				return
			}
			currentTime := time.Now().UnixNano() / 1000
			fmt.Println("current time =", currentTime)
			useTime := currentTime - srcTime
			fmt.Println("datachannel use time diff =", useTime)
		}
	}

}

func (obj *JanusData) generateOffer() {
	fmt.Println("Generating offer...")
	offer, err := obj.pc.CreateOffer() // blocking
	if err != nil {
		fmt.Println(err)
		return
	}
	obj.pc.SetLocalDescription(offer)
	//fmt.Println("type sdp offer ", offer)
}

// Attach callbacks to a newly created data channel.
// In this demo, only one data channel is expected, and is only used for chat.
// But it is possible to send any sort of bytes over a data channel, for many
// more interesting purposes.

func (obj *JanusData) prepareDataChannel(channel *webrtc.DataChannel) {
	channel.OnOpen = func() {
		fmt.Println("Data Channel Opened!")
		// for {
		// 	obj.dc.SendText("hello, world")
		// 	fmt.Println("Data Channel sent!")
		// 	time.Sleep(1 * time.Second)
		// }
		//startChat()
		if 1 == obj.ptype {
			//obj.SendDataChannelData()
			obj.SendSignalDataChannel()
		}

	}
	channel.OnClose = func() {
		fmt.Println("Data Channel closed.")
		//endChat()
	}
	channel.OnMessage = func(msg []byte) {
		//fmt.Println("Data Channel message received.")
		obj.logCount += 1
		if obj.logCount > 1000 {
			length := len(msg)
			fmt.Println("receive from datachannel:len=%d end=%d", length, string(msg[length-3:]))
			obj.logCount = 0
		}
		obj.ParseSignalDataChannel(msg)

		//receiveChat(string(msg))
	}
}

func (obj *JanusData) CreateWebrtcConnection() {
	webrtc.SetLoggingVerbosity(1)
	// TODO: Try with TURN servers.
	config := webrtc.NewConfiguration(webrtc.OptionIceServer("stun:stun.ekiga.net:3478"))
	var err error
	obj.pc, err = webrtc.NewPeerConnection(config)
	if nil != err {
		fmt.Println("Failed to create PeerConnection.")
		//todo close janus http
		return
	}
	// OnNegotiationNeeded is triggered when something important has occurred in
	// the state of PeerConnection (such as creating a new data channel), in which
	// case a new SDP offer must be prepared and sent to the remote peer.
	obj.pc.OnNegotiationNeeded = func() {
		if 1 == obj.ptype {
			go obj.generateOffer()
		}
	}

	obj.pc.OnIceCandidate = func(ic webrtc.IceCandidate) {
		//fmt.Println("onicecandidate gathering ICE candidates %v", ic)
		//todo send trick
		obj.SendTrick(ic)
	}
	// Once all ICE candidates are prepared, they need to be sent to the remote
	// peer which will attempt reaching the local peer through NATs.
	obj.pc.OnIceComplete = func() {
		fmt.Println("Finished gathering ICE candidates.")
		sdp := obj.pc.LocalDescription().Serialize()
		obj.sdp <- sdp
		//send offer to server
		//	fmt.Println("send sdp offer ", sdp)
		sdpContent := gjson.Get(sdp, "sdp").String()
		//fmt.Println("generated sdp: %v", sdpContent)
		if 1 == obj.ptype {
			obj.SendOffer(sdpContent)
		} else if 2 == obj.ptype {
			obj.SendAnswer(sdpContent)
		}
		obj.SendTrickComplete()
	}
	fmt.Println("Initializing datachannel....")
	obj.dc, err = obj.pc.CreateDataChannel("JanusDataChannel", webrtc.Init{Ordered: false})
	if nil != err {
		fmt.Println("Unexpected failure creating Channel.")
		obj.connected = false
		return
	}
	obj.prepareDataChannel(obj.dc)
}
func (obj *JanusData) generateAnswer() {
	fmt.Println("Generating answer...")
	answer, err := obj.pc.CreateAnswer() // blocking
	if err != nil {
		fmt.Println(err)
		return
	}
	obj.pc.SetLocalDescription(answer)
}

func (obj *JanusData) SetRemoteSdp(answerJsep string) {
	answerSDP := webrtc.DeserializeSessionDescription(answerJsep)
	if answerSDP != nil {
		err := obj.pc.SetRemoteDescription(answerSDP)
		if err != nil {
			fmt.Println("ERROR: %v", err)
			return
		}
		fmt.Println("SDP " + answerSDP.Type + " successfully received.")
	}
}

// 需要传指针对象进去
func (obj *JanusData) StartPlugin() {
	if obj.status == CreateHandleSuccess {
		if obj.ptype == 1 {
			// publisher
			obj.Register()
		} else if 2 == obj.ptype {
			obj.Join(obj.ptype)
			obj.CreateWebrtcConnection()
		} else {
			fmt.Println("unknown action to do")
		}
	} else if obj.status == RegisterSuccess {
		obj.Join(1)
	} else if obj.status == JoinSuccess {
		if 1 == obj.ptype {
			// publisher
			obj.CreateWebrtcConnection()
		} else if 2 == obj.ptype {
			//listener
			// waitfor jsep offer
		}

	}

}

func (obj *JanusData) ProcessRecv(buffer []byte) {
	//	buffer := make([]byte, 10240)
	n := len(buffer)
	//	for {
	obj.retryCount = 0

	typ := gjson.GetBytes(buffer[:n], JANUS).String()
	if typ == EVENT {
		fmt.Println("Signal receive: %s", string(buffer))
		// sessionID := gjson.GetBytes(buffer[:n], "session_id").Int()
		// transaction := gjson.GetBytes(buffer[:n], TRANSACTION).String()
		// sender := gjson.GetBytes(buffer[:n], "sender").Int()
		//	fmt.Println("event receive: %s", sender)
		//pluginData := gjson.GetBytes(buffer[:n], "plugindata").String()
		// if pluginData != "" {
		// 	// has data
		// 	plugin := gjson.GetBytes(buffer[:n], "plugindata.plugin").String()
		// //	fmt.Println("get plugin", plugin)
		// }
		respone := gjson.GetBytes(buffer[:n], "plugindata.data.broadcast").String()
		if respone == "registered" {
			// registered
			rclientID := gjson.GetBytes(buffer[:n], "plugindata.data.client_id").Int()
			if rclientID != 0 {
				fmt.Println("get register and handleid", rclientID, obj.handleID)
				obj.status = RegisterSuccess
				obj.StartPlugin()
			}
		} else if respone == "joined" {
			obj.status = JoinSuccess
			obj.StartPlugin()
		} else if respone == "event" {
			result := gjson.GetBytes(buffer[:n], "plugindata.data.error").String()
			if result != "" {
				// get error
				obj.connected = false
				fmt.Println("error get uprtc", result)
			}
		}
		jsep := gjson.GetBytes(buffer[:n], "jsep").String()
		if jsep != "" {
			fmt.Println("get jsep data: %s", jsep)
			// have jsep
			jsepType := gjson.GetBytes(buffer[:n], "jsep.type").String()
			if jsepType == "answer" {
				// answer
				obj.SetRemoteSdp(jsep)
			} else if jsepType == "offer" {
				// offer
				obj.SetRemoteSdp(jsep)
				go obj.generateAnswer()
			}

		}

		//fmt.Println("type: %v, sessionID: %v, transaction: %v, sender: %v", typ, sessionID, transaction, sender)
		// if ch, ok := obj.transMap[transaction]; ok {
		// 	ch <- buffer[:n]
		// }
	} else if typ == SUCCESS {
		fmt.Println("Signal receive: %s", string(buffer))
		//	transaction := gjson.GetBytes(buffer[:n], TRANSACTION).String()
		getSessionID := gjson.GetBytes(buffer[:n], "session_id").Int()
		sessionID := gjson.GetBytes(buffer[:n], "data.id").Int()
		if getSessionID != 0 {
			// response has sessionid
			if sessionID != 0 {
				fmt.Println("************** set handle id is ", sessionID)
				obj.handleID = sessionID
			}

			// set status
			// createhandle
			obj.status = CreateHandleSuccess
			obj.StartPlugin()
		} else {
			// first create session response
			fmt.Println("************** set session id is ", sessionID)
			obj.sessionID = sessionID

			go func() {
				obj.EventHandle()
			}()
			//	go obj.EventHandle()
			obj.CreateHandle()
		}

		//fmt.Println(" success type: %v, sessionID: %v, transaction: %v", typ, sessionID, transaction)
		// if ch, ok := obj.transMap[transaction]; ok {
		// 	ch <- buffer[:n]
		// }
		//	fmt.Println(" success type: %v", obj.transMap[transaction])

	} else if typ == TIMEOUT {
		fmt.Println("Signal receive: %s", string(buffer))
		// sessionID := gjson.GetBytes(buffer[:n], "session_id").Int()
		// fmt.Println("timeout session_id: %v", sessionID)
	} else if typ == ACK {
		fmt.Println("Signal receive: %s", string(buffer))
		//	transaction := gjson.GetBytes(buffer[:n], TRANSACTION).String()

		//	fmt.Println("type: %v, transaction: %v", typ, transaction)
		// if ch, ok := obj.transMap[transaction]; ok {
		// 	ch <- buffer[:n]
		// }
	} else if typ == KEEPALIVE {
		obj.retryCount = 0
	} else if typ == "webrtcup" {
		fmt.Println("Signal receive: %s", string(buffer))
		obj.retryCount = 0
	} else
	/*if typ == EVENT {
		transaction := gjson.GetBytes(buffer[:n], TRANSACTION).String()

		fmt.Println("type: %v, transaction: %v", typ, transaction)
		if ch, ok := obj.transMap[transaction]; ok {
			ch <- buffer[:n]
		}
	} else */
	{
		fmt.Println("Signal receive: %s", string(buffer))
		fmt.Println("receive respone unkwon !!!!")
		// transaction := gjson.GetBytes(buffer[:n], TRANSACTION).String()

		// fmt.Println("type: %v, transaction: %v", typ, transaction)
		// if ch, ok := obj.transMap[transaction]; ok {
		// 	ch <- buffer[:n]
		// }
	}
	//	}
}
