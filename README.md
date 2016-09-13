# dbx-converter - DBX to MBOX converter

This was converted to Go from the original C++ code located here: http://www.ukrebs-software.de/english/dbxconv/dbxconv.html

## Version 1.0.0 (12/09/2016)

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 2 of the License, or (at your option) any later version.

## 1\. Description

This program will extract the messages from an Outlook Express (5.0 - 6.0) mailbox and convert it either to the standard mbox or the Outlook Express eml format. The advantage of saving your mail in mbox format is, that it's a plain text format, which can be read by many mail-clients. Converting to eml format is a convenient way to re-import the messages into Outlook Express.

The handling of eml export is a little bit smarter than the one offered by Outlook Express itself. Outlook Express will overwrite messages with same sender and subject, while dbx-converter enumerates the messages, so you can be sure none is lost due to conversion.

## 2\. Usage

The easiest way to convert Outlook Express dbx-files is to copy the mailboxes to a directory with dbx-converter.exe in it. Do not try to convert folders, which have the same extension (like "Folders.dbx"), it will not work. Still it will do no harm.

Then open a DOS-box and type "dbx-converter *.dbx". This will convert all dbx-files into mbx-files. I'd suggest, that you keep backup copies of the original dbx-files, at least until you have verified, that other mail-clients can read the mbx-files.

**Hint for Entourage users:**  
For Entourage to recognize the mbox-file it is necessary to change the extension to "mbox".

To convert all dbx-files into eml format, type "dbx-converter -eml *.dbx". For each mailbox a folder with the name of the mailbox is created, and all messages will be extracted into the respective folder. To do both, converting and extracting, type "dbx-converter -mbx -eml *.dbx".

## 3\. Available Options

**-mbx[naming]**  
Converts the specified dbx-file into mboxo format. The mboxo format uses simple "From " quoting. Any line of a message starting with "From " is quoted by putting a ">" in front of it. The mbox-file is by default named as the dbx-file but has the extension "mbx".  
For the optional file naming parameter see section 4 of this manual. Make sure you have enclosed this option in quotation marks in case the naming parameter contains spaces!

**-mbxrd[naming]**  
Converts the specified dbx-file into mboxrd format. The mboxrd format uses a more advanced "From " quoting scheme. Any line starting with "From " or any number of ">" followed by "From " is quoted by putting a ">" in front of it.  
For the optional file naming parameter see section 4 of this manual. Make sure you have enclosed this option in quotation marks in case the naming parameter contains spaces!

**-eml[naming]**  
Converts the specified dbx-file into eml format. All eml-files are saved into a new subdirectory named as the dbx-file right under the current working directory or the specified output folder. The names of the eml-files are by default generated from the sender's name and the subject.  
For the optional file naming parameter see section 4 of this manual. Make sure you have enclosed this option in quotation marks in case the naming parameter contains spaces!  
The file date is normally set to the send date of the message, but this can be changed using the date stamp option -rcvdate.

**-dn**  
Inserts a double newline after each message within a mbox file.

**-ic**  
Turns off case sensitivity when quoting "From ". That means, "froM " will also be quoted.  
Normally, "From " quoting is done only if the exact word "From " is found. Some mail clients (e.g. Pegasus Mail) are not case sensitive when parsing the mbox-file. For them it is required to quote any line starting with "from ", regardless of the case.

**-ff**  
Turns off the automatic generation of subdirectories when saving in eml-format. All eml-files from all dbx-files will be saved directly into the current working directory or the specified output folder.

**-senddate**  
This option tells dbx-converter to use the "Date" field of the messages as timestamp. The timestamp is used to set the file date of eml messages and to set the "From" header of the exported mbox.

**-rcvdate**  
With this option, dbx-converter uses the "Received" date of the messages as timestamp.

**-merge**        
Merge multiple input DBX files into single output MBX(MBXRD)
  
**-?**  
Shows a quick reference.

## 4\. File naming for eml and mbox files

From version 1.3.0 on, dbx-converter has a highly configurable naming scheme, allowing to create filenames in a by far more flexible manner than before. With the mbx, mbxrd and the eml option, a format string for generating file names can be supplied. The format string may consist of variables, options and constants. During runtime, variables are evaluated, formatted according to the supplied options and concatenated with the constants to form the final filename for mbox and eml files. Variables including their respective formatting options are enclosed with the dollar symbol "$". The variable name itself must be the first item of a variable definition. Supplied options for that variable are separated by an underscore. So a full variable definition follows this scheme (where the squared brackets show optional components):

<pre>$VARNAME[_OPTIONNAME:OPTIONVALUE][_OPTIONNAME:OPTIONVALUE]...$</pre>

### 4.1 Available variables

**The following variables will work with mbox and eml files:**

**DBXNAME**  
The name of the current dbx file without the trailing dbx extension.

**DBXDATE**  
The last modification date of the dbx file.

**The following variables will work with eml files only:**

**SADDR**  
Email address of the sender.

**RADDR**  
Email address of the receiver.

**SNAME**  
Name of the sender (if no name is specified, the output will be the same as if using SADDR).

**RNAME**  
Name of the receiver (if no name is specified, the output will be the same as if using RADDR).

**SUBJ**  
The subject of the message.

**RDATE**  
The receive date of the message.

**SDATE**  
The send date of the message.

### 4.2 Variable formatting

Formatting options will be supplied to the variables seperated by underscores "_". Each option is declared with a prefix, telling which option is to set. Then follows a colon and finally the value for the option.  
The following options are available:

**L**  
Specifies the maximum length of this entry. If the resulting string from evaluating the variable is longer then this number, it'll be truncated. Per default the output string may have unlimited size.

**N**  
Specifies the maximum number of items that'll be included into the output string. This options works only for the RADDR and RNAME variables, where the output might have multiple entries (i.e. multiple receivers).

**C**  
Allows capitalization of the resulting string. If this option is set to 0, all characters are converted to lower case. If this option is set to 1, all characters are converted to upper case. Per default the case remains unchanged.

**E**  
If the resulting string is empty, it will be replaced by the string defined by this option. The default setting is an empty string.

**F**  
Makes only sense with the date variables. The date output will be formatted according to the format string defined with this option. The format string may consist of constants and variables. The following variables are available:

%y - Two digit year  
%Y - Four digit year  
%m - Month  
%d - Day of month  
%H - Hour (24)  
%I - Hour (12)  
%p - AM/PM  
%M - Minutes  
%S - Second  
%b - Short month name  
%B - Full month name  
%a - Short weekday name  
%A - Full weekday name  
%W - Weeknumber

The default setting for this option is %Y-%m-%d.

### 4.3 Examples for file name formatting

To form a complete file name definition you can mix literal strings, variables as required.

The default setting for mbox file names is:

<pre>$DBXNAME$.mbx</pre>

This means, the file name for the mbox is generated from the name of the dbx file with an appended mbx extension.

The default setting for eml file names is:

<pre>$SNAME_L:32_E:Unknown$ - $SUBJ_L:64_E:No Subject$.eml</pre>

This means, the file name for eml files is generated from the first 32 characters of the sender name, where a "Unknown" is inserted in case there is no sender information. Then the literal string " - " is inserted, followed by the first 64 characters of the subject, where a "No Subject" is inserted in case the message has no subject. Finally the extension ".eml" is appended.

If you want to prepend the message date and time, use the receiver instead and make the emails behave like text files, the file name pattern for eml files would look as follows:

<pre>($RDATE_F:%Y-%m-%d %H-%M-%S$) $RNAME_L:32_E:Unknown$ - $SUBJ_L:64_E:No Subject$.txt</pre>

To inlude this file naming option into the full command, you'll have to write

<pre>dbx-converter "-eml($RDATE_F:%Y-%m-%d %H-%M-%S$) $RNAME_L:32_E:Unknown$ - $SUBJ_L:64_E:No Subject$.txt" *.dbx</pre>

Please note the quotation marks, which are absolutely necessary to make this command work.
