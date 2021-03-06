//
// Copyright © 2021 Kris Nóva <kris@nivenly.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
//   ███╗   ██╗ █████╗ ███╗   ███╗██╗
//   ████╗  ██║██╔══██╗████╗ ████║██║
//   ██╔██╗ ██║███████║██╔████╔██║██║
//   ██║╚██╗██║██╔══██║██║╚██╔╝██║██║
//   ██║ ╚████║██║  ██║██║ ╚═╝ ██║███████╗
//   ╚═╝  ╚═══╝╚═╝  ╚═╝╚═╝     ╚═╝╚══════╝

package tests

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/kris-nova/naml"
)

func TestManifests(t *testing.T) {
	files, err := ioutil.ReadDir("manifests")
	if err != nil {
		t.Errorf("unable to list test_nivenly.yaml directory: %v", err)
	}
	for _, file := range files {
		t.Logf("testing [%s]", file.Name())
		err := generateCompileRunYAML(filepath.Join("manifests", file.Name()))
		if err != nil {
			t.Errorf(err.Error())
			t.FailNow()
		}
	}
	t.Logf("Manifest tests complete")
}

func CodifyValues(name string) *naml.CodifyValues {
	return &naml.CodifyValues{
		AppNameLower:  "app",
		AppNameTitle:  "App",
		AuthorName:    "Björn Nóva",
		AuthorEmail:   "barnaby@nivenly.com",
		CopyrightYear: "1999",
		Description:   "Test Description.",
		PackageName:   "main",
	}
}

// generateCompileRunYAML will build a Go program from YAML and try to compile and run it :)
func generateCompileRunYAML(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("unable to read file: %s: %v", filename, err)
	}
	buffer := bytes.Buffer{}
	buffer.Write(data)
	output, err := naml.Codify(&buffer, CodifyValues(filename))
	if err != nil && len(output) < 0 {
		return fmt.Errorf("unable to codify: %s: %v", filename, err)
	} else if err != nil && len(output) > 0 {
		// Delta measured! We could only codify some.
	}
	program, err := naml.Compile(output)
	if program != nil {
		defer program.Remove()
	}
	if err != nil {
		return fmt.Errorf("unable to compile: %s: %v", filename, err)
	}
	//	_, stderr, err := program.Execute([]string{"list"})
	_, stderr, err := program.Execute([]string{""})
	if stderr.Len() > 0 {
		return fmt.Errorf("failed executing binary: %s: %v: %s", filename, err, stderr.String())
	}
	if err != nil {
		return fmt.Errorf("failed executing binary: %s: %v", filename, err)
	}
	return nil
}
