package common

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"unsafe"
)

type SliceMock struct {
	addr uintptr
	len  int
	cap  int
}

func Struct2bytes(mattr *Mattr) []byte {
	Len := unsafe.Sizeof(*mattr)
	bytes := &SliceMock{addr: uintptr(unsafe.Pointer(mattr)), cap: int(Len), len: int(Len)}
	data := *(*[]byte)(unsafe.Pointer(bytes))
	return data
}

func Bytes2Struct(data []byte) {
	var preStruct *Mattr = *(**Mattr)(unsafe.Pointer(&data))
	fmt.Println(preStruct)
}

func String2bytes(str string) []byte {
	return []byte(str)
}

func byte2String(p []byte) string {
	for i := 0; i < len(p); i++ {
		if p[i] == 0 {
			return string(p[0:i])
		}
	}
	return string(p)
}

func Int64ToByte(num int64) (error, []byte) {
	var buffer bytes.Buffer
	err := binary.Write(&buffer, binary.BigEndian, num) //注意大小端
	if err != nil {
		return err, nil
	}
	return err, buffer.Bytes()
}

func Int32ToByte(num int32) (error, []byte) {
	var buffer bytes.Buffer
	err := binary.Write(&buffer, binary.LittleEndian, num) //注意大小端
	if err != nil {
		return err, nil
	}
	return err, buffer.Bytes()
}

//BytesCombine 多个[]byte数组合并成一个[]byte
func BytesCombine(pBytes ...[]byte) []byte {
	return bytes.Join(pBytes, []byte(""))
}
