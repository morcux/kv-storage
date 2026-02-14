package entry

import (
	"encoding/binary"
	"errors"
	"hash/crc32"
	"time"
)

type Entry struct {
	Key   string
	Value []byte
}

func (e *Entry) Encode() []byte {
	crc32Value := crc32.ChecksumIEEE([]byte(e.Key))
	timestamp := uint64(time.Now().Unix())
	keySize := uint32(len(e.Key))
	valueSize := uint32(len(e.Value))

	buffer := make([]byte, 4+8+4+4+keySize+valueSize)
	binary.BigEndian.PutUint32(buffer, crc32Value)
	binary.BigEndian.PutUint64(buffer[4:], timestamp)
	binary.BigEndian.PutUint32(buffer[12:], keySize)
	binary.BigEndian.PutUint32(buffer[16:], valueSize)
	copy(buffer[20:], []byte(e.Key))
	copy(buffer[20+keySize:], []byte(e.Value))

	return buffer
}

func (e *Entry) Decode(data []byte) error {
	if len(data) < 20 {
		return errors.New("invalid entry data")
	}

	keySize := binary.BigEndian.Uint32(data[12:])
	valueSize := binary.BigEndian.Uint32(data[16:])

	if len(data) < 20+int(keySize)+int(valueSize) {
		return errors.New("invalid entry data")
	}

	e.Key = string(data[20 : 20+keySize])
	e.Value = data[20+keySize : 20+keySize+valueSize]

	return nil
}
