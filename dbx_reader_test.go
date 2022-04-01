package main

import (
	"strings"
	"testing"
)

func TestDBXReader_Open(t *testing.T) {
	dbx := &DBXReader{}
	err := dbx.Open("./fixtures/test.dbx")
	if err != nil {
		t.Fatal("Unable to open test file!")
	}
	dbx.Close()
}

func TestDBXReader_Close(t *testing.T) {
	dbx := &DBXReader{}
	err := dbx.Open("./fixtures/test.dbx")
	if err != nil {
		t.Fatal("Unable to open test file!")
	}
	err = dbx.Close()
	if err != nil {
		t.Fatal("Unable to close test file!")
	}
}

func TestDBXReader_GetFileDate(t *testing.T) {
	dbx := &DBXReader{}
	err := dbx.Open("./fixtures/test.dbx")
	if err != nil {
		t.Fatal("Unable to open test file!")
	}
	if dbx.GetFileDate().Format("2006-01-02 15:04:05") != "2016-09-12 02:04:02" {
		t.Fatal("DBX File date is wrong!")
	}
	dbx.Close()
}

func TestDBXReader_GetType(t *testing.T) {
	dbx := &DBXReader{}
	err := dbx.Open("./fixtures/test.dbx")
	if err != nil {
		t.Fatal("Unable to open test file!")
	}
	if dbx.GetType() != DBX_TYPE_EMAIL {
		t.Fatal("File is not Outlook Express DBX file!")
	}
	dbx.Close()
}

func TestDBXReader_GetIndex(t *testing.T) {
	dbx := &DBXReader{}
	err := dbx.Open("./fixtures/test.dbx")
	if err != nil {
		t.Fatal("Unable to open test file!")
	}
	if dbx.GetIndex(0) != 11044 {
		t.Fatal("Wrongs index(0)!")
	}
	if dbx.GetIndex(1) != 13016 {
		t.Fatal("Wrongs index(1)!")
	}
	dbx.Close()
}

func TestDBXReader_GetMessage(t *testing.T) {
	dbx := &DBXReader{}
	err := dbx.Open("./fixtures/test.dbx")
	if err != nil {
		t.Fatal("Unable to open test file!")
	}
	if len(dbx.GetMessage(0)) != 1195 {
		t.Fatal("Wrong message(0)!", len(dbx.GetMessage(0)))
	}
	if len(dbx.GetMessage(1)) != 18490 {
		t.Fatal("Wrong message(1)!", len(dbx.GetMessage(1)))
	}
	dbx.Close()
}

func TestDBXReader_GetFileName(t *testing.T) {
	dbx := &DBXReader{}
	err := dbx.Open("./fixtures/test.dbx")
	if err != nil {
		t.Fatal("Unable to open test file!")
	}
	if dbx.GetFileName() != "./fixtures/test.dbx" {
		t.Fatal("Wrong DBX file name:", dbx.GetFileName())
	}
	dbx.Close()
}

func TestDBXReader_GetItemCount(t *testing.T) {
	dbx := &DBXReader{}
	err := dbx.Open("./fixtures/test.dbx")
	if err != nil {
		t.Fatal("Unable to open test file!")
	}
	if dbx.GetItemCount() != 2 {
		t.Fatal("Wrong item count:", dbx.GetItemCount())
	}
	dbx.Close()
}

func TestDBXReader_GetReceiveDate(t *testing.T) {
	dbx := &DBXReader{}
	err := dbx.Open("./fixtures/test.dbx")
	if err != nil {
		t.Fatal("Unable to open test file!")
	}
	if dbx.GetReceiveDate(0).Unix() != 1473634712 {
		t.Fatal("Wrong receive date(0)!")
	}
	if dbx.GetReceiveDate(1).Unix() != 1473634973 {
		t.Fatal("Wrong receive date(1)!")
	}
	dbx.Close()
}

func TestDBXReader_GetReceiver(t *testing.T) {
	dbx := &DBXReader{}
	err := dbx.Open("./fixtures/test.dbx")
	if err != nil {
		t.Fatal("Unable to open test file!")
	}
	if dbx.GetReceiver(0) != "test@domain.com" {
		t.Fatal("Wrong reciever(0)!")
	}
	if dbx.GetReceiver(1) != "test-2@domain.com" {
		t.Fatal("Wrong reciever(1)!")
	}
	dbx.Close()
}

func TestDBXReader_GetReceiverAddress(t *testing.T) {
	dbx := &DBXReader{}
	err := dbx.Open("./fixtures/test.dbx")
	if err != nil {
		t.Fatal("Unable to open test file!")
	}
	if dbx.GetReceiverAddress(0) != "<test@domain.com>" {
		t.Fatal("Wrong reciever address(0)!")
	}
	if dbx.GetReceiverAddress(1) != "<test-2@domain.com>" {
		t.Fatal("Wrong reciever address(1)!")
	}
	dbx.Close()
}

func TestDBXReader_GetSendDate(t *testing.T) {
	dbx := &DBXReader{}
	err := dbx.Open("./fixtures/test.dbx")
	if err != nil {
		t.Fatal("Unable to open test file!")
	}
	if dbx.GetSendDate(0).Unix() != 1473634712 {
		t.Fatal("Wrong send date(0)!")
	}
	if dbx.GetSendDate(1).Unix() != 1473634973 {
		t.Fatal("Wrong send date(1)!")
	}
	dbx.Close()
}

func TestDBXReader_GetSender(t *testing.T) {
	dbx := &DBXReader{}
	err := dbx.Open("./fixtures/test.dbx")
	if err != nil {
		t.Fatal("Unable to open test file!")
	}
	if strings.Index(dbx.GetSender(0), "Ro") != 0 {
		t.Fatal("Wrong sender(0)!")
	}
	if strings.Index(dbx.GetSender(1), "Ro") != 0 {
		t.Fatal("Wrong sender(1)!")
	}
	dbx.Close()
}

func TestDBXReader_GetSenderAddress(t *testing.T) {
	dbx := &DBXReader{}
	err := dbx.Open("./fixtures/test.dbx")
	if err != nil {
		t.Fatal("Unable to open test file!")
	}
	if !strings.Contains(dbx.GetSenderAddress(0), "@mail") {
		t.Fatal("Wrong sender address(0)!")
	}
	if !strings.Contains(dbx.GetSenderAddress(1), "@mail") {
		t.Fatal("Wrong sender address(1)!")
	}
	dbx.Close()
}

func TestDBXReader_GetSubject(t *testing.T) {
	dbx := &DBXReader{}
	err := dbx.Open("./fixtures/test.dbx")
	if err != nil {
		t.Fatal("Unable to open test file!")
	}
	if dbx.GetSubject(0) != "Subject #1" {
		t.Fatal("Wrong subject(0)!")
	}
	if dbx.GetSubject(1) != "Subject #2" {
		t.Fatal("Wrong subject(0)!")
	}
	dbx.Close()
}

func TestDBXReader_GetFName(t *testing.T) {
	dbx := &DBXReader{}
	err := dbx.Open("./fixtures/test.dbx")
	if err != nil {
		t.Fatal("Unable to open test file!")
	}
	if dbx.GetFName() != "test.dbx" {
		t.Fatal("Wrong file name!")
	}
	dbx.Close()
}
