package javaclassparser

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/yaklang/yaklang/common/javaclassparser/classes"
	"github.com/yaklang/yaklang/common/javaclassparser/decompiler/core"
	"os"
	"strings"
	"testing"
)

func TestArrayAnfIfClass(t *testing.T) {
	classesContent, err := classes.FS.ReadFile("test/array_if_test.class")
	if err != nil {
		t.Fatal(err)
	}
	expectSource, err := classes.FS.ReadFile("test/array_if_test.java")
	if err != nil {
		t.Fatal(err)
	}
	cf, err := Parse(classesContent)
	if err != nil {
		t.Fatal(err)
	}
	source, err := cf.Dump()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, string(expectSource), source)
}
func TestDemoClass(t *testing.T) {
	classesContent, err := classes.FS.ReadFile("test/Demo.class")
	if err != nil {
		t.Fatal(err)
	}
	expectSource, err := classes.FS.ReadFile("test/Demo.java")
	if err != nil {
		t.Fatal(err)
	}
	cf, err := Parse(classesContent)
	if err != nil {
		t.Fatal(err)
	}
	source, err := cf.Dump()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, string(expectSource), source)
}
func TestAddSupperInterface(t *testing.T) {
	classesContent, _ := os.ReadFile("/Users/z3/Downloads/cfr-master/src/org/benf/cfr/reader/ForTraditionTest.class")
	cf, err := Parse(classesContent)
	if err != nil {
		t.Fatal(err)
	}
	source, err := cf.Dump()
	if err != nil {
		t.Fatal(err)
	}
	println(source)
}

func TestModifyOpcode(t *testing.T) {
	classesContent, err := classes.FS.ReadFile("Demo.class")
	if err != nil {
		t.Fatal(err)
	}
	cf, err := Parse(classesContent)
	if err != nil {
		t.Fatal(err)
	}
	codeAttr := cf.Methods[1].Attributes[0].(*CodeAttribute)
	ParseBytesCode(nil, codeAttr)
}
func TestParseRawType(t *testing.T) {
	content, _ := classes.FS.ReadFile("raw_type.json")
	data := []*core.RawJavaType{}
	json.Unmarshal(content, &data)
	items := []string{}
	for _, datum := range data {
		items = append(items, fmt.Sprintf(`RT_%s = NewRawJavaType("%v","%v",%v,%v,"%v",%v,%v,%v,%v)`,
			strings.ToUpper(datum.Name), datum.Name, datum.SuggestedVarName, "ST_"+strings.ToUpper(datum.StackType.Name),
			datum.UsableType, datum.BoxedName, datum.IsNumber, datum.IsObject, datum.IntMin, datum.IntMax))
	}
	println(strings.Join(items, "\n"))
}

func TestParseStackType(t *testing.T) {
	content, _ := classes.FS.ReadFile("stack_type.json")
	data := []*core.StackType{}
	json.Unmarshal(content, &data)
	items := []string{}
	for _, datum := range data {
		items = append(items, fmt.Sprintf(`ST_%s = NewStackType(%v,%v,"%v")`, strings.ToUpper(datum.Name), datum.ComputationCategory, datum.Closed, datum.Name))
	}
	println(strings.Join(items, "\n"))
}