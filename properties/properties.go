// Package properties manages persistent property tables. It behaves like
// Java's Properties class.
package properties

import (
	"bufio"
	"bytes"
	"io"
	"strings"
	"unicode/utf16"
	"unicode/utf8"
)

// escapeRune writes into p the '\uxxxx' sequence representing the rune. If
// the rune is out of range or if no escaping is needed, writes the escape
// sequence of utf8.RuneError.
// If the rune is greater than 0xffff, writes the '\uxxxx' sequences of the
// two surrogates.
// The function returns the number of bytes written.
func escapeRune(p []byte, r rune) int {
	if r > 0xffff {
		r1, r2 := utf16.EncodeRune(r)
		return escapeRune(p, r1) + escapeRune(p[6:], r2)
	}
	if 0x20 <= r && r <= 0x7e {
		return 0
	}
	if r < 0 {
		r = utf8.RuneError
	}
	p[0] = '\\'
	p[1] = 'u'
	for i := 5; i >= 2; i-- {
		b := byte(0x0f & r)
		if b > 9 {
			b += 'a' - 10
		} else {
			b += '0'
		}
		p[i] = b
		r >>= 4
	}
	return 6
}

// unescapeRune parses the first escape sequence in p. It recognizes the
// sequences '\t', '\n', '\f', '\r', '\uxxxx'. If a '\uxxxx' sequence holds
// a surrogate, a second '\uxxxx' sequence must be present, holding the
// next surrogate.
// The function returns the rune and number of bytes parsed. If p doesn't
// start with an escape sequence, returns utf8.RuneError and 0.
func unescapeRune(p []byte) (rune, int) {
	n := len(p)
	if n < 1 || p[0] != '\\' {
		return utf8.RuneError, 0
	}
	r, size := utf8.DecodeRune(p[1:])
	if r == 't' {
		return '\t', 2
	}
	if r == 'n' {
		return '\n', 2
	}
	if r == 'f' {
		return '\f', 2
	}
	if r == 'r' {
		return '\r', 2
	}
	if r != 'u' {
		return r, size + 1
	}
	if n > 6 {
		n = 6
	}
	r = 0
	for i := 2; i < n; i++ {
		b := p[i]
		if '0' <= b && b <= '9' {
			b -= '0'
		} else if 'a' <= b && b <= 'f' {
			b -= 'a' - 10
		} else if 'A' <= b && b <= 'F' {
			b -= 'A' - 10
		} else {
			n = i
			break
		}
		r = (r << 4) | rune(b)
	}
	if n < 6 {
		r = utf8.RuneError
	}
	// here, n = 6 (the length of a '\uxxxx' sequence)
	if utf16.IsSurrogate(r) {
		q := r
		r, size = unescapeRune(p[6:])
		if size != 6 || !utf16.IsSurrogate(r) {
			return utf8.RuneError, 6
		}
		r = utf16.DecodeRune(q, r)
		n = 12
	}
	return r, n
}

func isDelimiter(r rune) bool {
	return (r == '=' || r == ':')
}

func isSpace(r rune) bool {
	return (r == '\t' || r == '\f' || r == ' ')
}

func isCommentPrefix(r rune) bool {
	return (r == '#' || r == '!')
}

// unescape replaces the escape sequences in p with the real characters. If
// split is true, the parser stops at the beginning of the value in a
// key-value pair.
// If split is true, then the function returns the key string (excluding any
// key-value separator) and the position in p where the parsing should be
// resumed.
// If split is false, then the function returns the full string contained
// in p and the number of bytes parsed.
func unescape(p []byte, split bool) (string, int) {
	var b strings.Builder
	n := 0
	for len(p) > 0 {
		r, size := unescapeRune(p)
		if size == 0 {
			r, size = utf8.DecodeRune(p)
			if split && (isSpace(r) || isDelimiter(r)) {
				p = p[size:]
				n += size
				for len(p) > 0 {
					r, size = utf8.DecodeRune(p)
					if !(isSpace(r) || isDelimiter(r)) {
						return b.String(), n
					}
					p = p[size:]
					n += size
				}
			}
		}
		b.WriteRune(r)
		p = p[size:]
		n += size
	}
	return b.String(), n
}

func loadBytes(r *bufio.Reader) ([]byte, error) {
	var b []byte
	done := false
	for !done {
		x, e := r.ReadByte()
		if e != nil {
			return b, e
		}
		for x == '\t' || x == '\f' || x == ' ' {
			x, e = r.ReadByte()
			if e != nil {
				return b, e
			}
		}
		if (x == '#' || x == '!') && len(b) == 0 {
			done = true
		}
		esc := false
		for x != '\n' && x != '\r' {
			if x == '\\' {
				esc = !esc
			} else {
				esc = false
			}
			b = append(b, x)
			x, e = r.ReadByte()
			if e != nil {
				return b, e
			}
		}
		if x == '\r' {
			x, e = r.ReadByte()
			if e != nil {
				return b, e
			}
			if x != '\n' {
				e = r.UnreadByte()
				if e != nil {
					return b, e
				}
			}
		}
		if !done {
			if esc {
				b = b[:len(b)-1]
			} else {
				done = true
			}
		}
	}
	return b, nil
}

// Table represents a property table. It contains:
// - a map of key-value pairs;
// - a pointer to a 'defaults' property table.
// The 'defaults' table is searched if a property key isn't found within
// the first one.
type Table struct {
	data     map[string]string
	defaults *Table
}

// Load reads the key-value pairs from r. Reading is done line by line. There
// are natural lines and logical lines. A natural line is a character sequence
// ending either by the standard end-of-line ('\n', '\r' or '\r\n') or by the
// end-of-file. A natural line may hold only a part of a key-value pair. A
// logical line holds all the data of a key-value pair. It may spread across
// several adjacent natural lines by escaping the end-of-line with the backslash
// character '\\'.
// Lines are read from the input until the end-of-file is reached.
// A natural line containing only white space characters is considered blank
// and is ignored.
// A comment line has an ASCII '#' or '!' as the first non-space character.
// Comment lines are ignored and do not encode data. A comment line can't spread
// across several natural lines.
// The characters ' ' ('\u0020'), '\t' ('\u0009'), and '\f' ('\u000C')
// are considered white space.
// If a logical line spreads on several natural lines, the backslash escaping
// the end-of-line sequence, the end-of-line sequence itself and any white space
// character at the beginning of the following line are discarded. There must
// be an odd number of contiguous backslash characters for the end-of-line to be
// escaped. A number of 2n consecutive backslashes encodes n backslashes after
// unescaping.
// The key contains all of the characters in the line starting with the first
// non-space character and up to, but not including, the first unescaped '=',
// ':', or white space character other than end-of-line. All of these
// key-value delimiters may be included in the key by escaping them with a
// backslash character. For example,
// ```
// \=\:\=
// ```
// would be the key "=:=". End-of-line can be included using '\r' and '\n'
// sequences. Any space character after the key is skipped. The first '=', ':'
// or white space after the key is ignored and any space characters after it are
// also skipped. All the remaining characters on the line become part of the
// associated value. If there are no more characters in the line, the value is
// the empty string "". As an example, each of the following lines specifies the
// key "Go" and the associated value "The Best Language":
// ```
// Go = The Best Language
//     Go:The Best Language
// Go                    :The Best Language
// ```
// As another example, the following lines specify a single property:
// ```
// languages                       Assembly, Lisp, Pascal, \
//                                 BASIC, C, \
//                                 Perl, Tcl, Lua, Java, Python, \
//                                 C#, Go
// ```
// The key is "languages" and the associated value is: "Assembly, Lisp, Pascal,
// BASIC, C, Perl, Tcl, Lua, Java, Python, C#, Go". Note that a space appears
// before each '\\' so that a space will appear in the final result; the '\\',
// the end-of-line and the leading white space on the continuation line are
// discarded and not replaced by other characters.
// Other escapes are not recognized, excepting Unicode sequences like '\u0009'.
// Only a single 'u' character is allowed in a Unicode escape sequence. Unicode
// runes above 0xffff should be stored as two consecutive '\uxxxx' sequeces
// encoding the surrogates.
// A backslash character within an invalid escape sequence is not an error, the
// backslash is silently dropped. Escapes are not necessary for single and
// double quotes, but single and double quote characters preceded by a backslash
// yield single and double quote characters, respectively.
// The method returns the number of key-value pairs loaded and any error
// encountered.
func (p *Table) Load(r io.Reader) (int, error) {
	var reader = bufio.NewReader(r)
	count := 0
	done := false
	if p.data == nil {
		p.data = make(map[string]string)
	}
	for !done {
		b, e := loadBytes(reader)
		if len(b) > 0 && b[0] != '#' && b[0] != '!' {
			key, i := unescape(b, true)
			value, _ := unescape(b[i:], false)
			p.data[key] = value
			count += 1
		}
		if e != nil {
			if e != io.EOF {
				return count, e
			}
			done = true
		}
	}
	return count, nil
}

// LoadString loads a property table using the given string as input.
// The method returns the number of key-value pairs loaded and any error
// encountered.
func (p *Table) LoadString(s string) (int, error) {
	r := strings.NewReader(s)
	return p.Load(r)
}

func escape(key, value string, ascii bool) []byte {
	var b bytes.Buffer
	var buffer [12]byte
	var r rune
	for _, r = range key {
		size := 0
		if ascii {
			size = escapeRune(buffer[:], r)
		}
		if size == 0 {
			if r == '\n' {
				b.WriteString("\\n")
				continue
			}
			if r == '\r' {
				b.WriteString("\\r")
				continue
			}
			if isSpace(r) || isDelimiter(r) || isCommentPrefix(r) {
				b.WriteByte('\\')
			}
			size = utf8.EncodeRune(buffer[:], r)
		}
		b.Write(buffer[:size])
	}
	b.WriteRune('=')
	r, _ = utf8.DecodeRuneInString(value)
	if isSpace(r) || isDelimiter(r) {
		b.WriteByte('\\')
	}
	for _, r = range value {
		size := 0
		if ascii {
			size = escapeRune(buffer[:], r)
		}
		if size == 0 {
			if r == '\n' {
				b.WriteString("\\n")
				continue
			}
			if r == '\r' {
				b.WriteString("\\r")
				continue
			}
			if isCommentPrefix(r) {
				b.WriteByte('\\')
			}
			size = utf8.EncodeRune(buffer[:], r)
		}
		b.Write(buffer[:size])
	}
	return b.Bytes()
}

func escapeComment(text string, ascii bool) []byte {
	var b bytes.Buffer
	var buffer [12]byte
	last := rune('\n')
	for _, r := range text {
		if r == '\n' || r == '\r' {
			b.WriteRune(r)
			last = r
			continue
		}
		if (last == '\n' || last == '\r') && !isCommentPrefix(r) {
			b.WriteByte('#')
		}
		size := 0
		if ascii {
			size = escapeRune(buffer[:], r)
		}
		if size == 0 {
			size = utf8.EncodeRune(buffer[:], r)
		}
		b.Write(buffer[:size])
		last = r
	}
	return b.Bytes()
}

// Store writes the key-value pairs of this property table to w in a format
// suitable for using the Load method.
// The key-value pairs in the 'defaults' table (if any) are not written out by
// this method.
// If ascii is true, then any rune lesser than 0x20 or greater than 0x7e
// is converted to its '\uxxxx'  escape sequence(s).
// Every key-value pair in the table is then written out, one per line. For
// each pair the key is written, then an ASCII '=', then the associated value.
// For the key, all space characters are written with a preceding '\\'
// character. For the value, only the leading white space characters are
// written with a preceding '\\' character. The key and value characters '#',
// '!', '=', and ':' are written with a preceding '\\' to ensure that they are
// properly loaded.
// The method returns the number of key-value pairs written and any error
// encountered.
func (p *Table) Store(w io.Writer, ascii bool) (int, error) {
	count := 0
	eol := []byte("\n")
	for key, value := range p.data {
		if _, e := w.Write(escape(key, value, ascii)); e != nil {
			return count, e
		}
		if _, e := w.Write(eol); e != nil {
			return count, e
		}
		count += 1
	}
	return count, nil
}

// Save writes the key-value pairs of this property table to w in a format
// suitable for using the Load method.
// If comments is not empty, then an ASCII '#' character, the comments string
// and an end-of-line are first written to w. Any sequence of end-of-line
// characters is replaced with only one end-of-line. If the character following
// end-of-line in comments is not '#' or '!', then an ASCII '#' is written out
// after the end-of-line.
// The key-value pairs are then written using the Store method.
// The method returns the number of key-value pairs written and any error
// encountered.
func (p *Table) Save(w io.Writer, comments string, ascii bool) (int, error) {
	eol := []byte("\n")
	if _, e := w.Write(escapeComment(comments, ascii)); e != nil {
		return 0, e
	}
	if _, e := w.Write(eol); e != nil {
		return 0, e
	}
	return p.Store(w, ascii)
}

// SaveString returns the text form of the property table and any error
// encountered. The ascii parameter has the same meaning as for the Store
// function above.
func (p *Table) SaveString(comments string, ascii bool) (string, error) {
	var b strings.Builder
	_, e := p.Save(&b, comments, ascii)
	return b.String(), e
}

// String returns the (UTF-8) text representation of the property table (not
// including the key-value pairs of the 'defaults' table). The text can be
// then reused by LoadString.
func (p *Table) String() string {
	var b strings.Builder
	eol := []byte("\n")
	for key, value := range p.data {
		b.Write(escape(key, value, false))
		b.Write(eol)
	}
	return b.String()
}

// NewTableDefaults creates and initializes a property table using defaults
// for the 'defaults' table. The new table takes ownership of defaults, which
// shouldn't be used further after this call.
func NewTableDefaults(defaults *Table) *Table {
	return &Table{
		map[string]string{},
		defaults,
	}
}

// NewTable creates and initializes a property table using the given data.
// The new table takes ownership of data, which shouldn't be used after
// this call.
func NewTable(data map[string]string) *Table {
	return &Table{
		data,
		nil,
	}
}

// Lookup searches the value associated with key. If key isn't present in this
// table, the function searches the 'defaults' table.
// The method returns the value (or the empty string) and a boolean indicating
// whether the value was found or not.
func (p *Table) Lookup(key string) (string, bool) {
	if value, found := p.data[key]; found {
		return value, true
	}
	if p.defaults != nil {
		if value, found := p.defaults.Lookup(key); found {
			return value, true
		}
	}
	return "", false
}

// Get returns the value associated with the string key. If key isn't present
// in this table, it searches the 'defaults' table. If the key isn't found,
// returns the empty string.
func (p *Table) Get(key string) string {
	value, _ := p.Lookup(key)
	return value
}

// Set associates key with value in this table. If key is already present,
// then the associated value is replaced.
func (p *Table) Set(key string, value string) {
	if p.data == nil {
		p.data = make(map[string]string)
	}
	p.data[key] = value
}

// Delete removes the key and the associated value from this table. If the key
// isn't present, calling this function does nothing.
func (p *Table) Delete(key string) {
	delete(p.data, key)
}

// Clear deletes all the key-value pairs in this table. It doesn't alter the
// pairs in the 'defaults' table.
func (p *Table) Clear() {
	p.data = make(map[string]string)
}

// ClearAll deletes all the key-value pairs from this table and from the
// 'defaults' table (if present).
func (p *Table) ClearAll() {
	p.Clear()
	if p.defaults != nil {
		p.defaults.ClearAll()
	}
}

// Keys returns all the distinct keys in this table and the 'defaults' table.
func (p *Table) Keys() []string {
	t := make(map[string]bool)
	for p != nil {
		for k := range p.data {
			if !t[k] {
				t[k] = true
			}
		}
		p = p.defaults
	}
	s := make([]string, len(t))
	i := 0
	for k := range t {
		s[i] = k
		i++
	}
	return s
}
