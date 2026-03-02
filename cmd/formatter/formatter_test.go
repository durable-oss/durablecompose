/*
   Copyright 2020 Docker Compose CLI authors

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package formatter

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"go.uber.org/goleak"
	"gotest.tools/v3/assert"
)

type testStruct struct {
	Name   string
	Status string
}

// Print prints formatted lists in different formats
func TestPrint(t *testing.T) {
	testList := []testStruct{
		{
			Name:   "myName1",
			Status: "myStatus1",
		},
		{
			Name:   "myName2",
			Status: "myStatus2",
		},
	}

	b := &bytes.Buffer{}
	assert.NilError(t, Print(testList, TABLE, b, func(w io.Writer) {
		for _, t := range testList {
			_, _ = fmt.Fprintf(w, "%s\t%s\n", t.Name, t.Status)
		}
	}, "NAME", "STATUS"))
	assert.Equal(t, b.String(), "NAME                STATUS\nmyName1             myStatus1\nmyName2             myStatus2\n")

	b.Reset()
	assert.NilError(t, Print(testList, JSON, b, func(w io.Writer) {
		for _, t := range testList {
			_, _ = fmt.Fprintf(w, "%s\t%s\n", t.Name, t.Status)
		}
	}, "NAME", "STATUS"))
	assert.Equal(t, b.String(), `[{"Name":"myName1","Status":"myStatus1"},{"Name":"myName2","Status":"myStatus2"}]
`)

	b.Reset()
	assert.NilError(t, Print(testList, TemplateLegacyJSON, b, func(w io.Writer) {
		for _, t := range testList {
			_, _ = fmt.Fprintf(w, "%s\t%s\n", t.Name, t.Status)
		}
	}, "NAME", "STATUS"))
	json := b.String()
	assert.Equal(t, json, `{"Name":"myName1","Status":"myStatus1"}
{"Name":"myName2","Status":"myStatus2"}
`)
}

func TestColorsGoroutinesLeak(t *testing.T) {
	goleak.VerifyNone(t)
}

func TestPrintInvalidFormat(t *testing.T) {
	testData := []testStruct{
		{Name: "test", Status: "ok"},
	}
	b := &bytes.Buffer{}
	err := Print(testData, "invalid_format", b, func(w io.Writer) {
		fmt.Fprintf(w, "test\n")
	}, "NAME")
	assert.ErrorContains(t, err, "could not be parsed")
}

func TestPrintEmptyList(t *testing.T) {
	var testList []testStruct
	b := &bytes.Buffer{}
	assert.NilError(t, Print(testList, TABLE, b, func(w io.Writer) {
		for _, t := range testList {
			_, _ = fmt.Fprintf(w, "%s\t%s\n", t.Name, t.Status)
		}
	}, "NAME", "STATUS"))
	assert.Equal(t, b.String(), "NAME                STATUS\n")
}

func TestPrintSingleItem(t *testing.T) {
	testList := []testStruct{
		{Name: "single", Status: "active"},
	}
	b := &bytes.Buffer{}
	assert.NilError(t, Print(testList, JSON, b, func(w io.Writer) {
		for _, t := range testList {
			_, _ = fmt.Fprintf(w, "%s\t%s\n", t.Name, t.Status)
		}
	}, "NAME", "STATUS"))
	assert.Equal(t, b.String(), `[{"Name":"single","Status":"active"}]`+"\n")
}

func TestPrintPrettyFormat(t *testing.T) {
	testList := []testStruct{
		{Name: "name1", Status: "status1"},
	}
	b := &bytes.Buffer{}
	assert.NilError(t, Print(testList, PRETTY, b, func(w io.Writer) {
		for _, t := range testList {
			_, _ = fmt.Fprintf(w, "%s\t%s\n", t.Name, t.Status)
		}
	}, "NAME", "STATUS"))
	assert.Equal(t, b.String(), "NAME                STATUS\nname1               status1\n")
}
