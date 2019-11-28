// Simple go application that opens a pcap file and print the packets

package ipdrlib

import (
	"bytes"
	"fmt"

	"github.com/lunixbochs/struc"

	"github.com/google/gopacket/pcap"
)

const (
	connectMsgType                  = 0x05
	connectResponseMsgType          = 0x06
	disconnectMsgType               = 0x07
	flowStartMsgType                = 0x01
	flowStopMsgType                 = 0x03
	sessionStartMsgType             = 0x08
	sessionStopMsgType              = 0x09
	keepAliveMsgType                = 0x40
	templateDataMsgType             = 0x10
	modifyTemplateMsgType           = 0x1a
	modifyTemplateResposeMsgType    = 0x1b
	finalTemplateDataAckMsgType     = 0x13
	startNegotiationMsgType         = 0x1d
	startNegotiationRejectMsgType   = 0x1e
	getSessionsMsgType              = 0x14
	getSessionsResponseMsgType      = 0x15
	getTemplatesMsgType             = 0x16
	getTemplatesResponseMsgType     = 0x17
	dataMsgType                     = 0x20
	dataAckMsgType                  = 0x21
	requrestMsgType                 = 0x30
	responseMsgType                 = 0x31
	errorMsgType                    = 0x23
	structuresCapabilities          = 0x01
	multisessionCapabilities        = 0x02
	templateNegotiationCapabilities = 0x04
)

// IPDRStreamingHeaderIdl : The packet header.  you like?
type IPDRStreamingHeaderIdl struct {
	Version      uint8
	MessageID    uint8
	SessionID    uint8
	MessageFlags uint8
	MessageLen   uint32
}

// ConnectIdl : The connected structure.
type ConnectIdl struct {
	CollectorAddress  uint32
	CollectorPort     uint16
	Capabilities      uint32
	KeepAliveInterval uint32
	VendorID          string `struc:"[7]byte"` //This should always be ipdrgo with a null character at the end
}

// ConnectResponseIdl : This command is the response
type ConnectResponseIdl struct {
	Capabilites       uint32
	KeepAliveInterval uint32
	VendorLen         uint32 `struc:"uint32,sizeof=VendorID"`
	VendorID          string
}

// ErrorResponseIdl : This is the error struct
type ErrorResponseIdl struct {
	TimeStamp      uint32
	ErrorCode      uint16
	DescriptionLen uint32 `struc:"uint32,sizeof=Description"`
	Description    string
}
type sessionStartIdl struct {
	ExporterBootTime          int32
	FirstRecordSequenceNumber int64
	DroppedRecordCount        int64
	Primary                   uint8
	AckTimeInterval           int32
	AckSequenceInterval       int32
	DocumentID                [16]byte `struc:"[16]uint8"`
}
type flowStopIdl struct {
	ReasonCode    uint16
	ReasonInfoLen uint8 `struc:"uint32,sizeof=VendorID"`
	ReasonInfo    string
}
type sessionStopIdl struct {
	ReasonCode    uint16
	ReasonInfoLen uint8 `struc:"uint32,sizeof=VendorID"`
	ReasonInfo    string
}
type templateDataIdl struct {
	ConfigID           uint16
	Flags              uint8
	TemplateBlockCount uint32 `struc:"uint32,sizeof=ResultTemplates"`
	ResultTemplates    []templateBlockIdl
}
type templateBlockIdl struct {
	TemplateID   int16
	SchmaNameLen int32 `struc:"uint32,sizeof=SchemaName"`
	SchemaName   string
}
type typeDefinitionIdl struct {
}
type modifyTemplateIdl struct {
	ConfigID        uint16
	Flags           uint8
	ChangeTemplates []templateBlockIdl
}
type modifyTemplateResponseIdl struct {
	ConfigID        uint16
	Flags           uint8
	ResultTemplates []templateBlockIdl
}

//type finalTemplateDataAckIdl {}
//type startNegotiationReject {}
type dataIdl struct {
	TemplateID  uint16
	ConfigID    uint16
	Flags       uint8
	SequenceNum uint32
	//opaque dataRecord<>
}
type dataAckIdl struct {
	ConfigID    uint16
	SequenceNum uint32
}

////SOFAR

type requestIdl struct {
	TemplateID    uint16
	ConfigID      uint16
	Flags         uint8
	RequestNumber uint64
	//opaque dataRecord<>
}
type responseIdl struct {
	TemplateID    uint16
	ConfigID      uint16
	Flags         uint8
	RequestNumber uint32
	//opaque dataRecord<>

}
type getSessionsIdl struct {
	RequestID uint16
}

type getSessionsResponseIdl struct {
	RequestID         uint16
	SessionBlockCount uint32 `struc:"uint32,sizeof=SessionBlocks"`
	SessionBlocks     []sessionBlockIdl
}

type getTemplatesIdl struct {
	RequestID uint16
}

type getTemplateResponseIdl struct {
	RequestID          uint16
	ConfigID           uint16
	TemplateBlockCount uint32 `struc:"uint32, sizeof=CurrentTemplates"`
	CurrentTemplates   []templateBlockIdl
}

//type keepAliveIdl struct {}

type getTemplateResponseWithUDTsIdl struct {
	RequestID            uint16
	ConfigID             uint16
	UserDefinedTypeCount uint32 `struc:"uint32, sizeof=UserDefinedTypes"`
	UserDefinedTypes     []typeDefinitionIdl
	TemplateBlockCount   uint32 `struc:"uint32, sizeof=CurrentTemplates"`
	CurrentTemplates     []templateBlockIdl
}

type fieldDescriptorIdl struct {
	TypeID    uint32
	FieldID   uint32
	FieldName string
	IsEnabled uint8
}
type sessionBlockIdl struct {
	SessionID             uint8
	Reserved              uint8
	SessionNameLen        uint32 `struc:"uint32,sizeof=SessionName"`
	SessionName           string
	SessionDescriptionLen uint32 `struc:"uint32,sizeof=SessionDescription"`
	SessionDescription    string
	AckTimeInterval       uint32
	AckSequenceInterval   uint32
}
type versionRequestIdl struct {
	requesterAddress  uint32
	requesterBootTime uint32
	msg               []byte //4 characters
}
type protocolInfoIdl struct {
	transportType   uint32
	protocolVersion uint32
	portNumber      uint16
	reserved        uint16
}

var (
	pcapFile string = "data/ipdrnetwork.pcap"
	handle   *pcap.Handle
	err      error
)

//Connect : Just Connect
func Connect(hdr IPDRStreamingHeaderIdl, details ConnectIdl) bytes.Buffer {
	var bufhdr bytes.Buffer
	var bufdtls bytes.Buffer
	var result bytes.Buffer
	struc.Pack(&bufhdr, &hdr)
	struc.Pack(&bufdtls, &details)
	fmt.Println(bufdtls)
	result.Write(bufhdr.Bytes())
	result.Write(bufdtls.Bytes())
	return result

}

//ParseMessageByType : Parses message by message type
func ParseMessageByType(packet *bytes.Buffer, messageID uint8, messageLen uint32) {
	switch messageID {
	case 5:
		fmt.Println("CONNECT")

		var connectObj ConnectIdl
		err := struc.Unpack(packet, &connectObj)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(connectObj)
		}
	case 6:
		fmt.Println("CONNECT RESPONSE")
		var connectResponseObj ConnectResponseIdl
		err := struc.Unpack(packet, &connectResponseObj)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(connectResponseObj)
		}
	case 7:
		fmt.Println("Disconnect")
	case 0x23:
		fmt.Println("Errors")
		var errorObj ErrorResponseIdl
		struc.Unpack(packet, &errorObj)
		fmt.Println(errorObj)

	case 1:
		fmt.Println("Flow Start")
	case 8:
		fmt.Println("Session Start")
		var sessionStartObj sessionStartIdl
		err := struc.Unpack(packet, &sessionStartObj)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("We got data")
			fmt.Println(sessionStartObj)
		}
	case 3:
		fmt.Println("Flow Stop")

	case 0x09:
		fmt.Println("Session Stop")

	case 0x10:
		fmt.Println("Template Data")
		var templateDataObj templateDataIdl
		err := struc.Unpack(packet, &templateDataObj)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(templateDataObj)
		}
	case 0x1a:
		fmt.Println("Modify Template") //Ignore
	case 0x1b:
		fmt.Println("Modify Template Response") //ignore
	case 0x13:
		fmt.Println("Final Template DATA Ack") //ignore
	case 0x1d:
		fmt.Println("START Negotiation") //Ignore
	case 0x1e:
		fmt.Println("START Negotiation Reject") //Ignore
	case 0x20:
		fmt.Println("DATA")
		//var data dataIdl

		//data.DataRecord, err = packet.ReadString(0)
	case 0x21:
		fmt.Println("Data Acknowledge")
	case 0x30:
		fmt.Println("REQUEST")
	case 0x31:
		fmt.Println("RESPONSE")
	case 0x14:
		fmt.Println("GET Sessions")
		var dataObj getSessionsIdl

		err := struc.Unpack(packet, &dataObj)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(dataObj)
		}
		fmt.Println(dataObj)
	case 0x15:

		fmt.Println("Get Sessions Response")
		var dataObj getSessionsResponseIdl

		err := struc.Unpack(packet, &dataObj)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(dataObj)
		}
		fmt.Println(dataObj)

	case 0x16:
		fmt.Println("Get Templates")
	case 0x17:
		fmt.Println("Get Templates Response")
	case 0x40:
		//fmt.Println("Keep Alive")
	}
}

// func main() {
// 	// Open file instead of device
// 	handle, err = pcap.OpenOffline(pcapFile)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer handle.Close()

// 	// Loop through packets in file
// 	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

// 	for packet := range packetSource.Packets() {
// 		if packet != nil && packet.ApplicationLayer() != nil && packet.ApplicationLayer().Payload() != nil {
// 			buffer := packet.ApplicationLayer().Payload()
// 			//buffer := []byte{0x18, 0x2d, 0x44, 0x54, 0xfb, 0x21, 0x09, 0x40}
// 			partialPacket := bytes.NewBuffer(buffer)
// 			if partialPacket.Len() >= 8 {
// 				var headerObj iPDRStreamingHeaderIdl
// 				err := struc.Unpack(partialPacket, &headerObj)
// 				if err != nil {
// 					fmt.Println(err)
// 				}
// 				fmt.Println(headerObj)
// 				parseMessage(partialPacket, headerObj.MessageID, headerObj.MessageLen)
// 				//fmt.Println(header)

// 			}
// 		}
// 	}
// }