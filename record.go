package dtpatch

import (
	"encoding/binary"
	"errors"
	"fmt"
)

var fixedTag = [28]byte{
	0xEE, 0xA8, 0x50, 0xA3, 0xFF, 0x4D, 0xE3, 0xD0, 0xAA, 0x23,
	0x4C, 0xFA, 0xFF, 0x07, 0x8A, 0x87, 0xC6, 0x9A, 0xB0, 0xBB,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
}

var trailingPadding = [21]byte{}

var (
	ErrRecordSize = errors.New("origRecord is not OldRecordSize bytes")
	ErrTruncated  = errors.New("length prefix overruns buffer")
)

func buildPatchRecord(origRecord []byte, patchNum int) ([]byte, error) {
	if len(origRecord) != OldRecordSize {
		return nil, ErrRecordSize
	}
	counter := binary.LittleEndian.Uint32(origRecord[8:12])
	count := binary.LittleEndian.Uint32(origRecord[12:16])
	str1, _, err := readString(origRecord, 16)
	if err != nil {
		return nil, err
	}

	out := append([]byte{}, origRecord[0:8]...)
	out = binary.LittleEndian.AppendUint32(out, counter+1)
	out = binary.LittleEndian.AppendUint32(out, count)
	out = appendString(out, str1)
	out = appendString(out, str1+".stream")
	out = append(out, 0x00)
	out = append(out, fixedTag[:]...)
	out = binary.LittleEndian.AppendUint32(out, count)
	out = appendString(out, str1+fmt.Sprintf(".patch_%d", patchNum))
	out = appendString(out, str1+fmt.Sprintf(".stream.patch_%d", patchNum))
	out = append(out, trailingPadding[:]...)
	return out, nil
}

func readString(b []byte, offset int) (string, int, error) {
	if offset+4 > len(b) {
		return "", 0, ErrTruncated
	}
	length := binary.LittleEndian.Uint32(b[offset : offset+4])
	start := offset + 4
	end := start + int(length)
	if end > len(b) {
		return "", 0, ErrTruncated
	}
	return string(b[start:end]), end, nil
}

func appendString(out []byte, s string) []byte {
	out = binary.LittleEndian.AppendUint32(out, uint32(len(s)))
	return append(out, s...)
}
