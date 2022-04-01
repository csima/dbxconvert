package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Generates file name accordingly to provided format
// More about formats in README file
func FormatFilename(dbx *DBXReader, msgNr int, format string) (outString string) {
	var varName string
	dateFormatReplacer := strings.NewReplacer(
		"%y", "06",
		"%Y", "2006",
		"%m", "01",
		"%d", "02",
		"%H", "15",
		"%I", "03",
		"%p", "PM",
		"%M", "04",
		"%S", "05",
		"%b", "Jan",
		"%B", "January",
		"%a", "Mon",
		"%A", "Monday",
	)

	outString = format
	rx := regexp.MustCompile(`\$(.*?)\$`)
	submatches := rx.FindAllStringSubmatch(format, -1)
	var tokens []string
	for _, submatch := range submatches {
		varName = submatch[1]
		tokens = []string{}
		for _, part := range strings.Split(varName, "_") {
			tokens = append(tokens, part)
		}

		var s string
		//		var parsePos int
		var parseOK bool
		var maxLength int
		var capitalization int
		var numItems int
		var empty string
		var dateFormat string = "2006-01-02"

		for _, token := range tokens {
			if strings.Index(token, "L:") == 0 {
				maxLength, _ = strconv.Atoi(token[2:])
			} else if strings.Index(token, "C:") == 0 {
				capitalization, _ = strconv.Atoi(token[2:])
			} else if strings.Index(token, "E:") == 0 {
				empty = token[2:]
			} else if strings.Index(token, "F:") == 0 {
				dateFormat = token[2:]
				dateFormat = dateFormatReplacer.Replace(dateFormat)
			} else if strings.Index(token, "N:") == 0 {
				numItems, _ = strconv.Atoi(token[2:])
			}
		}

		if tokens[0] == "DBXNAME" {
			s = dbx.GetFileName()
			pos := strings.LastIndex(s, ".")
			if pos >= 0 {
				s = s[:pos]
			}
			pos = strings.LastIndex(s, string(os.PathSeparator))
			if pos >= 0 {
				s = s[pos:]
			}
			pos = strings.LastIndex(s, "/")
			if pos >= 0 {
				s = s[pos:]
			}
			parseOK = true
		} else if tokens[0] == "DBXDATE" {
			s = dbx.GetFileDate().Format(dateFormat)
			parseOK = true
		} else {
			if msgNr >= 0 {
				if tokens[0] == "SADDR" {
					s = dbx.GetSenderAddress(msgNr)
					parseOK = true
				} else if tokens[0] == "RADDR" {
					s = dbx.GetReceiverAddress(msgNr)
					parseOK = true
				} else if tokens[0] == "SNAME" {
					s = dbx.GetSender(msgNr)
					parseOK = true
				} else if tokens[0] == "RNAME" {
					s = dbx.GetReceiver(msgNr)
					parseOK = true
				} else if tokens[0] == "SUBJ" {
					s = dbx.GetSubject(msgNr)
					parseOK = true
				} else if tokens[0] == "RDATE" {
					s = dbx.GetReceiveDate(msgNr).Format(dateFormat)
					parseOK = true
				} else if tokens[0] == "SDATE" {
					s = dbx.GetSendDate(msgNr).Format(dateFormat)
					parseOK = true
				}
			}
		}

		if parseOK {
			if numItems > 0 && (tokens[0] == "RADDR" || tokens[0] == "RNAME") {
				buf := s
				pos := 0
				item := 0

				for pos < len(buf) {
					if buf[pos] == ';' {
						item++
						if item == numItems {
							break
						}
					}
					pos++
				}

				s = buf[0:pos]
			}

			if empty != "" && s == "" {
				s = empty
			}

			if maxLength > 0 && maxLength < 256 {
				if len(s) > maxLength {
					s = s[:maxLength]
				}
			}

			if capitalization == 0 {
				s = strings.ToLower(s)
			} else if capitalization == 1 {
				s = strings.ToUpper(s)
			}
			outString = strings.Replace(outString, submatch[0], s, -1)
		}
	}

	replacer := strings.NewReplacer(
		"<", " ",
		">", " ",
		"|", " ",
		"?", " ",
		`\`, " ",
		`/`, " ",
		`"`, " ",
		":", " ",
		"*", " ",
		"\r", " ",
		"\n", " ",
	)
	outString = replacer.Replace(outString)
	rx2 := regexp.MustCompile("[ +]")
	outString = rx2.ReplaceAllString(outString, " ")
	return strings.TrimSpace(outString)
}

// Escapes From-s in message body
func ReplaceFrom(s string) string {
	findPos := 0
	for {
		findPos = strings.Index(s, "\n") + 1
		if findPos != 0 {
			break
		}

		if findPos >= len(s) {
			break
		}

		if argMbxRd != "" {
			for {
				if s[findPos] != '>' {
					break
				} else {
					findPos++
				}
			}
		}

		if argIc {
			if strings.ToLower(s[findPos:findPos+5]) == "from " {
				s = s[0:findPos] + ">" + s[findPos:]
			}
		} else {
			if s[findPos:findPos+5] == "From " {
				s = s[0:findPos] + ">" + s[findPos:]
			}
		}
	}
	return s
}

var (
	argMbx       string = ""
	argMbxRd     string = ""
	argDn        bool   = false
	argIc        bool   = false
	argFf        bool   = false
	argEml       string = ""
	argRcvDate   bool   = false
	argOverwrite bool   = false
	argHelp      bool   = false
	//argSendDate  bool   = false
	argMerge bool   = false
	fileSpec string = ""
	outDir   string = ""

	mbxFilenameFormat string = "$DBXNAME$.mbx"
	emlFilenameFormat string = "$SNAME_L:32_E:Unknown$ - $SUBJ_L:64_E:No Subject$.eml"
)

const (
	EXPECT_NONE   = 0
	EXPECT_OPTION = 1
	EXPECT_INFILE = 2
	EXPECT_OUTDIR = 4
)

func parseArguments() {
	expect := EXPECT_OPTION | EXPECT_INFILE
	for i := 1; i < len(os.Args); i++ {
		s := os.Args[i]
		if (expect&EXPECT_OPTION != 0) && (s[0] == '-') {
			if strings.ToLower(s) == "-dn" {
				argDn = true
				continue
			}

			if strings.ToLower(s) == "-ic" {
				argIc = true
				continue
			}

			if strings.ToLower(s) == "-ff" {
				argFf = true
				continue
			}

			if len(s) > 5 && strings.ToLower(s[:6]) == "-mbxrd" {
				if len(s) > 6 {
					argMbxRd = strings.Trim(strings.TrimSpace(s[6:]), `"`)
				} else {
					argMbxRd = mbxFilenameFormat
				}
				continue
			}

			if len(s) > 3 && strings.ToLower(s[:4]) == "-mbx" {
				if len(s) > 4 {
					argMbx = strings.Trim(strings.TrimSpace(s[4:]), `"`)
				} else {
					argMbx = mbxFilenameFormat
				}
				continue
			}

			if len(s) > 3 && strings.ToLower(s[:4]) == "-eml" {
				if len(s) > 4 {
					argEml = strings.Trim(strings.TrimSpace(s[4:]), `"`)
				} else {
					argEml = emlFilenameFormat
				}
				continue
			}

			if strings.ToLower(s) == "-rcvdate" {
				argRcvDate = true
				continue
			}

			if strings.ToLower(s) == "-senddate" {
				//argSendDate = true
				continue
			}

			if strings.ToLower(s) == "-overwrite" {
				argOverwrite = true
				continue
			}

			if strings.ToLower(s) == "-merge" {
				argMerge = true
				continue
			}

			if strings.ToLower(s) == "-?" {
				argHelp = true
				break
			}

			fmt.Printf("Unknown option: %s\n", s)
			return
		}

		if (expect & EXPECT_INFILE) != 0 {
			fileSpec = s
			expect = EXPECT_OUTDIR
			continue
		}

		if (expect & EXPECT_OUTDIR) != 0 {
			outDir = s
			expect = EXPECT_NONE
			continue
		}
	}
}

func displayHelp() {
	fmt.Println("DBX to MBOX and EML Converter, Ver. 1.0.0, by ***")
	fmt.Println()
	fmt.Println("Usage: dbxconv [Options] Infile [Output directory]")
	fmt.Println()
	fmt.Println("Available options:")
	fmt.Println("  -mbx[naming]   Saves all messages in mboxo-format (default)")
	fmt.Println("  -mbxrd[naming] Saves all messages in mboxrd-format")
	fmt.Println("  -dn            Always add a double newline at the end of a message")
	fmt.Println("  -ic            Ignore case when replacing 'From'")
	fmt.Println("  -ff            Flat folders, do not create subdirectories for eml files")
	fmt.Println("  -eml[naming]   Saves all messages in eml-format")
	fmt.Println("                 By default, all messages will be saved into a")
	fmt.Println("                 subdirectory named as the dbx-file.")
	fmt.Println("                 Use -ff to store all messages into one folder.")
	fmt.Println("  -senddate      Uses send date as timestamp (default)")
	fmt.Println("  -rcvdate       Uses receive date as timestamp")
	fmt.Println("  -overwrite     Overwrite existing messages with same name")
	fmt.Println("  -merge         Merge multiple input DBX files into single output MBX(MBXRD)")
	fmt.Println("  -?             Shows this message")
}

func main() {
	parseArguments()
	if len(os.Args) == 1 || argHelp {
		displayHelp()
		return
	}

	filePaths := []string{}

	fi, err := os.Stat(fileSpec)
	if os.IsNotExist(err) {
		fmt.Println("Nothing to do!")
		return
	}

	if !fi.IsDir() {
		fp, err := filepath.Abs(fileSpec)
		if err == nil {
			filePaths = append(filePaths, fp)
		}
	} else {
		files, err := ioutil.ReadDir(fileSpec)
		if err == nil {
			if fileSpec[len(fileSpec)-1] == os.PathSeparator {
				fileSpec = fileSpec[:len(fileSpec)-1]
			}
			for _, f := range files {
				fp, err := filepath.Abs(fileSpec + string(os.PathSeparator) + f.Name())
				if err == nil {
					filePaths = append(filePaths, fp)
				}
			}
		}
	}

	if len(filePaths) == 0 {
		fmt.Println("Nothing to do!")
		return
	}

	if (argMbx == "") && (argMbxRd == "") && (argEml == "") {
		fmt.Println("No conversion options selected, using MBOXO-converter...")
		argMbx = mbxFilenameFormat
	}

	mboxSuccess := 0
	emlSuccess := 0
	fileName := ""
	oemFileName := ""
	outFileDir := ""
	outFileName := ""
	outFilePath := ""
	oemOutFilePath := ""
	var outFile *os.File

	for _, filePath := range filePaths {
		dbx := &DBXReader{}
		err := dbx.Open(filePath)
		if err != nil {
			fmt.Printf("Unable to open file: %s\n", filePath)
			continue
		}
		oemFileName = dbx.GetFName()

		if dbx.GetType() != DBX_TYPE_EMAIL {
			if dbx.GetType() == DBX_TYPE_OE4 {
				fmt.Println("Unable to handle Outlook Express mailboxes below ver. 5!")
			} else {
				fmt.Printf("\"%s\" is not an Outlook Express mailbox!\n", oemFileName)
			}
			dbx.Close()
			continue
		}

		if dbx.GetItemCount() == 0 {
			fmt.Printf("\"%s\" contains no messages, skipped...\n", oemFileName)
			dbx.Close()
			continue
		}

		fmt.Printf("Processing mailbox \"%s\" (%d messages)...\n", oemFileName, dbx.GetItemCount())

		// if output format MBX or MBXRD
		if argMbx != "" || argMbxRd != "" {
			var pos int
			if argMbxRd != "" {
				outFileName = FormatFilename(dbx, -1, argMbxRd)
			} else {
				outFileName = FormatFilename(dbx, -1, argMbx)
			}

			if !argMerge || outFile == nil {
				outFilePath = ""
				if outDir == "" {
					outFilePath = filePath
					pos = strings.LastIndex(outFilePath, string(os.PathSeparator))
					if pos != -1 {
						outFilePath = outFilePath[0:pos]
					}
					if outFilePath[len(outFilePath)-1] != os.PathSeparator {
						outFilePath += string(os.PathSeparator)
					}
					outFilePath = outFilePath + outFileName
				} else {
					if outDir[len(outDir)-1] != os.PathSeparator {
						outDir += string(os.PathSeparator)
					}
					outFilePath = outDir + outFileName
				}

				oemOutFilePath = outFilePath
				outFile, err = os.OpenFile(outFilePath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0777)
				if err != nil {
					fmt.Printf("Unable to open output file \"%s\"!\n", oemOutFilePath)
					continue
				}
			}

			for j := 0; j < dbx.GetItemCount(); j++ {
				var s string
				var t time.Time

				if argRcvDate {
					t = dbx.GetReceiveDate(j)
				} else {
					t = dbx.GetSendDate(j)
				}
				ts := t.Format("Mon Jan 02 15:04:05 2006")
				sender := dbx.GetSenderAddress(j)
				if sender == "" {
					sender = "-"
				}
				s = "From " + sender + " " + ts
				_, _ = outFile.WriteString(s)
				_, _ = outFile.WriteString("\r\n")

				s = dbx.GetMessage(j)

				s = ReplaceFrom(s)
				_, _ = outFile.WriteString(s)
				_, _ = outFile.WriteString("\r\n")

				if argDn || (len(s) > 0 && s[len(s)-1] != '\n') {
					_, _ = outFile.WriteString("\r\n")
				}
			}

			mboxSuccess++
			if !argMerge && outFile != nil {
				outFile.Close()
			}
		}

		// if output format EML
		if argEml != "" {
			var pos int
			outFileDir = ""
			if len(outDir) == 0 {
				outFileDir = filePath
				pos = strings.LastIndex(outFileDir, string(os.PathSeparator))
				if pos != -1 {
					outFileDir = outFileDir[:pos]
				}
			} else {
				outFileDir = outDir
			}

			if outFileDir[len(outFileDir)-1] != os.PathSeparator {
				outFileDir += string(os.PathSeparator)
			}

			if !argFf {
				pos = strings.LastIndex(fileName, ".")
				if pos != -1 {
					outFileDir += fileName[:pos]
				} else {
					outFileDir += fileName
				}
				if outFileDir[len(outFileDir)-1] != os.PathSeparator {
					outFileDir += string(os.PathSeparator)
				}
			}

			err = os.Mkdir(outFileDir, 0777)
			if err != nil && !os.IsExist(err) {
				fmt.Printf("Unable to create output folder: %s\n", outFileDir)
				continue
			}

			processedMails := 0

			for j := 0; j < dbx.GetItemCount(); j++ {
				outFileName = FormatFilename(dbx, j, argEml)
				outFilePath = outFileDir + outFileName

				if !argOverwrite {
					k := 1
					for {
						testFile, err := os.OpenFile(outFilePath, os.O_RDONLY, 0777)
						if err != nil {
							break
						}
						testFile.Close()
						pos = strings.LastIndex(outFileName, ".")
						if pos != -1 {
							outFilePath = outFileDir + outFileName[:pos] + "(" + fmt.Sprintf("%d", k) + ")" + outFileName[pos:]
						} else {
							outFilePath = outFileDir + outFileName + "(" + fmt.Sprintf("%d", k) + ")"
						}
						k++
					}
				}

				oemOutFilePath = outFilePath

				outFileEml, err := os.OpenFile(outFilePath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0777)
				if err != nil {
					fmt.Printf("Unable to open output file \"%s\"!\n", oemOutFilePath)
					continue
				}

				s := dbx.GetMessage(j)
				_, _ = outFileEml.WriteString(s)
				outFileEml.Close()

				if argRcvDate {
					_ = os.Chtimes(outFilePath, dbx.GetReceiveDate(j), dbx.GetReceiveDate(j))
				} else {
					_ = os.Chtimes(outFilePath, dbx.GetSendDate(j), dbx.GetSendDate(j))
				}
				processedMails++
			}
			if processedMails == dbx.GetItemCount() {
				emlSuccess++
			}
		}

		dbx.Close()
	}

	if argMerge && outFile != nil {
		outFile.Close()
	}

	if argMbx != "" {
		fmt.Printf("%d of %d mailboxes converted to MBOXO-format!\n", mboxSuccess, len(filePaths))
	}

	if argMbxRd != "" {
		fmt.Printf("%d of %d mailboxes converted to MBOXRD-format!\n", mboxSuccess, len(filePaths))
	}

	if argEml != "" {
		fmt.Printf("%d of %d mailboxes converted to EML-format!\n", emlSuccess, len(filePaths))
	}
}
