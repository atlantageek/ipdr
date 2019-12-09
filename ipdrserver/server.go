package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"ipdrlib"
	"net"
	"strconv"
	"strings"
	"time"

	//"github.com/prprprus/scheduler"
	"runtime"

	"github.com/lunixbochs/struc"
	//"time"
)

var sessionID uint8 = 0
var configID uint16 = 0

const currentVersion = "0.01"

func getResponse(conn net.Conn) (interface{}, uint8) {
	response := make([]byte, 8)
	var messageID uint8 = 0x40
	var hdr ipdrlib.IPDRStreamingHeaderIdl

	r := bufio.NewReader(conn)
	for messageID == 0x40 {
		var nbr, err = io.ReadFull(r, response)
		if err != nil {
			panic(fmt.Sprintf("Error:%d\n", err))
		}
		if nbr < 8 {
			panic("Not enough bytes read")
		}
		struc.Unpack(bytes.NewBuffer(response), &hdr)
		messageID = hdr.MessageID
		if messageID == 0x40 {
			fmt.Println("Keep Alive")
			keepAlive(conn)
		}
	}
	var remainingMsgLen = hdr.MessageLen - 8

	response2 := make([]byte, remainingMsgLen)
	io.ReadFull(r, response2)
	var result = ipdrlib.ParseMessageByType(bytes.NewBuffer(response2), hdr.MessageID, hdr.MessageLen)
	return result, hdr.MessageID
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
		MessageLen:   8,
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
	dataObj, msgType := getResponse(conn)
	currentTime := time.Now()
	fmt.Println("Response Message Type:", msgType, currentTime.String())
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
		keepAlive(conn)
	} else if msgType == ipdrlib.KeepAliveMsgType {
		keepAlive(conn)
	} else if msgType == ipdrlib.SessionStopMsgType {
		keepAlive(conn)
	} else if msgType == ipdrlib.DataMsgType {
		data := dataObj.(ipdrlib.DataIdl)
		dataAck(conn, data.SequenceNum)
		fmt.Println("Data Is here :", data)

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
	for {
		checkDataAvailable(conn)
	}
}
