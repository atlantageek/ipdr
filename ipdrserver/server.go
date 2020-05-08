package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"ipdrlib"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
	"xdrlib"

	//"github.com/prprprus/scheduler"
	//"runtime"

	"github.com/lunixbochs/struc"

	//"time"
	"net/http"
)

type sessionTrack struct {
	xdrf        *os.File
	sessionID   uint8
	configID    uint16
	seq         uint64
	sessionName string
}

type responseResult struct {
	Result    interface{}
	MessageID uint8
}

var devices map[string]map[string]string
var xdrf *os.File
var configID uint16 = 0

var sessionMap = make(map[uint8]sessionTrack)

const currentVersion = "0.01"

var r io.Reader = nil
var templateMap = make(map[uint8][]ipdrlib.FieldDescriptorIdl)

func getResponse(conn net.Conn) (interface{}, ipdrlib.IPDRStreamingHeaderIdl) {
	response := make([]byte, 8)
	var messageID uint8 = 0x40
	var hdr ipdrlib.IPDRStreamingHeaderIdl

	for messageID == 0x40 {
		var nbr, err = io.ReadFull(r, response)
		if nbr == 0 && err == io.EOF {
			return nil, hdr
		} else if err != nil {

			panic(fmt.Sprintf("xxError:%d\n", err))
		} else if nbr < 8 {
			panic("Not enough bytes read")
		}
		struc.Unpack(bytes.NewBuffer(response), &hdr)
		messageID = hdr.MessageID
		if messageID == 0x40 {
			keepAlive(conn)
		}
	}
	var remainingMsgLen = hdr.MessageLen - 8
	response2 := make([]byte, remainingMsgLen)
	io.ReadFull(r, response2)

	var result = ipdrlib.ParseMessageByType(bytes.NewBuffer(response2), hdr.MessageID, hdr.MessageLen)

	return result, hdr
}

func getMultipleResponses(conn net.Conn) []responseResult {
	results := make([]responseResult, 0)
	var moreBytes bool = true
	for moreBytes == true {
		result, hdr := getResponse(conn)
		r := responseResult{result, hdr.MessageID}
		results = append(results, r)
		_, err := bufio.NewReader(conn).Peek(1)
		if err == bufio.ErrBufferFull {
			moreBytes = false
		}
	}
	return results
}
func connect(conn net.Conn) bool {
	devices = make(map[string]map[string]string)
	var hdrObj = ipdrlib.IPDRStreamingHeaderIdl{
		Version:      2,
		MessageID:    5,
		SessionID:    0,
		MessageFlags: 0,
		MessageLen:   29,
	}

	var ipport = strings.Split(conn.LocalAddr().String(), ":")
	var ipStrList = strings.Split(ipport[0], ".")
	var ipNbrList [4]int
	for i := 0; i < len(ipStrList); i++ {
		var str = ipStrList[i]
		var octet, erroctet = strconv.ParseInt(str, 10, 0)
		if erroctet != nil {
			fmt.Println("Octet failed to be parsed)")
			return false
		}
		ipNbrList[i] = int(octet)
	}
	var port, ipParseErr = strconv.ParseInt(ipport[1], 10, 0)
	if ipParseErr != nil {
		fmt.Println("Parse of port number failed")
		return false
	}
	var connectObj = ipdrlib.ConnectIdl{
		CollectorAddress:  uint32(ipNbrList[0]<<24 + ipNbrList[1]<<16 + ipNbrList[2]<<8 + ipNbrList[3]),
		CollectorPort:     uint16(port), //uint16(port)
		Capabilities:      1,
		KeepAliveInterval: 60,
		VendorID:          "IPDRGO",
	}
	result := ipdrlib.Connect(hdrObj, connectObj)
	conn.Write(result.Bytes())

	//responseObj := nil
	//Load Header
	return true

}
func getSessions(conn net.Conn) {
	var hdrObj = ipdrlib.IPDRStreamingHeaderIdl{
		Version:      2,
		MessageID:    0x14,
		SessionID:    0,
		MessageFlags: 0,
		MessageLen:   10,
	}
	var getSessionsObj = ipdrlib.GetSessionsIdl{
		RequestID: 20,
	}

	result := ipdrlib.GetSessions(hdrObj, getSessionsObj)

	conn.Write(result.Bytes())

}
func flowStart(conn net.Conn, sessionID uint8) {

	var hdrObj = ipdrlib.IPDRStreamingHeaderIdl{
		Version:      2,
		MessageID:    1,
		SessionID:    sessionID,
		MessageFlags: 0,
		MessageLen:   8,
	}
	result := ipdrlib.Hdr(hdrObj)
	conn.Write(result.Bytes())

}
func finalTemplateData(conn net.Conn, sessionID uint8) bool {

	var hdrObj = ipdrlib.IPDRStreamingHeaderIdl{
		Version:      2,
		MessageID:    0x13,
		SessionID:    sessionID,
		MessageFlags: 0,
		MessageLen:   8,
	}
	result := ipdrlib.Hdr(hdrObj)
	conn.Write(result.Bytes())

	return true

}

func keepAlive(conn net.Conn) {
	var hdrObj = ipdrlib.IPDRStreamingHeaderIdl{
		Version:      2,
		MessageID:    0x40,
		SessionID:    0,
		MessageFlags: 0,
		MessageLen:   8,
	}
	result := ipdrlib.Hdr(hdrObj)
	conn.Write(result.Bytes())
}
func dataAck(conn net.Conn, sequenceNum uint64, sessionID uint8) {
	var hdrObj = ipdrlib.IPDRStreamingHeaderIdl{
		Version:      2,
		MessageID:    0x21,
		SessionID:    sessionID,
		MessageFlags: 0,
		MessageLen:   18,
	}
	var dataAckObj = ipdrlib.DataAckIdl{
		ConfigID:    configID,
		SequenceNum: sequenceNum,
	}

	result := ipdrlib.DataAck(hdrObj, dataAckObj)

	conn.Write(result.Bytes())
}
func checkDataAvailable(conn net.Conn) {
	dataObj, msgHdr := getResponse(conn)

	currentTime := time.Now()
	fmt.Println("Response Message Type:", msgHdr.MessageID, msgHdr.MessageLen, currentTime.String())
	if msgHdr.MessageID == ipdrlib.ConnectResponseMsgType {
		getSessions(conn)

	} else if msgHdr.MessageID == ipdrlib.GetSessionsResponseMsgType {
		var firstSession uint8 = 255
		sessions := dataObj.(ipdrlib.GetSessionsResponseIdl)
		//sessionID = sessions.SessionBlocks[1].SessionID
		for i := 0; i < len(sessions.SessionBlocks); i++ {

			sess := sessions.SessionBlocks[i]
			if firstSession == 255 {
				firstSession = sess.SessionID
			}
			if sess.SessionType == 1 || sess.SessionType == 4 {
				sessionMap[sess.SessionID] =
					sessionTrack{
						sessionID:   sess.SessionID,
						seq:         0,
						sessionName: sess.SessionName,
					}
				fmt.Println("Session: ", sess.SessionType, sess.SessionName, sess.SessionID)
				flowStart(conn, sessionMap[sess.SessionID].sessionID)
			}
			// fmt.Println("Session: ", sess.SessionType, sess.SessionName, sess.SessionID)
			// flowStart(conn, sessionMap[firstSession].sessionID)
		}
		//Start flow on the first session in the session List
		fmt.Println(sessionMap)
		//flowStart(conn, sessionMap[firstSession].sessionID)
	} else if msgHdr.MessageID == ipdrlib.TemplateDataMsgType {
		data := dataObj.(ipdrlib.TemplateDataIdl)
		configID = data.ConfigID

		templateMap[msgHdr.SessionID] = data.ResultTemplates[0].FieldDescriptors
		finalTemplateData(conn, msgHdr.SessionID)
	} else if msgHdr.MessageID == ipdrlib.SessionStartMsgType {
		data := dataObj.(ipdrlib.SessionStartIdl)
		fname := fmt.Sprintf("/tmp/doc_%s_%x.xdr", sessionMap[msgHdr.SessionID].sessionName, data.DocumentID)

		var err error
		xdrf, err = os.Create(fname)
		if err != nil {
			fmt.Println(err)
		}
		keepAlive(conn)
	} else if msgHdr.MessageID == ipdrlib.SessionStopMsgType {
		keepAlive(conn)
		xdrf.Sync()
		tempSession := sessionMap[msgHdr.SessionID]
		tempSession.seq = 0
		sessionMap[msgHdr.SessionID] = tempSession
	} else if msgHdr.MessageID == ipdrlib.KeepAliveMsgType {
		keepAlive(conn)
	} else if msgHdr.MessageID == ipdrlib.SessionStopMsgType {
		keepAlive(conn)
	} else if msgHdr.MessageID == ipdrlib.DataMsgType {
		data := dataObj.(ipdrlib.FullDataIdl)

		//if deadline.Before(time.Now()) {
		if sessionMap[msgHdr.SessionID].seq == data.SequenceNum {
			dataAck(conn, sessionMap[msgHdr.SessionID].seq, msgHdr.SessionID)
			xdrf.Write(data.Data)
			result, resultStr := xdrlib.ParseData(templateMap[msgHdr.SessionID], data.Data)
			macAddr, ok := result["CmMacAddr"]
			if ok {
				devices[macAddr] = result
			}
			fmt.Println(result)
			fmt.Println(resultStr)
		}

		tempSession := sessionMap[msgHdr.SessionID]
		tempSession.seq++
		sessionMap[msgHdr.SessionID] = tempSession

	} else if msgHdr.MessageID == ipdrlib.ErrorMsgType {
		data := dataObj.(ipdrlib.ErrorResponseIdl)
		fmt.Println("**************ERROR***********************************")
		fmt.Println(data.Description)
	} else {
		fmt.Println("Did Not Handle:", msgHdr.MessageID)
	}

}
func index(w http.ResponseWriter, _ *http.Request) {
	io.WriteString(w, "<html><body>Hello Wurld")
	io.WriteString(w, "<table border=\"1\"><tr><th>Mac</th> <th>IPv4</th><th>IPv6</th><th>CM Last Registration Time</th><th>Reg Status</th><th>Service Packets Passed</th><th>Service SLA Packets Delayed</th><th>Service Packets Dropped</th></tr>")
	for k, v := range devices {
		fmt.Printf("key[%s] value[%s]\n", k, v)
		ipv6 := v["CmIpv6Addr"][0:4] + ":" + v["CmIpv6Addr"][4:8] + ":" + v["CmIpv6Addr"][8:12] + ":" + v["CmIpv6Addr"][12:16] + ":" + v["CmIpv6Addr"][16:20] + ":" + v["CmIpv6Addr"][20:24] + ":" + v["CmIpv6Addr"][24:28] + ":" + v["CmIpv6Addr"][28:32]
		fmt.Fprintf(w, "<tr><td> %s</td> <td> %s </td><td> %s </td><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>", k, v["CmIpv4Addr"], ipv6, v["CmLastRegTime"], v["CmRegStatusValue"], v["ServicePktsPassed"], v["ServiceSlaDelayPkts"], v["ServiceSlaDropPkts"])
	}
	io.WriteString(w, "</table></body></html>")
}
func ws() {
	http.HandleFunc("/sessions", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, sessionMap)
	})
	http.HandleFunc("/devices", func(w http.ResponseWriter, r *http.Request) {
		jsonString, err := json.Marshal(devices)
		if err != nil {
			fmt.Fprint(w, err)
		} else {
			fmt.Fprint(w, string(jsonString))
		}
	})
	http.HandleFunc("/", index)
	http.ListenAndServe(":8081", nil)
}
func main() {
	// s, schedulerErr := scheduler.NewScheduler(1000)
	// if schedulerErr != nil  {
	// 	panic(schedulerErr)
	// }
	go ws()
	var tcpAddr = "192.168.115.231:4737"

	var conn, err = net.Dial("tcp", tcpAddr)
	if err != nil {
		fmt.Println(err)
		return
	}
	connect(conn)

	r = bufio.NewReader(conn)
	for {
		checkDataAvailable(conn)
	}
}
