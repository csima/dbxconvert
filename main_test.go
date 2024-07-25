package main

import (
	"fmt"
	"testing"
	"time"
)

var testFormats = [][]string{
	[]string{
		"$DBXNAME$.mbx",
		"test.mbx",
	},
	// Broken due to time issues. Git doesn't store create / modified times.
	//[]string{
	//	"$DBXDATE$.mbx",
	//	"2016-09-12.mbx",
	//},
	[]string{
		"$SNAME_L:2_E:Unknown$ - $SUBJ_L:64_E:No Subject$.eml",
		"ro - subject #1.eml",
	},
	[]string{
		"($RDATE_F:%Y-%m-%d %H-%M-%S$) $RNAME_L:32_E:Unknown$ - $SUBJ_L:64_E:No Subject$.txt",
		"(" + time.Date(2016, 9, 11, 22, 58, 32, 0, time.UTC).Local().Format("2006-01-02 15-04-05") + ") test@domain.com - subject #1.txt",
	},
}

// Time date issues.
func TestFormatFilename(t *testing.T) {
	dbx := &DBXReader{}
	err := dbx.Open("./fixtures/test.dbx")
	if err != nil {
		t.Fatal("Unable to open test file!")
	}

	for ti, format := range testFormats {
		t.Run(fmt.Sprintf("%d", ti), func(t *testing.T) {
			actual := FormatFilename(dbx, 0, format[0])
			if actual != format[1] {
				t.Errorf("Formatting error: %#v didn't convert to %#v but converted to %#v", format[0], format[1], actual)
			}
		})
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
