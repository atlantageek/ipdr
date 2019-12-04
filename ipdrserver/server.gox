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
	"github.com/prprprus/scheduler"
	"github.com/lunixbochs/struc"
	"runtime"
	"time"
)

const currentVersion = "0.01"
func getResponse(conn net.Conn) (interface{},uint8) {
	response := make([]byte, 8)
	var hdr ipdrlib.IPDRStreamingHeaderIdl

	r := bufio.NewReader(conn)
	io.ReadFull(r, response)
	struc.Unpack(bytes.NewBuffer(response), &hdr)

	var remainingMsgLen = hdr.MessageLen - 8
	response2 := make([]byte, remainingMsgLen)
	io.ReadFull(r, response2)
	var result = ipdrlib.ParseMessageByType(bytes.NewBuffer(response2), hdr.MessageID, hdr.MessageLen)
	return result, hdr.MessageID
}
func connect(conn net.Conn) bool {
	fmt.Println("--Connect")
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
	//Load Header
	responseObj, messageID := getResponse(conn)
	fmt.Println( responseObj.(ipdrlib.ConnectResponseIdl).KeepAliveInterval)

	if messageID == 6 {
		return true
	} else {
		return false
	}
}
func getSessions(conn net.Conn) []uint8 {
	fmt.Println("--Get Sessions")
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
	//Load Header
	resultObj, messageID := getResponse(conn)

	if messageID == 0x15 {
		sessionBlocks := resultObj.(ipdrlib.GetSessionsResponseIdl).SessionBlocks
		result := make([]uint8,len(sessionBlocks))
		for i:=0; i<len(sessionBlocks);i++ {
			fmt.Println(sessionBlocks[i].SessionName)
			result[i] = sessionBlocks[i].SessionID
		}
		return result
	}
	return make([]uint8, 0)
}
func flowStart(conn net.Conn, sessionID uint8) bool {
	fmt.Println("--Flow Start")

	var hdrObj = ipdrlib.IPDRStreamingHeaderIdl{
		Version:      2,
		MessageID:    1,
		SessionID:    sessionID,
		MessageFlags: 0,
		MessageLen:   8,
	}
	result := ipdrlib.Hdr(hdrObj)
	conn.Write(result.Bytes())

	//Load Header
	_, messageID := getResponse(conn)
	if (messageID == 0x10){
		return true
	} else {
		return false
	}
}
func finalTemplateData(conn net.Conn, sessionID uint8) bool {
	fmt.Println("--Final Template")

	var hdrObj = ipdrlib.IPDRStreamingHeaderIdl{
		Version:      2,
		MessageID:    0x13,
		SessionID:    sessionID,
		MessageFlags: 0,
		MessageLen:   8,
	}
	result := ipdrlib.Hdr(hdrObj)
	conn.Write(result.Bytes())

	//Load Header
	resultObj, messageID := getResponse(conn)
	fmt.Println(resultObj)
	if (messageID == 0x08){
		return true
	} else {
		return false
	}
}
func keepAlive2() {
	fmt.Println("Keep alive 2")
}
func keepAlive(conn net.Conn) {
	fmt.Println("Keep Alive")
	// var hdrObj = ipdrlib.IPDRStreamingHeaderIdl{
	// 	Version:      2,
	// 	MessageID:    0x40,
	// 	SessionID:    0,
	// 	MessageFlags: 0,
	// 	MessageLen:   8,
	// }
	// result := ipdrlib.Hdr(hdrObj)
	// conn.Write(result.Bytes())
}
func main() {
	s, schedulerErr := scheduler.NewScheduler(1000)
	if schedulerErr != nil  {
		panic(schedulerErr)
	}
	var tcpAddr = "192.168.115.231:4737"
	fmt.Println("ipdrgo ", currentVersion)
	fmt.Println(runtime.NumCPU())
	var conn, err = net.Dial("tcp", tcpAddr)
	if err != nil {
		fmt.Println(err)
		return
	}
	if connect(conn) {
		fmt.Println("Do Every second")
		s.Every().Second(1).Do(keepAlive2)
		sessions := getSessions(conn)
		for i:=0;i<len(sessions);i++ {
			if flowStart(conn,sessions[i]) {
				time.Sleep(5 * time.Second)
				finalTemplateData(conn, sessions[i])
			}
		}
		
	}

}
