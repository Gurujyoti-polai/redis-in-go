package storage

// memory.go

import (
	"strconv"
	"strings"
	"sync"
	"fmt"
	"time"
)

type ValueType int

const (
	TypeString ValueType=iota
	TypeList
	TypeStream
)

type BlockedClient struct {
	Keys []string
	Ch   chan []any
}

type StreamEntry struct {
	ID     string
	Fields map[string]string
}

type Entry struct {
	ValueType ValueType
	stringVal string
	listVal   []string

	streamVal []StreamEntry   // ← NEW

	lastID string 
	expiresAt int64 // unix timestamp in ms, 0 = no expiry
}

type Store struct {
	Mu      sync.RWMutex 
	data map[string]Entry
	Waiting map[string] []*BlockedClient
}

func NewStore() *Store {
	return &Store{
		data: make(map[string]Entry),
		Waiting: make(map[string][]*BlockedClient),
	}
}

func (s *Store) Set(key, value string, expiresAt int64) {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	s.data[key] = Entry{
		stringVal:     value,
		expiresAt: expiresAt,
	}
}

func (s *Store) Get(key string) (any, bool) {
	s.Mu.RLock()
	defer s.Mu.RUnlock()

	entry, ok := s.data[key]
	if !ok {
		return "", false
	}

	return entry.stringVal, true
}

func (s *Store) Delete(key string) {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	delete(s.data, key)
}

func (s *Store) LPUSH(key string, values []string) (int) {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	if clients := s.Waiting[key]; len(clients) > 0 {
		client := clients[0]
		s.Unregister(client)
		client.Ch <- []any{key, values[0]}
		return 1
	}


	entry, exists := s.data[key]
	// if exist and is a list type
	if exists && entry.ValueType == TypeList {
		entry.listVal = append(values, entry.listVal...)
		s.data[key] = entry
		return len(entry.listVal)
	}
	// if key exist and it is a string type
	if exists && entry.ValueType == TypeString {
		// return WRONG error type
		return -1
	}

	//if there is no key make one
	entry = Entry{
		ValueType: TypeList,
		listVal:   values,
		expiresAt: 0,
	}
	s.data[key] = entry
	return len(values)
}

func (s *Store) RPUSH(key string, values []string) (int) {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	if clients := s.Waiting[key]; len(clients) > 0 {
		client := clients[0]
		s.Unregister(client)
		client.Ch <- []any{key, values[0]}
		return 1
	}

	entry, exists := s.data[key]
	// if exist and is a list type
	if exists && entry.ValueType == TypeList {
		entry.listVal = append(entry.listVal, values...)
		s.data[key] = entry
		return len(entry.listVal)
	}
	// if key exist and it is a string type
	if exists && entry.ValueType == TypeString {
		// return WRONG error type
		return -1
	}

	//if there is no key make one
	entry = Entry{
		ValueType: TypeList,
		listVal:   values,
		expiresAt: 0,
	}
	s.data[key] = entry
	return len(values)

}

func (s *Store) LLEN(key string) (int) {
	s.Mu.RLock()
	defer s.Mu.RUnlock()

	entry, exists := s.data[key]

	if !exists {
		return 0
	}
	if entry.ValueType != TypeList{
		return -1
	}
    	
	return len(entry.listVal)
}


// negaive starts from -3, -2, -1
// positive starts from 0, 1, 2
func (s *Store) LRANGE(key string, start int, stop int) ([]string, bool) {
	s.Mu.RLock()
	defer s.Mu.RUnlock()

	entry, exists := s.data[key]

	// Case 1: key does not exist → empty list
	if !exists {
		return []string{}, true
	}

	// Case 2: key exists but wrong type
	if entry.ValueType != TypeList {
		return nil, false // caller converts to WRONGTYPE error
	}

	list := entry.listVal
	length := len(list)

	// Handle negative indexes
	if start < 0 {
		start = length + start
	}
	if stop < 0 {
		stop = length + stop
	}

	// Clamp bounds
	if start < 0 {
		start = 0
	}
	if stop >= length {
		stop = length - 1
	}

	// Empty result cases
	if start > stop || start >= length {
		return []string{}, true
	}

	// Redis LRANGE is inclusive of stop index
	return list[start : stop+1], true
}

func (s *Store) LREM(key string, count int, value string) int {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	entry, exists := s.data[key]
	if !exists {
		return 0
	}

	if entry.ValueType != TypeList {
		return -1 // WRONGTYPE
	}

	list := entry.listVal
	newList := []string{}
	removed := 0

	// Case 1: remove all
	if count == 0 {
		for _, v := range list {
			if v == value {
				removed++
				continue
			}
			newList = append(newList, v)
		}
	}

	// Case 2: remove from left
	if count > 0 {
		for _, v := range list {
			if v == value && removed < count {
				removed++
				continue
			}
			newList = append(newList, v)
		}
	}

	// Case 3: remove from right
	if count < 0 {
		count = -count
		for i := len(list) - 1; i >= 0; i-- {
			if list[i] == value && removed < count {
				removed++
				continue
			}
			newList = append([]string{list[i]}, newList...)
		}
	}

	entry.listVal = newList
	s.data[key] = entry

	return removed
}

func (s *Store) TryLPop(key string) (any, bool, bool) {
	entry, exists := s.data[key]
	if !exists {
		return nil, false, false
	}
	if entry.ValueType != TypeList {
		return nil, false, true // WRONGTYPE
	}
	if len(entry.listVal) == 0 {
		return nil, false, false
	}

	val := entry.listVal[0]
	entry.listVal = entry.listVal[1:]
	s.data[key] = entry
	return val, true, false
}

func (s *Store) TypeOf(key string) string {
	s.Mu.RLock()
	defer s.Mu.RUnlock()

	entry, exists := s.data[key]
	if !exists {
		return "none"
	}

	switch entry.ValueType {
	case TypeString:
		return "string"
	case TypeList:
		return "list"
	case TypeStream:
		return "stream"
	default:
		return "none"
	}
}

func IsNewerID(newMs, newSeq, lastMs, lastSeq int64) bool {
	if newMs > lastMs {
		return true
	}
	if newMs == lastMs && newSeq > lastSeq {
		return true
	}
	return false
}


func (s *Store) AddStreamEntry(key string, id string, fields map[string]string) (string, bool) {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	entry, exists := s.data[key]
	if exists && entry.ValueType != TypeStream {
		return "", false
	}

	// Initialize stream if needed
	if !exists {
		entry = Entry{
			ValueType: TypeStream,
			streamVal: []StreamEntry{},
		}
	}

	// ---------- AUTO ID (*) ----------
	if id == "*" {
		nowMs := time.Now().UnixMilli()
		seq := int64(0)

		if len(entry.streamVal) > 0 {
			last := entry.streamVal[len(entry.streamVal)-1]
			lastMs, lastSeq, _ := ValidateStreamID(last.ID)

			if nowMs == lastMs {
				seq = lastSeq + 1
			} else if nowMs < lastMs {
				nowMs = lastMs
				seq = lastSeq + 1
			}
		}

		id = fmt.Sprintf("%d-%d", nowMs, seq)
	}

	// ---------- EXPLICIT ID ----------
	newMs, newSeq, ok := ValidateStreamID(id)
	if !ok {
		return "", false
	}

	// ---------- ORDER ENFORCEMENT ----------
	if len(entry.streamVal) > 0 {
		last := entry.streamVal[len(entry.streamVal)-1]
		lastMs, lastSeq, _ := ValidateStreamID(last.ID)

		if !IsNewerID(newMs, newSeq, lastMs, lastSeq) {
			return "", false
		}
	}

	// ---------- APPEND ----------
	entry.streamVal = append(entry.streamVal, StreamEntry{
		ID:     id,
		Fields: fields,
	})

	s.data[key] = entry
	return id, true
}


func ValidateStreamID(id string) (int64, int64, bool) {
	parts := strings.Split(id, "-")
	if len(parts) != 2 {
		return 0, 0, false
	}

	ms, err1 := strconv.ParseInt(parts[0], 10, 64)
	seq, err2 := strconv.ParseInt(parts[1], 10, 64)

	if err1 != nil || err2 != nil {
		return 0, 0, false
	}

	if ms < 0 || seq < 0 {
		return 0, 0, false
	}

	return ms, seq, true
}