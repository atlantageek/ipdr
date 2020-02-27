package xdrlib

import (
	"bytes"
	"ipdrlib"
	"fmt"
	"encoding/binary"
	"time"
	"encoding/json"
	"encoding/hex"
	"strconv"
)

const (
	intType = 0x21
	unsignedIntType = 0x22
	longType = 0x23
	unsignedLongType = 0x24
	floatType = 0x25
	doubleType = 0x26
	hexBinaryType = 0x27
	stringType = 0x28
	booleanType = 0x29
	byteType = 0x2A
	unsignedByteType = 0x2b
	shortType = 0x2c
	unsignedShortType = 0x2d
	dateTimeType = 0x122
	dateTimeMsecType = 0x224
	ipv4AddrType = 0x322
	ipv6AddrType = 0x427
	ipAddrType = 0x827
	uuidType = 0x527
	dateTimeUsecType = 0x623
	macAddrType = 0x723

)

func ParseData( fields []ipdrlib.FieldDescriptorIdl, data []byte )(map[string]string,string) {
	var reclen uint32
	buf := bytes.NewReader(data)
	result := make(map[string]string)
	err := binary.Read(buf, binary.BigEndian, &reclen)
	if err != nil {
		fmt.Println("Error:", err)
		return result,""
	}

	
	for _, field:= range fields {

		switch typeID := field.TypeID; typeID {
		case intType: 
			var val int32
			binary.Read(buf, binary.BigEndian, &val)

			result[field.FieldName] = strconv.Itoa(int(val))

		case unsignedIntType:
			var val int32
			binary.Read(buf, binary.BigEndian, &val)
			
			result[field.FieldName] = strconv.Itoa(int(val))
		case longType:
			var val int64
			binary.Read(buf, binary.BigEndian, &val)
			
			result[field.FieldName] = strconv.Itoa(int(val))
		case unsignedLongType:
			var val int64
			binary.Read(buf, binary.BigEndian, &val)
			
			result[field.FieldName] = strconv.Itoa(int(val))
		case floatType:
			var val float32
			binary.Read(buf, binary.BigEndian, &val)
			
			result[field.FieldName] = strconv.Itoa(int(val))
		case doubleType:
			var val float64
			binary.Read(buf, binary.BigEndian, &val)
			
			result[field.FieldName] = strconv.Itoa(int(val))
		case hexBinaryType:
			var hexLen int32
			 binary.Read(buf, binary.BigEndian, &hexLen)
			val := make([]byte, hexLen )
			binary.Read(buf, binary.BigEndian, val)
			
			result[field.FieldName] = hex.EncodeToString(val)

		case stringType:
			var strLen int32
			binary.Read(buf, binary.BigEndian, &strLen)
			val := make([]byte, strLen)
			binary.Read(buf, binary.BigEndian, val)
	
			result[field.FieldName] = string(val)

		case booleanType:
			var boolVal byte
			var val bool = false
			binary.Read(buf, binary.BigEndian, &boolVal)
			if boolVal == 1 {
				val = true
			}
			result[field.FieldName] = strconv.FormatBool(val)
		case byteType:
			var val byte
			 binary.Read(buf, binary.BigEndian, &val)
			 
			 result[field.FieldName] = strconv.Itoa(int(val))
		case unsignedByteType:
			var val uint8
			 binary.Read(buf, binary.BigEndian, &val)
			 
			 result[field.FieldName] = strconv.Itoa(int(val))
		case shortType:
			var val int16
			 binary.Read(buf, binary.BigEndian, &val)
			 
			 result[field.FieldName] = strconv.Itoa(int(val))
		case unsignedShortType:
			var val uint16
			 binary.Read(buf, binary.BigEndian, &val)
			 
			 result[field.FieldName] = strconv.Itoa(int(val))
		case dateTimeType:
			var val uint32
			binary.Read(buf, binary.BigEndian, &val)
			
			result[field.FieldName] = time.Unix(int64(val),0).String()
		case dateTimeMsecType:
			var val uint64
			binary.Read(buf, binary.BigEndian, &val)
			result[field.FieldName] = time.Unix(0,int64(val) * 1000000).String()
			
		case ipv4AddrType:
			var val [4]byte
			binary.Read(buf, binary.BigEndian, &val)

			result[field.FieldName] = fmt.Sprintf("%d.%d.%d.%d", val[0],val[1],val[2],val[3])
		case ipv6AddrType:
			val := make([]byte, 20)
			binary.Read(buf, binary.BigEndian, val)
			
			result[field.FieldName] = hex.EncodeToString(val[4:20])
		case ipAddrType:
			var valLen int32
			binary.Read(buf, binary.BigEndian, &valLen)
		   val := make([]byte, valLen )
		   binary.Read(buf, binary.BigEndian, val)
		   
		   result[field.FieldName] = hex.EncodeToString(val)
		case uuidType:
			var valLen int32
			binary.Read(buf, binary.BigEndian, &valLen)
		   val := make([]byte, valLen )
		   binary.Read(buf, binary.BigEndian, val)
		   
		   result[field.FieldName] = hex.EncodeToString(val)
		case dateTimeUsecType:
			var val uint64
			binary.Read(buf, binary.BigEndian, &val)
			
			result[field.FieldName] = strconv.Itoa(int(val))
		case macAddrType:
			val := make([]byte, 8 )
			binary.Read(buf, binary.BigEndian, val)
			
			result[field.FieldName] = hex.EncodeToString(val)
		}
		//fmt.Println(err)
	}
	jsondata,err := json.Marshal(result)
	jsonStr :=""
	if (err != nil) {
		fmt.Println(err)
		fmt.Println("ERROR")
		return result,""
	} else {

		jsonStr = string(jsondata[:])

	}
	return result,jsonStr
}