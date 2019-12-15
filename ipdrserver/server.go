package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"ipdrlib"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
	

	//"github.com/prprprus/scheduler"
	"runtime"

	"github.com/lunixbochs/struc"
	//"time"
)
type sessionTrack struct {
	xdrf *os.File
	sessionID uint8 
	configID uint16 
	seq uint64
}

type responseResult struct {
	Result    interface{}
	MessageID uint8
}
var xdrf *os.File
var sessionID uint8 = 0
var configID uint16 = 0
var deadline time.Time = time.Now()
var seq uint64 = 0

const currentVersion = "0.01"

var r io.Reader = nil

func getResponse(conn net.Conn) (interface{}, uint8, uint32) {
	fmt.Println("-----------------------------------------")
	response := make([]byte, 8)
	var messageID uint8 = 0x40
	var hdr ipdrlib.IPDRStreamingHeaderIdl

	for messageID == 0x40 {
		var nbr, err = io.ReadFull(r, response)
		if nbr == 0 && err == io.EOF {
			return nil, uint8(99), 8
		} else if err != nil {

			panic(fmt.Sprintf("xxError:%d\n", err))
		} else if nbr < 8 {
			panic("Not enough bytes read")
		}
		struc.Unpack(bytes.NewBuffer(response), &hdr)
		fmt.Println(response)
		messageID = hdr.MessageID
		if messageID == 0x40 {
			fmt.Println("Keep Alive")
			keepAlive(conn)
		}
	}
	var remainingMsgLen = hdr.MessageLen - 8
	fmt.Println("Remaining message len:", remainingMsgLen)
	response2 := make([]byte, remainingMsgLen)
	io.ReadFull(r, response2)
	fmt.Println("Response2:", response2)
	var result = ipdrlib.ParseMessageByType(bytes.NewBuffer(response2), hdr.MessageID, hdr.MessageLen)

	return result, hdr.MessageID, hdr.MessageLen
}

func getMultipleResponses(conn net.Conn) []responseResult {
	results := make([]responseResult, 0)
	var moreBytes bool = true
	for moreBytes == true {
		result, messageID, _ := getResponse(conn)
		r := responseResult{result, messageID}
		results = append(results, r)
		_, err := bufio.NewReader(conn).Peek(1)
		if err == bufio.ErrBufferFull {
			moreBytes = false
		}
	}
	return results
}
func connect(conn net.Conn) bool {
	fmt.Println("Sending Connect")
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
	fmt.Println("Sending Get Sessions")
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

	fmt.Println(result)
	conn.Write(result.Bytes())

}
func flowStart(conn net.Conn, sessionID uint8) {
	fmt.Println("Sending Flow Start")

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
	fmt.Println("SendingFinal Template")

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
	fmt.Println("Sending Keep Alive")
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
func dataAck(conn net.Conn, sequenceNum uint64) {
	fmt.Println("Sending data Ack")
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
	fmt.Println("Config ID:", configID)
	fmt.Println("Sequence Num:", sequenceNum)
	result := ipdrlib.DataAck(hdrObj, dataAckObj)

	conn.Write(result.Bytes())
}
func checkDataAvailable(conn net.Conn) {
	dataObj, msgType, msgLen := getResponse(conn)

	currentTime := time.Now()
	fmt.Println("Response Message Type:", msgType, msgLen, currentTime.String())
	if msgType == ipdrlib.ConnectResponseMsgType {
		getSessions(conn)

	} else if msgType == ipdrlib.GetSessionsResponseMsgType {
		sessions := dataObj.(ipdrlib.GetSessionsResponseIdl)
		sessionID = sessions.SessionBlocks[1].SessionID
		fmt.Println("SESSIONS AVAIL")
		for i := 0; i < len(sessions.SessionBlocks); i++ {
			sess := sessions.SessionBlocks[i]
			fmt.Println("Session ID", sess.SessionType, sess.SessionName, sess.SessionID)
		}
		flowStart(conn, sessionID)
	} else if msgType == ipdrlib.TemplateDataMsgType {
		data := dataObj.(ipdrlib.TemplateDataIdl)
		configID = data.ConfigID
		fmt.Println("Config id:", configID)
		finalTemplateData(conn, sessionID)
	} else if msgType == ipdrlib.SessionStartMsgType {
		data := dataObj.(ipdrlib.SessionStartIdl)
		fname := fmt.Sprintf("/tmp/doc-%x.xdr", data.DocumentID)
		
		var err error
		xdrf, err = os.Create(fname)
		if err!= nil {
			fmt.Println(err)
		}
		keepAlive(conn)
	} else if msgType == ipdrlib.SessionStopMsgType {
		keepAlive(conn)
		xdrf.Sync()
		
		seq = 0
	} else if msgType == ipdrlib.KeepAliveMsgType {
		keepAlive(conn)
	} else if msgType == ipdrlib.SessionStopMsgType {
		keepAlive(conn)
	} else if msgType == ipdrlib.DataMsgType {
		data := dataObj.(ipdrlib.FullDataIdl)
		fmt.Println("*************************************************")
		fmt.Println(data.Data)
		//if deadline.Before(time.Now()) {
		if seq == data.SequenceNum {
			dataAck(conn, seq)
			xdrf.Write(data.Data)
		}
		seq++

		//	deadline = time.Now().Add(time.Second * 3)
		//	fmt.Println("Update:",time.Now(), deadline)
		//}
		//data := dataObj.(ipdrlib.DataIdl)
		// _, err := bufio.NewReader(conn).Peek(1)
		// if err == bufio.ErrBufferFull {
		// 	dataAck(conn, data.SequenceNum)
		// 	fmt.Println("Data Is here with ACK! :", data)
		// } else {
		// 	fmt.Println("Data Is here :", data)
		// }

	} else {
		fmt.Println("Did Not Handle:", msgType)
	}

}
func main() {
	// s, schedulerErr := scheduler.NewScheduler(1000)
	// if schedulerErr != nil  {
	// 	panic(schedulerErr)
	// }
	var tcpAddr = "192.168.115.231:4737"
	fmt.Println("ipdrgo ", currentVersion)
	fmt.Println(runtime.NumCPU())
	var conn, err = net.Dial("tcp", tcpAddr)
	if err != nil {
		fmt.Println(err)
		return
	}
	connect(conn)
	fmt.Println("Running checkDataAvailable")
	r = bufio.NewReader(conn)
	for {
		checkDataAvailable(conn)
	}
}
