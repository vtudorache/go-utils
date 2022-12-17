// Package getopt provides simple command-line argument parsing, similar to
// the C function getopt described by POSIX. Optional arguments are not
// supported.
package getopt

import (
	"errors"
	"strings"
)

const (
	// The value EndOption is returned when a non-option argument
	// is encountered or there are no arguments left.
	EndOption = -1
)

var (
	// ErrOption is returned when an invalid option is encountered.
	ErrOption = errors.New("getopt: option not supported")
	// ErrNoArg is returned when a required option argument is missing.
	ErrNoArg  = errors.New("getopt: no argument given")
)

// A Parser holds the slice of strings containing the arguments given on the
// command line (the first one being the program's name), the index of the
// argument currently processed, the position of the next character to parse,
// the options string, a boolean telling whether the last option has an
// argument, a boolean telling whether all options have been parsed.
type Parser struct {
	args     []string // the arguments received by the program (os.Args)
	optIndex int      // the index in args of the current option(s)
	optPos   int      // the position in bytes of the option
	opts     string   // the options definition string
	hasArg   bool     // whether the current option has an argument
	done     bool     // whether all options were parsed
}

// Args returns a slice of strings containing the arguments that were not
// processed yet.
func (p *Parser) Args() []string {
	i := p.optIndex
	if p.hasArg {
		// if there is an option argument, skip it
		i++
	}
	return p.args[i:]
}

// Option returns the next option encountered as a rune and an error value. It
// returns (EndOption, nil) when a non-option argument is seen or arguments
// are exhausted. The error is not nil if the option is not valid or its
// required argument is missing.
func (p *Parser) Option() (rune, error) {
	if p.done {
		return EndOption, nil
	}
	if p.hasArg {
		// if there is an option argument, skip it
		p.optIndex++
		p.optPos = 0
		p.hasArg = false
	}
	if p.optIndex >= len(p.args) {
		p.done = true
		return EndOption, nil
	}
	if p.optPos == 0 {
		s := p.args[p.optIndex]
		if len(s) <= 1 || s[0] != '-' {
			p.done = true
			return EndOption, nil
		}
		if len(s) == 2 && s[1] == '-' {
			p.optIndex++
			p.done = true
			return EndOption, nil
		}
		p.optPos = 1
	}
	b := p.args[p.optIndex][p.optPos]
	p.optPos++
	if p.optPos >= len(p.args[p.optIndex]) {
		p.optIndex++
		p.optPos = 0
	}
	if b <= 0x20 || b == ':' || b == '-' || b >= 0x7f {
		return rune(b), ErrOption
	}
	i := strings.IndexByte(p.opts, b)
	if i < 0 {
		return rune(b), ErrOption
	}
	i++
	if i < len(p.opts) && p.opts[i] == ':' {
		if p.optIndex >= len(p.args) {
			return rune(b), ErrNoArg
		}
		p.hasArg = true
	}
	return rune(b), nil
}

// NewParser returns a pointer to a Parser initialized with the given args
// and opts. The first item of args will be skipped by the parser. This
// allows the direct use of os.Args as args array.
// The opts string has the same format as the one used by the C function
// getopt described by POSIX. Optional arguments are not supported.
func NewParser(args []string, opts string) *Parser {
	return &Parser{args, 1, 0, opts, false, false}
}

// OptArg returns the argument of the last option returned by Option, or the
// empty string if none was given.
func (p *Parser) OptArg() string {
	if !p.hasArg {
		return ""
	}
	return p.args[p.optIndex][p.optPos:]
}
