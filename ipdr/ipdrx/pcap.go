// Simple go application that opens a pcap file and print the packets

package main

import (
	"bytes"
	"github.com/lunixbochs/struc" 
	"fmt"
	"log"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

// IPDRStreamingHeader : here you tell us what Salutation is

type headerIPDR struct {
	Version      uint8
	MessageID    uint8
	SessionID    uint8
	MessageFlags uint8
	MessageLen   uint32
}
type connectIPDR struct {

	IndicatorID       uint32
	InitiatorPort     uint16
	Capabilities      uint32
	KeepAliveInterval uint32
	VendorId          string
}
type connectResponseIPDR struct {
	Capabilites       uint32
	KeepAliveInterval uint32
}
type getSessionsIPDR struct {
	RequestID uint16
}
type getSessionsResponseIPDR struct {
	RequestID     uint16
	VendorID      string
	SessionBlocks []sessionBlockIPDR
}
type sessionBlockIPDR struct {
	SessionID           uint8
	_                   uint8
	sessionName         string
	sessionDescription  string
	AckTimeInterval     uint32
	AckSequenceInterval uint32
}
type errorIPDR struct {
	timestamp uint32
	errorCode uint16
	description string
}
type sessionStartIPDR struct {
	exporterBootTime uint32
	firstRecordSequenceNumber uint64
	droppedRecordCount uint64
	primary uint8
	ackTimeInterval uint32
	ackSequenceInterval uint32
	documentID [16]byte
}
type stopIPDR struct {
	reasonCode uint16
	reasonInfo string
}

type templateDataIPDR struct {
	ConfigID uint16
	Flags uint8
	TemplateBlock []templateBlockIPDR
}
type templateBlockIPDR struct {
	TemplateID uint16
	SchemaName string
	TypeName string
	fields []fieldDescriptorIPDR
}

type fieldDescriptorIPDR struct {
	TypeID uint32
	FieldID uint32
	FieldName string
	IsEnabled uint8
}

type dataIPDR struct {
	TemplateID uint16
	ConfigID uint16
	Flags uint8
	SequenceNbr uint32
	DataRecord string
}

var (
	pcapFile string = "data/ipdrnetwork.pcap"
	handle   *pcap.Handle
	err      error
)

func parseMessage(packet *bytes.Buffer, messageID uint8, messageLen uint32) {
	switch messageID {
	case 5:
		fmt.Println("CONNECT")

		var connectObj connectIPDR
		err := struc.Unpack(packet, &connectObj)
		if (err != nil) {
			fmt.Println(err)
		}
		fmt.Println(connectObj)
	case 6:
		fmt.Println("CONNECT RESPONSE")
		var connectResponseObj connectResponseIPDR


		fmt.Println(connectResponseObj)
	case 7:
		fmt.Println("Disconnect")
	case 0x23:
		fmt.Println("Errors")

	case 1:
		fmt.Println("Flow Start")
	case 8:
		fmt.Println("Session Start")

	case 3:
		fmt.Println("Flow Stop")


	case 0x09:
		fmt.Println("Session Stop")


	case 0x10:
		fmt.Println("Template Data")


	case 0x1a:
		fmt.Println("Modify Template")//Ignore
	case 0x1b:
		fmt.Println("Modify Template Response")//ignore
	case 0x13:
		fmt.Println("Final Template DATA Ack")//ignore
	case 0x1d:
		fmt.Println("START Negotiation")//Ignore
	case 0x1e:
		fmt.Println("START Negotiation Reject")//Ignore
	case 0x20:
		fmt.Println("DATA")
		var data dataIPDR

		data.DataRecord,err = packet.ReadString(0)
	case 0x21:
		fmt.Println("Data Acknowledge")
	case 0x30:
		fmt.Println("REQUEST")
	case 0x31:
		fmt.Println("RESPONSE")
	case 0x14:
		fmt.Println("GET Sessions")
		var dataObj getSessionsIPDR


		fmt.Println(dataObj)
	case 0x15:

		fmt.Println("Get Sessions Response")


	case 0x16:
		fmt.Println("Get Templates")
	case 0x17:
		fmt.Println("Get Templates Response")
	case 0x40:
		//fmt.Println("Keep Alive")
	}
}
func main() {
	// Open file instead of device
	handle, err = pcap.OpenOffline(pcapFile)
	if err != nil {
		log.Fatal(err)
	}
	defer handle.Close()

	// Loop through packets in file
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

	for packet := range packetSource.Packets() {
		if packet != nil && packet.ApplicationLayer() != nil && packet.ApplicationLayer().Payload() != nil {
			buffer := packet.ApplicationLayer().Payload()
			//buffer := []byte{0x18, 0x2d, 0x44, 0x54, 0xfb, 0x21, 0x09, 0x40}
			partialPacket := bytes.NewBuffer(buffer)
			if partialPacket.Len() >= 8 {
				var headerObj headerIPDR
				err := struc.Unpack(partialPacket, &headerObj)
				if (err != nil) {
					fmt.Println(err)
				}
				fmt.Println(headerObj)
				parseMessage(partialPacket, headerObj.MessageID, headerObj.MessageLen)
				//fmt.Println(header)

			}
		}
	}
}
