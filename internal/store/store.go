package store

import (
	"encoding/binary"
	"errors"
	"io"
	"kv-storage/internal/entry"
	"os"
	"sync"
)

const HeaderSize = 20

type Store struct {
	file  *os.File
	index map[string]int64
	mu    sync.RWMutex
}

func NewStore(path string) (*Store, error) {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}

	index := make(map[string]int64)
	var currentOffset int64 = 0

	for {
		header := make([]byte, HeaderSize)
		_, err := file.ReadAt(header, currentOffset)
		if err == io.EOF {
			break
		}
		if err != nil {
			if len(header) == 0 {
				break
			}
			break
		}

		keySize := binary.BigEndian.Uint32(header[12:16])
		valSize := binary.BigEndian.Uint32(header[16:20])

		totalSize := int64(HeaderSize + keySize + valSize)

		recordData := make([]byte, totalSize)
		_, err = file.ReadAt(recordData, currentOffset)
		if err != nil {
			break
		}

		e := entry.Entry{}
		err = e.Decode(recordData)
		if err != nil {
			break
		}

		index[e.Key] = currentOffset

		currentOffset += totalSize
	}

	return &Store{
		file:  file,
		index: index,
	}, nil
}

func (s *Store) Set(key string, value []byte) error {
	e := entry.Entry{
		Key:   key,
		Value: value,
	}
	encoded := e.Encode()

	s.mu.Lock()
	defer s.mu.Unlock()

	stat, err := s.file.Stat()
	if err != nil {
		return err
	}
	offset := stat.Size()

	_, err = s.file.Write(encoded)
	if err != nil {
		return err
	}

	s.index[key] = offset

	return nil
}

func (s *Store) Get(key string) (entry.Entry, error) {
	s.mu.RLock()
	offset, ok := s.index[key]
	s.mu.RUnlock()

	if !ok {
		return entry.Entry{}, errors.New("key not found")
	}

	header := make([]byte, HeaderSize)
	_, err := s.file.ReadAt(header, offset)
	if err != nil {
		return entry.Entry{}, err
	}

	keySize := binary.BigEndian.Uint32(header[12:16])
	valSize := binary.BigEndian.Uint32(header[16:20])
	totalSize := HeaderSize + keySize + valSize

	data := make([]byte, totalSize)
	_, err = s.file.ReadAt(data, offset)
	if err != nil {
		return entry.Entry{}, err
	}

	e := entry.Entry{}
	err = e.Decode(data)
	if err != nil {
		return entry.Entry{}, err
	}
	return e, nil
}

func (s *Store) GetAll() (map[string][]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	all := make(map[string][]byte)
	for _, offset := range s.index {
		header := make([]byte, HeaderSize)
		_, err := s.file.ReadAt(header, offset)
		if err != nil {
			return nil, err
		}

		keySize := binary.BigEndian.Uint32(header[12:16])
		valSize := binary.BigEndian.Uint32(header[16:20])
		totalSize := HeaderSize + keySize + valSize

		data := make([]byte, totalSize)
		_, err = s.file.ReadAt(data, offset)
		if err != nil {
			return nil, err
		}

		e := entry.Entry{}
		err = e.Decode(data)
		if err != nil {
			return nil, err
		}
		all[e.Key] = e.Value
	}
	return all, nil
}

func (s *Store) RestoreFromMap(data map[string][]byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.file.Close(); err != nil {
		return err
	}
	s.file.Truncate(0)
	s.file.Seek(0, 0)
	for _, v := range data {
		s.file.Write(v)
	}
	s.index = make(map[string]int64)
	var offset int64 = 0
	for k, v := range data {
		s.index[k] = offset
		offset += int64(len(v))
	}
	return nil
}

func (s *Store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.file.Close()
}
