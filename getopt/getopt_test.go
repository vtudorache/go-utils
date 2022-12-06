package getopt

import (
	"strings"
	"testing"
)

var aFlag = false
var bOption = ""

func TestOptions(t *testing.T) {
	args := []string{"test", "-ab", "cdef"}
	opts := "ab:"
	p := NewParser(args, opts)
	end := false
	for !end {
		switch o, e := p.Next(); o {
		case 'a':
			aFlag = true
		case 'b':
			bOption = p.OptArg()
		default:
			if e != nil {
				t.Errorf("%s: -%c", e, o)
			}
			end = true
		}
	}
	if !aFlag {
		t.Error("aFlag should have been set")
	}
	if strings.Compare(bOption, "cdef") != 0 {
		t.Error("bOption should have been \"cdef\"")
	}
}

func TestBadOptions(t *testing.T) {
	args := []string{"test", "-ab", "cdef", "-x"}
	opts := "ab:"
	p := NewParser(args, opts)
	end := false
	for !end {
		switch o, e := p.Next(); o {
		case 'a':
			aFlag = true
		case 'b':
			bOption = p.OptArg()
		case 'x':
			if e == nil {
				t.Error("option should not be defined: -x")
			}
		default:
			if e != nil {
				t.Errorf("%s: -%c", e, o)
			}
			end = true
		}
	}
	if !aFlag {
		t.Error("aFlag should have been set")
	}
	if strings.Compare(bOption, "cdef") != 0 {
		t.Error("bOption should have been \"cdef\"")
	}
}