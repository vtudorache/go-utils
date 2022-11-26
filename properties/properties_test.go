package properties

import (
	"sort"
	"testing"
)

func TestLoadString(t *testing.T) {
	var p Table
	p.LoadString(`firstKey=firstValue
		second\ key = second value
		third\ key third \
		  	extended value
		fourth\ key\ : \ fourth value
		fifth\ key = fifth value with \u20ac`)
	if p.Get("firstKey") != "firstValue" {
		t.Error(`p.Get("firstKey") != "firstValue"`)
	}
	if p.Get("second key") != "second value" {
		t.Error(`p.Get("second key") != "second value"`)
	}
	if p.Get("third key") != "third extended value" {
		t.Error(`p.Get("third key") != "third extended value"`)
	}
	if p.Get("fourth key ") != " fourth value" {
		t.Error(`p.Get("fourth key ") != " fourth value"`)
	}
	if p.Get("fifth key") != "fifth value with €" {
		t.Error(`p.Get("fifth key") != "fifth value with €"`)
	}
}

func TestKeys(t *testing.T) {
	var p Table
	p.LoadString(`1\ firstKey=firstValue
		2\ second\ key = second value
		3\ third\ key third \
		  	extended value
		4\ fourth\ key\ : \ fourth value
		5\ fifth\ key = fifth value with \u20ac`)
	k := p.Keys()
	sort.Strings(k)
	if k[0] != "1 firstKey" {
		t.Error(`"1 firstKey" != "`, k[0], `"`)
	}
	if k[1] != "2 second key" {
		t.Error(`"2 second key" != "`, k[1], `"`)
	}
	if k[2] != "3 third key" {
		t.Error(`"3 third key" != "`, k[2], `"`)
	}
	if k[3] != "4 fourth key " {
		t.Error(`"4 fourth key " != "`, k[3], `"`)
	}
	if k[4] != "5 fifth key" {
		t.Error(`"5 fifth key" != "`, k[4], `"`)
	}
}

func TestSaveString(t *testing.T) {
	var p Table
	var s string
	p.Set("firstKey", "firstValue")
	s, _ = p.SaveString("The first\r\nproperties entry", false)
	if s != "#The first\r\n#properties entry\nfirstKey=firstValue\n" {
		t.Error("SaveString() returned ", s)
	}
	p.Clear()
	p.Set("second key", "second value")
	s, _ = p.SaveString("!The second property", false)
	if s != "!The second property\nsecond\\ key=second value\n" {
		t.Error("SaveString() returned ", s)
	}
	p.Clear()
	p.Set("third #key", "third !value")
	s, _ = p.SaveString("The third property", false)
	if s != "#The third property\nthird\\ \\#key=third \\!value\n" {
		t.Error("SaveString() returned ", s)
	}
	p.Clear()
	p.Set("fourth \n#key", "fourth !value")
	s, _ = p.SaveString("The fourth property", false)
	if s != "#The fourth property\nfourth\\ \\n\\#key=fourth \\!value\n" {
		t.Error("SaveString() returned ", s)
	}
	p.Clear()
	p.Set("fifth key", "fifth value with €")
	s, _ = p.SaveString("The fifth property", true)
	if s != "#The fifth property\nfifth\\ key=fifth value with \\u20ac\n" {
		t.Error("SaveString() returned ", s)
	}
	p.Clear()
	p.Set("sixth key", "sixth value with 😀 objects")
	s, _ = p.SaveString("The sixth property", true)
	if s != "#The sixth property\nsixth\\ key=sixth value with \\ud83d\\ude00 objects\n" {
		t.Error("SaveString() returned ", s)
	}
}

func TestDefaults(t *testing.T) {
	var p = new(Table)
	var s string
	p.LoadString("firstKey=firstValue")
	p.LoadString("second\\ key = second value")
	p.LoadString("third\\ key third \\\n  \textended value")
	p.LoadString("fourth\\ key\\ : \\ fourth value\n")
	p = NewTableDefaults(p)
	if p.Get("firstKey") != "firstValue" {
		t.Error(`p.Get("firstKey") != "firstValue"`)
	}
	if p.Get("second key") != "second value" {
		t.Error(`p.Get("second key") != "second value"`)
	}
	if p.Get("third key") != "third extended value" {
		t.Error(`p.Get("third key") != "third extended value"`)
	}
	if p.Get("fourth key ") != " fourth value" {
		t.Error(`p.Get("fourth key ") != " fourth value"`)
	}
	s, _ = p.SaveString("Table with defaults", false)
	if s != "#Table with defaults\n" {
		t.Error("SaveString() returned ", s)
	}
	p.Set("fourth key", "a new fourth value")
	s, _ = p.SaveString("Table with defaults", false)
	if s != "#Table with defaults\nfourth\\ key=a new fourth value\n" {
		t.Error("SaveString() returned ", s)
	}
}
