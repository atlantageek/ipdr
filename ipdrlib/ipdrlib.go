// Simple go application that opens a pcap file and print the packets

package ipdrlib

import (
	"bytes"
	"fmt"

	"github.com/lunixbochs/struc"

	"github.com/google/gopacket/pcap"
)

const (
	// ConnectMsgType : Connection message Type ID
	ConnectMsgType = 0x05
	// ConnectResponseMsgType : Connection Response message Type ID
	ConnectResponseMsgType = 0x06
	// DisconnectMsgType : disconnect message Type ID
	DisconnectMsgType = 0x07
	// FlowStartMsgType : Connection message Type ID
	FlowStartMsgType = 0x01
	// FlowStopMsgType : Connection message Type ID
	FlowStopMsgType = 0x03
	// SessionStartMsgType : Connection message Type ID
	SessionStartMsgType = 0x08
	// SessionStopMsgType : Connection message Type ID
	SessionStopMsgType = 0x09
	// KeepAliveMsgType : Connection message Type ID
	KeepAliveMsgType = 0x40
	// TemplateDataMsgType : Connection message Type ID
	TemplateDataMsgType = 0x10
	// ModifyTemplateMsgType : Connection message Type ID
	ModifyTemplateMsgType = 0x1a
	// ModifyTemplateResponseMsgType : Connection message Type ID
	ModifyTemplateResponseMsgType = 0x1b
	// FinalTemplateDataAckMsgType : Connection message Type ID
	FinalTemplateDataAckMsgType = 0x13
	// StartNegotiationMsgType : Connection message Type ID
	StartNegotiationMsgType = 0x1d
	// ConnectMsgType : Connection message Type ID
	startNegotiationRejectMsgType = 0x1e
	// GetSessionsMsgType : Connection message Type ID
	GetSessionsMsgType = 0x14
	// GetSessionsResponseMsgType : Connection message Type ID
	GetSessionsResponseMsgType = 0x15
	// GetTemplatesMsgType : Connection message Type ID
	GetTemplatesMsgType = 0x16
	// GetTemplatesResponseMsgType : Connection message Type ID
	GetTemplatesResponseMsgType = 0x17
	// DataMsgType : Connection message Type ID
	DataMsgType = 0x20
	// DataAckMsgType : Connection message Type ID
	dataAckMsgType = 0x21
	// RequestMsgType : Connection message Type ID
	RequestMsgType = 0x30
	// ResponseMsgType : Connection message Type ID
	ResponseMsgType = 0x31
	// ErrorMsgType : Connection message Type ID
	ErrorMsgType = 0x23

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

// TemplateDataIdl: template data IDL
type TemplateDataIdl struct {
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

// DataIdl : This is the datastruct
type DataIdl struct {
	TemplateID  uint16
	ConfigID    uint16
	Flags       uint8
	SequenceNum uint64

	//opaque dataRecord<>
}

// DataAckIdl : This is the Data Acknowledgement
type DataAckIdl struct {
	ConfigID    uint16
	SequenceNum uint64
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

// GetSessionsIdl : This command is the GetSessions
type GetSessionsIdl struct {
	RequestID uint16
}

type GetSessionsResponseIdl struct {
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
	SessionType           uint8
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
	result.Write(bufhdr.Bytes())
	result.Write(bufdtls.Bytes())
	return result

}

//GetSessions : Just GetSessions
func GetSessions(hdr IPDRStreamingHeaderIdl, details GetSessionsIdl) bytes.Buffer {
	var bufhdr bytes.Buffer
	var bufdtls bytes.Buffer
	var result bytes.Buffer

	struc.Pack(&bufhdr, &hdr)
	struc.Pack(&bufdtls, &details)

	result.Write(bufhdr.Bytes())
	result.Write(bufdtls.Bytes())
	return result
}

//DataAck : Just GetSessions
func DataAck(hdr IPDRStreamingHeaderIdl, details DataAckIdl) bytes.Buffer {
	var bufhdr bytes.Buffer
	var bufdtls bytes.Buffer
	var result bytes.Buffer

	struc.Pack(&bufhdr, &hdr)
	struc.Pack(&bufdtls, &details)
	fmt.Println("------")
	fmt.Println(details)

	result.Write(bufhdr.Bytes())
	result.Write(bufdtls.Bytes())
	return result
}

//Hdr : Just Hdr
func Hdr(hdr IPDRStreamingHeaderIdl) bytes.Buffer {
	var bufhdr bytes.Buffer
	struc.Pack(&bufhdr, &hdr)
	return bufhdr

}

//ParseMessageByType : Parses message by message type
func ParseMessageByType(packet *bytes.Buffer, messageID uint8, messageLen uint32) interface{} {
	switch messageID {
	case ConnectMsgType:
		fmt.Println("CONNECT")

		var connectObj ConnectIdl
		err := struc.Unpack(packet, &connectObj)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(connectObj)
		}
	case ConnectResponseMsgType:
		fmt.Println("CONNECT RESPONSE")
		var connectResponseObj ConnectResponseIdl
		err := struc.Unpack(packet, &connectResponseObj)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(connectResponseObj)
		}
		return connectResponseObj
	case DisconnectMsgType:
		fmt.Println("Disconnect")
	case ErrorMsgType:
		fmt.Println("Errors")
		var errorObj ErrorResponseIdl
		struc.Unpack(packet, &errorObj)
		fmt.Println(errorObj)

	case FlowStartMsgType:
		fmt.Println("Flow Start")
	case SessionStartMsgType:
		fmt.Println("Session Start")
		var sessionStartObj sessionStartIdl
		err := struc.Unpack(packet, &sessionStartObj)
		if err != nil {
			fmt.Println(err)
		} else {

			fmt.Println(sessionStartObj)
		}
	case FlowStopMsgType:
		fmt.Println("Flow Stop")

	case SessionStopMsgType:
		fmt.Println("Session Stop")

	case TemplateDataMsgType:
		fmt.Println("Template Data")
		var templateDataObj TemplateDataIdl
		err := struc.Unpack(packet, &templateDataObj)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(templateDataObj)
		}
		return templateDataObj
	case ModifyTemplateMsgType:
		fmt.Println("Modify Template") //Ignore
	case ModifyTemplateResponseMsgType:
		fmt.Println("Modify Template Response") //ignore
	case FinalTemplateDataAckMsgType:
		fmt.Println("Final Template DATA Ack") //ignore
	case StartNegotiationMsgType:
		fmt.Println("START Negotiation") //Ignore
	case startNegotiationRejectMsgType:
		fmt.Println("START Negotiation Reject") //Ignore
	case DataMsgType:
		fmt.Println("DATA")
		var dataObj DataIdl
		err := struc.Unpack(packet, &dataObj)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(dataObj)
		}
		return dataObj
		//var data dataIdl

		//data.DataRecord, err = packet.ReadString(0)
	case dataAckMsgType:
		fmt.Println("Data Acknowledge")
	case RequestMsgType:
		fmt.Println("REQUEST")
	case ResponseMsgType:
		fmt.Println("RESPONSE")
	case GetSessionsMsgType:
		fmt.Println("GET Sessions")
		var dataObj GetSessionsIdl

		err := struc.Unpack(packet, &dataObj)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(dataObj)
		}
		fmt.Println(dataObj)
	case GetSessionsResponseMsgType:

		fmt.Println("Get Sessions Response")
		var dataObj GetSessionsResponseIdl

		err := struc.Unpack(packet, &dataObj)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(dataObj)
		}
		fmt.Println(dataObj)
		return dataObj

	case GetTemplatesMsgType:
		fmt.Println("Get Templates")
	case GetTemplatesResponseMsgType:
		fmt.Println("Get Templates Response")
	case KeepAliveMsgType:
		//fmt.Println("Keep Alive")
	}
	return nil
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
