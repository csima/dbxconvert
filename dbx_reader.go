package main

import (
	"encoding/binary"
	"errors"
	"os"
	"strings"
	"time"
)

const (
	DBX_TYPE_EMAIL   = 0
	DBX_TYPE_FOLDER  = 1
	DBX_TYPE_OE4     = 2
	DBX_TYPE_UNKNOWN = 3
)

//DBX Reader type
//Provides access to all DBX related functions
type DBXReader struct {
	f                 *os.File
	fSize             int64
	internalFile      bool
	dbxType           int
	indexes           []uint32
	subjects          []string
	senders           []string
	senderAddresses   []string
	receivers         []string
	receiverAddresses []string
	receiveDates      []time.Time
	sendDates         []time.Time
}

func (self *DBXReader) init() error {
	self.receiveDates = []time.Time{}
	self.sendDates = []time.Time{}

	signature := [4]uint32{}
	for i := 0; i < 4; i++ {
		_ = binary.Read(self.f, binary.LittleEndian, &signature[i])
	}

	if signature[0] == 0xFE12ADCF && signature[1] == 0x6F74FDC5 && signature[2] == 0x11D1E366 && signature[3] == 0xC0004E9A {
		self.dbxType = DBX_TYPE_EMAIL
	} else if signature[0] == 0x36464D4A && signature[1] == 0x00010003 {
		self.dbxType = DBX_TYPE_OE4
	} else if signature[0] == 0xFE12ADCF && signature[1] == 0x6F74FDC6 && signature[2] == 0x11D1E366 && signature[3] == 0xC0004E9A {
		self.dbxType = DBX_TYPE_FOLDER
	} else {
		self.dbxType = DBX_TYPE_UNKNOWN
	}

	if self.dbxType == DBX_TYPE_EMAIL {
		self.readIndexes()
		self.readInfos()
		return nil
	}

	return errors.New("File is not in Outlook Express DBX format")
}

func (self *DBXReader) readIndexes() {
	var INDEX_POINTER int64
	var ITEM_COUNT int64

	INDEX_POINTER = 0xE4
	ITEM_COUNT = 0xC4

	var indexPtr uint32
	var itemCount uint32

	_, _ = self.f.Seek(INDEX_POINTER, 0)
	_ = binary.Read(self.f, binary.LittleEndian, &indexPtr)

	_, _ = self.f.Seek(ITEM_COUNT, 0)
	_ = binary.Read(self.f, binary.LittleEndian, &itemCount)

	if itemCount > 0 {
		self.readIndex(int(indexPtr))
	}
}

func (self *DBXReader) readIndex(pos int) {
	if int64(pos) >= self.fSize {
		panic("Bad seek")
	}

	var nextTable uint32
	var ptrCount uint8
	var indexCount uint32

	_, _ = self.f.Seek(int64(pos)+8, 0)
	_ = binary.Read(self.f, binary.LittleEndian, &nextTable)
	_, _ = self.f.Seek(5, 1)
	_ = binary.Read(self.f, binary.LittleEndian, &ptrCount)
	_, _ = self.f.Seek(2, 1)
	_ = binary.Read(self.f, binary.LittleEndian, &indexCount)

	if indexCount > 0 {
		self.readIndex(int(nextTable))
	}

	pos += 24

	var i int64
	for i = 0; i < int64(ptrCount); i++ {
		_, _ = self.f.Seek(int64(pos), 0)
		var indexPtr uint32
		_ = binary.Read(self.f, binary.LittleEndian, &indexPtr)
		_ = binary.Read(self.f, binary.LittleEndian, &nextTable)
		_ = binary.Read(self.f, binary.LittleEndian, &indexCount)

		self.indexes = append(self.indexes, indexPtr)
		pos += 12

		if indexCount > 0 {
			self.readIndex(int(nextTable))
		}
	}
}

func (self *DBXReader) readInfos() {
	for i := 0; i < self.GetItemCount(); i++ {
		var index uint32
		index = uint32(self.GetIndex(i))
		_, _ = self.f.Seek(int64(index)+4, 0)
		var size uint32
		_ = binary.Read(self.f, binary.LittleEndian, &size)
		_, _ = self.f.Seek(2, 1)
		var count uint8
		_ = binary.Read(self.f, binary.LittleEndian, &count)
		_, _ = self.f.Seek(1, 1)

		var offset uint32
		offset = 0

		var sender bool
		var senderAddress bool
		var receiver bool
		var receiverAddress bool
		var subject bool
		var receiverDate bool
		var sendDate bool

		var pos uint32
		pos = index + 12

		var j uint8
		for j = 0; j < count; j++ {
			_, _ = self.f.Seek(int64(pos), 0)
			var tp uint8
			_ = binary.Read(self.f, binary.LittleEndian, &tp)
			var value uint32

			b := []byte{}
			var bt byte
			for n := 0; n < 3; n++ {
				_ = binary.Read(self.f, binary.LittleEndian, &bt)
				b = append(b, bt)
			}
			b = append(b, 0)
			value = binary.LittleEndian.Uint32(b)

			offset = uint32(int(index) + 12 + 4*int(count) + int(value))
			switch tp {
			case 0x02:
				self.sendDates = append(self.sendDates, self.readDate(int(offset)))
				sendDate = true
			case 0x0E:
				self.senderAddresses = append(self.senderAddresses, self.readString(int(offset)))
				senderAddress = true
			case 0x0D:
				self.senders = append(self.senders, self.readString(int(offset)))
				sender = true
			case 0x08:
				self.subjects = append(self.subjects, self.readString(int(offset)))
				subject = true
			case 0x12:
				self.receiveDates = append(self.receiveDates, self.readDate(int(offset)))
				receiverDate = true
			case 0x13:
				self.receivers = append(self.receivers, self.readString(int(offset)))
				receiver = true
			case 0x14:
				self.receiverAddresses = append(self.receiverAddresses, self.readString(int(offset)))
				receiverAddress = true
			}
			pos += 4
		}

		if !sender {
			self.senders = append(self.senders, "")
		}
		if !senderAddress {
			self.senderAddresses = append(self.senderAddresses, "")
		}
		if !receiver {
			self.receivers = append(self.receivers, "")
		}
		if !receiverAddress {
			self.receiverAddresses = append(self.receiverAddresses, "")
		}
		if !subject {
			self.subjects = append(self.subjects, "")
		}
		if !receiverDate {
			self.receiveDates = append(self.receiveDates, time.Time{})
		}
		if !sendDate {
			self.sendDates = append(self.sendDates, time.Time{})
		}

	}
}

func (self *DBXReader) readString(offset int) (s string) {
	if int64(offset) >= self.fSize {
		panic("Bad seek")
	}
	_, _ = self.f.Seek(int64(offset), 0)
	var c []byte
	var ch byte

	c = []byte{}
	for {
		_ = binary.Read(self.f, binary.LittleEndian, &ch)
		if ch != 0x00 {
			c = append(c, ch)
			continue
		} else {
			c = append(c, 0x00)
			s += string(c)
			if len(c) != 256 {
				break
			} else {
				c = []byte{}
			}
		}
	}
	return
}

func (self *DBXReader) readDate(offset int) time.Time {
	if int64(offset) >= self.fSize {
		panic("Bad seek")
	}
	_, _ = self.f.Seek(int64(offset), 0)
	var fileTime int64
	_ = binary.Read(self.f, binary.LittleEndian, &fileTime)
	if fileTime < 0 {
		return time.Time{}
	}
	var TICKS_PER_SECOND int64 = 10000000
	var EPOCH_DIFFERENCE int64 = 11644473600
	temp := fileTime / TICKS_PER_SECOND
	temp = temp - EPOCH_DIFFERENCE
	return time.Unix(temp, 0)
}

// Opens DBX file by file name or returns error
func (self *DBXReader) Open(fn string) error {
	var err error

	self.internalFile = false

	self.f, err = os.Open(fn)
	if err != nil {
		return err
	}

	stat, _ := self.f.Stat()
	self.fSize = stat.Size()

	return self.init()
}

// Closes DBX file
func (self *DBXReader) Close() error {
	self.fSize = 0
	self.internalFile = false
	self.dbxType = -1
	self.indexes = []uint32{}
	self.subjects = []string{}
	self.senders = []string{}
	self.senderAddresses = []string{}
	self.receivers = []string{}
	self.receiverAddresses = []string{}
	self.receiveDates = []time.Time{}
	self.sendDates = []time.Time{}
	return self.f.Close()
}

// Returns full loaded DBX file name
func (self *DBXReader) GetFileName() string {
	return self.f.Name()
}

// Returns loaded file name
func (self *DBXReader) GetFName() string {
	fi, err := self.f.Stat()
	if err != nil {
		return ""
	}
	return fi.Name()
}

// Returns loaded DBX file creation date
func (self *DBXReader) GetFileDate() time.Time {
	fi, _ := self.f.Stat()
	return fi.ModTime()
}

// Returns loaded DBX file type
func (self *DBXReader) GetType() int {
	return self.dbxType
}

// Returns number of messages
func (self *DBXReader) GetItemCount() int {
	return len(self.indexes)
}

// Returns message offset
func (self *DBXReader) GetIndex(i int) int {
	return int(self.indexes[uint32(i)])
}

// Returns message sender name
func (self *DBXReader) GetSender(msgNumber int) string {
	return strings.Trim(self.senders[msgNumber], "\x00")
}

// Returns message sender address
func (self *DBXReader) GetSenderAddress(msgNumber int) string {
	return strings.Trim(self.senderAddresses[msgNumber], "\x00")
}

// Returns message recipient name
func (self *DBXReader) GetReceiver(msgNumber int) string {
	return strings.Trim(self.receivers[msgNumber], "\x00")
}

// Returns message recipient address
func (self *DBXReader) GetReceiverAddress(msgNumber int) string {
	return strings.Trim(self.receiverAddresses[msgNumber], "\x00")
}

// Returns message subject
func (self *DBXReader) GetSubject(msgNumber int) string {
	return strings.Trim(self.subjects[msgNumber], "\x00")
}

// Returns message body
func (self *DBXReader) GetMessage(msgNumber int) string {
	index := self.GetIndex(msgNumber)
	var size uint32
	var count uint8

	_, _ = self.f.Seek(int64(index)+4, 0)
	_ = binary.Read(self.f, binary.LittleEndian, &size)
	_, _ = self.f.Seek(2, 1)
	_ = binary.Read(self.f, binary.LittleEndian, &count)
	_, _ = self.f.Seek(1, 1)

	var msgOffset uint32
	var msgOffsetPtr uint32
	var value uint32
	var t uint8
	var i uint8

	for i = 0; i < count; i++ {
		t = 0
		_ = binary.Read(self.f, binary.LittleEndian, &t)
		value = 0
		b := []byte{}
		var bt byte
		for n := 0; n < 3; n++ {
			_ = binary.Read(self.f, binary.LittleEndian, &bt)
			b = append(b, bt)
		}
		b = append(b, 0)
		value = binary.LittleEndian.Uint32(b)
		if t == 0x84 {
			msgOffset = value
			break
		}
		if t == 0x04 {
			msgOffsetPtr = uint32(int(index) + 12 + int(value) + 4*int(count))
			break
		}
	}

	var blockSize uint16

	if msgOffset == 0 && msgOffsetPtr != 0 {
		_, _ = self.f.Seek(int64(msgOffsetPtr), 0)
		_ = binary.Read(self.f, binary.LittleEndian, &msgOffset)
	}

	i2 := msgOffset
	var oldIndex uint32
	for {
		if i2 == 0 {
			break
		}
		if int64(i2) >= self.fSize {
			panic("Bad seek")
		}
		oldIndex = i2
		_, _ = self.f.Seek(int64(i2)+8, 0)
		blockSize = 0
		_ = binary.Read(self.f, binary.LittleEndian, &blockSize)
		_, _ = self.f.Seek(2, 1)
		_ = binary.Read(self.f, binary.LittleEndian, &i2)
		if i2 == oldIndex {
			panic("Generic exception")
		}
	}

	buf := ""

	var pos uint32
	var bbuf []byte

	i2 = msgOffset
	it := 0
	pl := []byte{}
	for i2 != 0 {
		_, _ = self.f.Seek(int64(i2)+8, 0)
		blockSize = 0
		_ = binary.Read(self.f, binary.LittleEndian, &blockSize)
		_, _ = self.f.Seek(2, 1)
		_ = binary.Read(self.f, binary.LittleEndian, &i2)
		bbuf = make([]byte, blockSize)
		self.f.Read(bbuf)
		pl = append(pl, bbuf...)
		pos += uint32(blockSize)
		it++
	}
	buf += string(pl)
	return buf
}

// Returns message receive date
func (self *DBXReader) GetReceiveDate(msgNumber int) time.Time {
	return self.receiveDates[msgNumber]
}

// Returns message send date
func (self *DBXReader) GetSendDate(msgNumber int) time.Time {
	return self.sendDates[msgNumber]
}
