package getopt

import (
	"errors"
	"strings"
)

const (
	EndOption = -1
)

var (
	ErrOption = errors.New("getopt: option not supported")
	ErrNoArg  = errors.New("getopt: no argument given")
)

type Parser struct {
	args     []string // the arguments received by the program (os.Args)
	optIndex int      // the index in args of the current option(s)
	optPos   int      // the position in bytes of the option
	opts     string   // the options definition string
	hasArg   bool     // whether the current option has an argument
}

func (p *Parser) Args() []string {
	i := p.optIndex
	if p.hasArg {
		// if there is an option argument, skip it
		i++
	}
	return p.args[i:]
}

func (p *Parser) Next() (rune, error) {
	if p.hasArg {
		// if there is an option argument, skip it
		p.optIndex++
		p.optPos = 0
		p.hasArg = false
	}
	if p.optIndex >= len(p.args) {
		return EndOption, nil
	}
	if p.optPos == 0 {
		s := p.args[p.optIndex]
		if len(s) <= 1 || s[0] != '-' {
			return EndOption, nil
		}
		if len(s) == 2 && s[1] == '-' {
			p.optIndex++
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

func NewParser(args []string, opts string) *Parser {
	return &Parser{args, 1, 0, opts, false}
}

func (p *Parser) OptArg() string {
	if !p.hasArg {
		return ""
	}
	return p.args[p.optIndex][p.optPos:]
}
