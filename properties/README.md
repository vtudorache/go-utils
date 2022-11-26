# Overview

Package properties manages persistent property tables. It behaves like
Java's Properties class.

# Index

[type Table](#type-table)  
[func NewTable(data map[string]string) *Table](#func-newtable)  
[func NewTableDefaults(defaults *Table) *Table](#func-newtabledefaults)  
[func (p *Table) Clear()](#func-p-table-clear)  
[func (p *Table) ClearAll()](#func-p-table-clearall)  
[func (p *Table) Delete(key string)](#func-p-table-delete)  
[func (p *Table) Get(key string) string](#func-p-table-get)  
[func (p *Table) Keys() []string](#func-p-table-keys)  
[func (p *Table) Load(r io.Reader) (int, error)](#func-p-table-load)  
[func (p *Table) LoadString(s string) (int, error)](#func-p-table-load-string)  
[func (p *Table) Lookup(key string) (string, bool)](#func-p-table-lookup)  
[func (p *Table) Save(w io.Writer, comments string, ascii bool) (int, error)](#func-p-table-save)  
[func (p *Table) SaveString(comments string, ascii bool) (string, error)](#func-p-table-savestring)  
[func (p *Table) Set(key string, value string)](#func-p-table-set)  
[func (p *Table) Store(w io.Writer, ascii bool) (int, error)](#func-p-table-store)  
[func (p *Table) String() string](#func-p-table-string)  

## type Table
```
type Table struct {
    // contains filtered or unexported fields
}
```
Table represents a property table. It contains:
- a map of key-value pairs;
- a pointer to a 'defaults' property table.
The 'defaults' table is searched if a property key isn't found within
the first one.

## func NewTable
```
func NewTable(data map[string]string) *Table
```  
NewTable creates and initializes a property table using the given data.
The new table takes ownership of data, which shouldn't be used after
this call.

## func NewTableDefaults
```
func NewTableDefaults(defaults *Table) *Table  
```
NewTableDefaults creates and initializes a property table using defaults
for the 'defaults' table. The new table takes ownership of defaults, which
shouldn't be used after this call.

## func (p *Table) Clear  
```
func (p *Table) Clear()
```
Clear deletes all the key-value pairs in this table. It doesn't alter
the pairs in the 'defaults' table.

## func (p *Table) ClearAll  
```
func (p *Table) ClearAll()
```
ClearAll deletes all the key-value pairs from this table and its 'defaults'
table (if present).

## func (p *Table) Delete
```
func (p *Table) Delete(key string)
```
Delete removes the key and the associated value from the property table. If 
the key isn't present, calling this function does nothing.

## func (p *Table) Get
```
func (p *Table) Get(key string) string  
```
Get returns the value associated with the string key. If key isn't present
in this table, it searches the 'defaults' table. If the key isn't found,
returns the empty string.

## func (p *Table) Keys
```
func (p *Table) Keys() []string  
```
Keys returns all the distinct keys in this table and the 'defaults' table.

## func (p *Table) Load
```
func (p *Table) Load(r io.Reader) (int, error)
```
Load reads the key-value pairs from r. Reading is done line by line. There 
are natural lines and logical lines. A natural line is a character sequence 
ending either by the standard end-of-line ('\n', '\r' or '\r\n') or by the 
end-of-file. A natural line may hold only a part of a key-value pair. A 
logical line holds all the data of a key-value pair. It may spread across
several adjacent natural lines by escaping the end-of-line with the backslash
character '\\'.  
Lines are read from the input until the end-of-file is reached.
A natural line containing only white space characters is considered blank
and is ignored.  
A comment line has an ASCII '#' or '!' as the first non-space character. 
Comment lines are ignored and do not encode data. A comment line can't spread 
across several natural lines.  
The characters ' ' ('\u0020'), '\t' ('\u0009'), and '\f' ('\u000C')
are considered white space.  
If a logical line spreads on several natural lines, the backslash escaping
the end-of-line sequence, the end-of-line sequence itself and any white space
character at the beginning of the following line are discarded. There must 
be an odd number of contiguous backslash characters for the end-of-line to be 
escaped. A number of 2n consecutive backslashes encodes n backslashes after 
unescaping.  
The key contains all of the characters in the line starting with the first
non-space character and up to, but not including, the first unescaped '=',
':', or white space character other than end-of-line. All of these
key-value delimiters may be included in the key by escaping them with a
backslash character. For example,
```
\=\:\=
```
would be the key "=:=". End-of-line can be included using '\r' and '\n'
sequences. Any space character after the key is skipped. The first '=', ':' 
or white space after the key is ignored and any space characters after it are 
also skipped. All the remaining characters on the line become part of the 
associated value. If there are no more characters in the line, the value is 
the empty string "". As an example, each of the following lines specifies the 
key "Go" and the associated value "The Best Language":
```
Go = The Best Language
    Go:The Best Language
Go                    :The Best Language
```
As another example, the following lines specify a single property:
```
languages                       Assembly, Lisp, Pascal, \
                                BASIC, C, \
                                Perl, Tcl, Lua, Java, Python, \
                                C#, Go
```
The key is "languages" and the associated value is: "Assembly, Lisp, Pascal, 
BASIC, C, Perl, Tcl, Lua, Java, Python, C#, Go". Note that a space appears 
before each '\\' so that a space will appear in the final result; the '\\', 
the end-of-line and the leading white space on the continuation line are 
discarded and not replaced by other characters.  
Other escapes are not recognized, excepting Unicode sequences like '\u0009'. 
Only a single 'u' character is allowed in a Unicode escape sequence. Unicode 
runes above 0xffff should be stored as two consecutive '\uxxxx' sequeces 
encoding the surrogates.
A backslash character within an invalid escape sequence is not an error, the 
backslash is silently dropped. Escapes are not necessary for single and 
double quotes, but single and double quote characters preceded by a backslash
yield single and double quote characters, respectively.   
The method returns the number of key-value pairs loaded and any error
encountered.

## func (p *Table) LoadString  
```
func (p *Table) LoadString(s string) (int, error)  
```
LoadString loads a property table using the given string as input.  
The method returns the number of key-value pairs loaded and any error
encountered.

## func (p *Table) Lookup  
```
func (p *Table) Lookup(key string) (string, bool)
```
Lookup searches the value associated with key. If key isn't present in this 
table, the function searches the 'defaults' table.  
The method returns the value (or the empty string) and a boolean indicating 
whether the value was found or not.

## func (p *Table) Save  
```
func (p *Table) Save(w io.Writer, comments string, ascii bool) (int, error)  
```
Save writes the key-value pairs of this property table to w in a format
suitable for using the Load method.  
If comments is not empty, then an ASCII '#' character, the comments string 
and an end-of-line are first written to w. Any sequence of end-of-line
characters is replaced with only one end-of-line. If the character following
end-of-line in comments is not '#' or '!', then an ASCII '#' is written out
after the end-of-line.  
The key-value pairs are then written using the Store method.  
The method returns the number of key-value pairs written and any error
encountered.

## func (p *Table) SaveString
```  
func (p *Table) SaveString(comments string, ascii bool) (string, error)  
```
SaveString returns the text form of the property table and any error
encountered. The ascii parameter has the same meaning as for the Store
function above.

## func (p *Table) Set  
```
func (p *Table) Set(key string, value string)  
```
Set associates key with value in this table. If key is already
present, then the associated value is replaced.

## func (p *Table) Store
```  
func (p *Table) Store(w io.Writer, ascii bool) (int, error)
```
Store writes the key-value pairs of this property table to w in a format
suitable for using the Load method.  
The key-value pairs in the 'defaults' table (if any) are not written out by 
this method.  
If ascii is true, then any rune lesser than 0x20 or greater than 0x7e
is converted to its '\uxxxx'  escape sequence(s).  
Every key-value pair in the table is then written out, one per line. For 
each pair the key is written, then an ASCII '=', then the associated value.
For the key, all space characters are written with a preceding '\\'
character. For the value, only the leading white space characters are 
written with a preceding '\\' character. The key and value characters '#', 
'!', '=', and ':' are written with a preceding '\\' to ensure that they are 
properly loaded.  
The method returns the number of key-value pairs written and any error
encountered.

## func (p *Table) String  
```
func (p *Table) String() string
```  
String returns the (UTF-8) text representation of the property table (not
including the key-value pairs of the 'defaults' table). The text can be
then reused by LoadString.

