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

	"github.com/lunixbochs/struc"
	//"os"
)

const currentVersion = "0.01"

func connect() {
	var tcpAddr = "192.168.115.231:4737"
	var hdrObj = ipdrlib.IPDRStreamingHeaderIdl{
		Version:      2,
		MessageID:    5,
		SessionID:    0,
		MessageFlags: 0,
		MessageLen:   29,
	}
	var conn, err = net.Dial("tcp", tcpAddr)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(conn.LocalAddr())
	var ipport = strings.Split(conn.LocalAddr().String(), ":")
	var ipStrList = strings.Split(ipport[0], ".")
	var ipNbrList [4]int
	for i := 0; i < len(ipStrList); i++ {
		var str = ipStrList[i]
		var octet, erroctet = strconv.ParseInt(str, 10, 0)
		if erroctet != nil {
			fmt.Println("Octet failed to be parsed)")
			return
		}
		ipNbrList[i] = int(octet)
	}
	var port, ipParseErr = strconv.ParseInt(ipport[1], 10, 0)
	if ipParseErr != nil {
		fmt.Println("Parse of port number failed")
		return
	}
	fmt.Println(port)
	fmt.Println("--------------------------------")
	var connectObj = ipdrlib.ConnectIdl{
		CollectorAddress:  uint32(ipNbrList[0]<<24 + ipNbrList[1]<<16 + ipNbrList[2]<<8 + ipNbrList[3]),
		CollectorPort:     uint16(port),     //uint16(port)
		Capabilities:      1,
		KeepAliveInterval: 60,
		VendorID:          "IPDRGO",
	}
	result := ipdrlib.Connect(hdrObj, connectObj)
	conn.Write(result.Bytes())
	for {

		//Load Header
		response := make([]byte, 8)
		var hdr ipdrlib.IPDRStreamingHeaderIdl

		r := bufio.NewReader(conn)
		io.ReadFull(r, response)
		struc.Unpack(bytes.NewBuffer(response), &hdr)

		fmt.Println(hdr.MessageLen)
		var remainingMsgLen = hdr.MessageLen - 8
		response2 := make([]byte, remainingMsgLen)
		io.ReadFull(r, response2)
		ipdrlib.ParseMessageByType(bytes.NewBuffer(response2), hdr.MessageID, hdr.MessageLen)

		break

	}
}
func main() {
	fmt.Println("ipdrgo ", currentVersion)
	connect()

}
