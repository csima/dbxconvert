package main

import (
	"testing"
)

var testFormats = [][]string{
	[]string{
		"$DBXNAME$.mbx",
		"test.mbx",
	},
	[]string{
		"$DBXDATE$.mbx",
		"2016-09-12.mbx",
	},
	[]string{
		"$SNAME_L:2_E:Unknown$ - $SUBJ_L:64_E:No Subject$.eml",
		"ro - subject #1.eml",
	},
	[]string{
		"($RDATE_F:%Y-%m-%d %H-%M-%S$) $RNAME_L:32_E:Unknown$ - $SUBJ_L:64_E:No Subject$.txt",
		"(2016-09-12 01-58-32) test@domain.com - subject #1.txt",
	},
}

func TestFormatFilename(t *testing.T) {
	dbx := &DBXReader{}
	err := dbx.Open("./fixtures/test.dbx")
	if err != nil {
		t.Fatal("Unable to open test file!")
	}

	for _, format := range testFormats {
		if FormatFilename(dbx, 0, format[0]) != format[1] {
			t.Error("Formatting error:", format[0], "converted to", format[1])
		}
	}

	dbx.Close()
}

func TestReplaceFrom(t *testing.T) {
	dbx := &DBXReader{}
	err := dbx.Open("./fixtures/test.dbx")
	if err != nil {
		t.Fatal("Unable to open test file!")
	}
	if len(ReplaceFrom(dbx.GetMessage(0))) != 1195 {
		t.Error("Wrong result of replace from (0)")
	}
	if len(ReplaceFrom(dbx.GetMessage(1))) != 18490 {
		t.Error("Wrong result of replace from (1)")
	}
	dbx.Close()
}
