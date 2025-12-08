// +build ignore

package main

import (
	"encoding/base64"
	"encoding/binary"
	"os"
	"time"
)

// DBX file structure constants
const (
	INDEX_POINTER_OFFSET = 0xE4
	ITEM_COUNT_OFFSET    = 0xC4
)

// Generate a test DBX file with emails containing attachments
func main() {
	os.MkdirAll(".", 0755)

	emails := []struct {
		sender       string
		senderAddr   string
		receiver     string
		receiverAddr string
		subject      string
		body         string
		sendDate     time.Time
		recvDate     time.Time
	}{
		{
			sender:       "Test Sender",
			senderAddr:   "sender@example.com",
			receiver:     "Test Receiver",
			receiverAddr: "receiver@example.com",
			subject:      "Email with text attachment",
			body:         createEmailWithTextAttachment(),
			sendDate:     time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			recvDate:     time.Date(2024, 1, 15, 10, 31, 0, 0, time.UTC),
		},
		{
			sender:       "Another Sender",
			senderAddr:   "another@example.com",
			receiver:     "Test Receiver",
			receiverAddr: "receiver@example.com",
			subject:      "Email with binary attachment",
			body:         createEmailWithBinaryAttachment(),
			sendDate:     time.Date(2024, 2, 20, 14, 45, 0, 0, time.UTC),
			recvDate:     time.Date(2024, 2, 20, 14, 46, 0, 0, time.UTC),
		},
		{
			sender:       "Multi Attach",
			senderAddr:   "multi@example.com",
			receiver:     "Test Receiver",
			receiverAddr: "receiver@example.com",
			subject:      "Email with multiple attachments",
			body:         createEmailWithMultipleAttachments(),
			sendDate:     time.Date(2024, 3, 10, 9, 0, 0, 0, time.UTC),
			recvDate:     time.Date(2024, 3, 10, 9, 1, 0, 0, time.UTC),
		},
	}

	f, err := os.Create("test_attachments.dbx")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// Write DBX header
	header := make([]byte, 0x600)

	// DBX signature for email type
	binary.LittleEndian.PutUint32(header[0:], 0xFE12ADCF)
	binary.LittleEndian.PutUint32(header[4:], 0x6F74FDC5)
	binary.LittleEndian.PutUint32(header[8:], 0x11D1E366)
	binary.LittleEndian.PutUint32(header[12:], 0xC0004E9A)

	// Item count at 0xC4
	binary.LittleEndian.PutUint32(header[ITEM_COUNT_OFFSET:], uint32(len(emails)))

	// Start data after header
	dataOffset := uint32(0x600)

	msgBlockOffsets := make([]uint32, len(emails))
	infoOffsets := make([]uint32, len(emails))

	var messageData []byte

	// Build message blocks
	for i, email := range emails {
		msgData := []byte(email.body)

		// Message block structure (from GetMessage):
		// bytes 0-7: header info
		// bytes 8-9: block size (uint16)
		// bytes 10-11: reserved
		// bytes 12-15: next block pointer
		// bytes 16+: message data

		msgBlockOffsets[i] = dataOffset + uint32(len(messageData))

		blockHeader := make([]byte, 16)
		binary.LittleEndian.PutUint32(blockHeader[0:], msgBlockOffsets[i]) // self reference
		binary.LittleEndian.PutUint32(blockHeader[4:], 0)                  // unused
		binary.LittleEndian.PutUint16(blockHeader[8:], uint16(len(msgData)))
		binary.LittleEndian.PutUint16(blockHeader[10:], 0)
		binary.LittleEndian.PutUint32(blockHeader[12:], 0) // no next block

		messageData = append(messageData, blockHeader...)
		messageData = append(messageData, msgData...)

		// Align to 4 bytes
		for len(messageData)%4 != 0 {
			messageData = append(messageData, 0)
		}
	}

	// Build info records
	// Info record structure (from readInfos):
	// bytes 0-3: self address (index)
	// bytes 4-7: size
	// bytes 8-9: reserved
	// byte 10: count
	// byte 11: reserved
	// bytes 12+: field entries (4 bytes each: type + 3-byte offset)
	// then: data area

	var infoData []byte

	for i, email := range emails {
		infoStart := dataOffset + uint32(len(messageData)) + uint32(len(infoData))
		infoOffsets[i] = infoStart

		// Field types from readInfos:
		// 0x02: send date
		// 0x08: subject
		// 0x0D: sender name
		// 0x0E: sender address
		// 0x12: receive date
		// 0x13: receiver name
		// 0x14: receiver address
		// 0x84: direct message offset (value IS the offset)
		// 0x04: indirect message offset

		type field struct {
			typ  byte
			data []byte
		}

		fields := []field{
			{0x08, append([]byte(email.subject), 0)},
			{0x0D, append([]byte(email.sender), 0)},
			{0x0E, append([]byte(email.senderAddr), 0)},
			{0x13, append([]byte(email.receiver), 0)},
			{0x14, append([]byte(email.receiverAddr), 0)},
			{0x02, timeToFileTime(email.sendDate)},
			{0x12, timeToFileTime(email.recvDate)},
		}

		// Add message pointer as special entry (type 0x84 = direct offset in value)
		fieldCount := len(fields) + 1 // +1 for message pointer

		// Calculate data offsets (relative to start of data area)
		// Data area starts at: 12 + (fieldCount * 4)
		dataAreaStart := 12 + fieldCount*4
		currentOffset := 0
		fieldOffsets := make([]int, len(fields))

		for j, fld := range fields {
			fieldOffsets[j] = currentOffset
			currentOffset += len(fld.data)
		}

		totalSize := dataAreaStart + currentOffset

		// Build info record header (12 bytes)
		rec := make([]byte, 12)
		binary.LittleEndian.PutUint32(rec[0:], infoStart)        // self address
		binary.LittleEndian.PutUint32(rec[4:], uint32(totalSize)) // size
		binary.LittleEndian.PutUint16(rec[8:], 0)                 // reserved
		rec[10] = byte(fieldCount)                                // count at offset 10!
		rec[11] = 0                                               // reserved

		// Add field entries (4 bytes each)
		for j, fld := range fields {
			entry := make([]byte, 4)
			entry[0] = fld.typ
			off := uint32(fieldOffsets[j])
			entry[1] = byte(off)
			entry[2] = byte(off >> 8)
			entry[3] = byte(off >> 16)
			rec = append(rec, entry...)
		}

		// Add message pointer field (0x84 = value is direct offset)
		msgEntry := make([]byte, 4)
		msgEntry[0] = 0x84
		msgOff := msgBlockOffsets[i]
		msgEntry[1] = byte(msgOff)
		msgEntry[2] = byte(msgOff >> 8)
		msgEntry[3] = byte(msgOff >> 16)
		rec = append(rec, msgEntry...)

		// Add field data
		for _, fld := range fields {
			rec = append(rec, fld.data...)
		}

		infoData = append(infoData, rec...)

		// Align to 4 bytes
		for len(infoData)%4 != 0 {
			infoData = append(infoData, 0)
		}
	}

	// Build index table
	indexOffset := dataOffset + uint32(len(messageData)) + uint32(len(infoData))

	// Index structure (from readIndex):
	// bytes 0-7: header (8 bytes seek skip)
	// bytes 8-11: nextTable (uint32)
	// bytes 12-16: skip 5 bytes
	// byte 17: ptrCount (uint8)
	// bytes 18-19: skip 2 bytes
	// bytes 20-23: indexCount (uint32)
	// bytes 24+: entries (12 bytes each: indexPtr, nextTable, indexCount)

	indexHeader := make([]byte, 24)
	binary.LittleEndian.PutUint32(indexHeader[0:], indexOffset) // self address
	binary.LittleEndian.PutUint32(indexHeader[4:], 0)           // unused
	binary.LittleEndian.PutUint32(indexHeader[8:], 0)           // nextTable = 0
	indexHeader[12] = 0
	indexHeader[13] = 0
	indexHeader[14] = 0
	indexHeader[15] = 0
	indexHeader[16] = 0
	indexHeader[17] = byte(len(emails)) // ptrCount
	indexHeader[18] = 0
	indexHeader[19] = 0
	binary.LittleEndian.PutUint32(indexHeader[20:], 0) // indexCount = 0

	var indexData []byte
	indexData = append(indexData, indexHeader...)

	// Add index entries
	for i := range emails {
		entry := make([]byte, 12)
		binary.LittleEndian.PutUint32(entry[0:], infoOffsets[i])
		binary.LittleEndian.PutUint32(entry[4:], 0) // nextTable
		binary.LittleEndian.PutUint32(entry[8:], 0) // indexCount
		indexData = append(indexData, entry...)
	}

	// Update header with index pointer
	binary.LittleEndian.PutUint32(header[INDEX_POINTER_OFFSET:], indexOffset)

	// Write everything
	f.Write(header)
	f.Write(messageData)
	f.Write(infoData)
	f.Write(indexData)

	println("Created test_attachments.dbx with", len(emails), "emails containing attachments")
	println("File size:", 0x600+len(messageData)+len(infoData)+len(indexData))
}

func timeToFileTime(t time.Time) []byte {
	const EPOCH_DIFFERENCE = 11644473600
	const TICKS_PER_SECOND = 10000000

	unix := t.Unix()
	fileTime := (unix + EPOCH_DIFFERENCE) * TICKS_PER_SECOND

	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(fileTime))
	return b
}

func createEmailWithTextAttachment() string {
	boundary := "----=_Part_0_1234567890"
	return "MIME-Version: 1.0\r\n" +
		"From: Test Sender <sender@example.com>\r\n" +
		"To: Test Receiver <receiver@example.com>\r\n" +
		"Subject: Email with text attachment\r\n" +
		"Date: Mon, 15 Jan 2024 10:30:00 +0000\r\n" +
		"Content-Type: multipart/mixed; boundary=\"" + boundary + "\"\r\n" +
		"\r\n" +
		"--" + boundary + "\r\n" +
		"Content-Type: text/plain; charset=utf-8\r\n" +
		"Content-Transfer-Encoding: 7bit\r\n" +
		"\r\n" +
		"This is the main body of the email.\r\n" +
		"It contains a text file attachment.\r\n" +
		"\r\n" +
		"--" + boundary + "\r\n" +
		"Content-Type: text/plain; name=\"document.txt\"\r\n" +
		"Content-Disposition: attachment; filename=\"document.txt\"\r\n" +
		"Content-Transfer-Encoding: 7bit\r\n" +
		"\r\n" +
		"This is the content of the attached text file.\r\n" +
		"It has multiple lines.\r\n" +
		"Line 3 of the attachment.\r\n" +
		"\r\n" +
		"--" + boundary + "--\r\n"
}

func createEmailWithBinaryAttachment() string {
	boundary := "----=_Part_1_9876543210"
	binaryData := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
		0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53,
		0xDE, 0x00, 0x00, 0x00, 0x0C, 0x49, 0x44, 0x41,
		0x54, 0x08, 0xD7, 0x63, 0xF8, 0xFF, 0xFF, 0x3F,
		0x00, 0x05, 0xFE, 0x02, 0xFE, 0xDC, 0xCC, 0x59,
		0xE7, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4E,
		0x44, 0xAE, 0x42, 0x60, 0x82,
	}
	encodedData := base64.StdEncoding.EncodeToString(binaryData)

	return "MIME-Version: 1.0\r\n" +
		"From: Another Sender <another@example.com>\r\n" +
		"To: Test Receiver <receiver@example.com>\r\n" +
		"Subject: Email with binary attachment\r\n" +
		"Date: Tue, 20 Feb 2024 14:45:00 +0000\r\n" +
		"Content-Type: multipart/mixed; boundary=\"" + boundary + "\"\r\n" +
		"\r\n" +
		"--" + boundary + "\r\n" +
		"Content-Type: text/plain; charset=utf-8\r\n" +
		"Content-Transfer-Encoding: 7bit\r\n" +
		"\r\n" +
		"This email contains a binary attachment (image).\r\n" +
		"\r\n" +
		"--" + boundary + "\r\n" +
		"Content-Type: image/png; name=\"test_image.png\"\r\n" +
		"Content-Disposition: attachment; filename=\"test_image.png\"\r\n" +
		"Content-Transfer-Encoding: base64\r\n" +
		"\r\n" +
		encodedData + "\r\n" +
		"\r\n" +
		"--" + boundary + "--\r\n"
}

func createEmailWithMultipleAttachments() string {
	boundary := "----=_Part_2_1111111111"

	textContent := "This is a plain text attachment.\r\nWith multiple lines.\r\n"
	csvContent := "Name,Email,Score\r\nAlice,alice@example.com,95\r\nBob,bob@example.com,87\r\nCharlie,charlie@example.com,92\r\n"
	jsonContent := `{"name": "Test", "values": [1, 2, 3], "active": true}`

	return "MIME-Version: 1.0\r\n" +
		"From: Multi Attach <multi@example.com>\r\n" +
		"To: Test Receiver <receiver@example.com>\r\n" +
		"Subject: Email with multiple attachments\r\n" +
		"Date: Sun, 10 Mar 2024 09:00:00 +0000\r\n" +
		"Content-Type: multipart/mixed; boundary=\"" + boundary + "\"\r\n" +
		"\r\n" +
		"--" + boundary + "\r\n" +
		"Content-Type: text/plain; charset=utf-8\r\n" +
		"Content-Transfer-Encoding: 7bit\r\n" +
		"\r\n" +
		"This email has multiple attachments:\r\n" +
		"1. A text file\r\n" +
		"2. A CSV file\r\n" +
		"3. A JSON file\r\n" +
		"\r\n" +
		"--" + boundary + "\r\n" +
		"Content-Type: text/plain; name=\"readme.txt\"\r\n" +
		"Content-Disposition: attachment; filename=\"readme.txt\"\r\n" +
		"Content-Transfer-Encoding: 7bit\r\n" +
		"\r\n" +
		textContent +
		"\r\n" +
		"--" + boundary + "\r\n" +
		"Content-Type: text/csv; name=\"data.csv\"\r\n" +
		"Content-Disposition: attachment; filename=\"data.csv\"\r\n" +
		"Content-Transfer-Encoding: 7bit\r\n" +
		"\r\n" +
		csvContent +
		"\r\n" +
		"--" + boundary + "\r\n" +
		"Content-Type: application/json; name=\"config.json\"\r\n" +
		"Content-Disposition: attachment; filename=\"config.json\"\r\n" +
		"Content-Transfer-Encoding: 7bit\r\n" +
		"\r\n" +
		jsonContent + "\r\n" +
		"\r\n" +
		"--" + boundary + "--\r\n"
}
