package tools

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"gopkg.in/yaml.v2"
)

func ExecCollect(name string, args ...string) string {
	var o bytes.Buffer
	cmd := exec.Command(name, args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = &o
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	return strings.TrimSpace(o.String())
}

func ExecPlain(name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}

// ReadUA regex file by name from https://github.com/matomo-org/device-detector
// root directory.
//
// Env var DEVICE_DETECTOR_ROOT is used to set the root path.
func ReadUA(name string, out any) {
	path := filepath.Join(os.Getenv("DEVICE_DETECTOR_ROOT"), "regexes", name)
	f, err := os.ReadFile(path)
	if err != nil {
		println(">>> failed to read  ", path, err.Error())
		println("    set DEVICE_DETECTOR_ROOT")
		println("    to path where you cloned ", "https://github.com/matomo-org/device-detector")
		os.Exit(1)
	}
	err = yaml.Unmarshal(f, out)
	if err != nil {
		Exit("failed to  decode ", path, err.Error())
	}
}

func WriteFile(path string, data []byte) {
	err := os.WriteFile(path, data, 0600)
	if err != nil {
		Exit("failed to write ", path, err.Error())
	}
	println("    write: ", path)
}

func Remove(path string) error {
	err := os.Remove(path)
	if err != nil {
		Exit("failed to delete file ", path, err.Error())
	}
	println("    delete: ", path)
	return nil
}

func Exit(a ...string) {
	println(">>> ", strings.Join(a, " "))
	os.Exit(1)
}

// Table generates nice looking markdown tables
type Table struct {
	bytes.Buffer
	b    bytes.Buffer
	line bytes.Buffer
	tw   tabwriter.Writer
}

func (t *Table) Init(header ...string) *Table {
	t.b.Reset()
	t.tw.Init(&t.b, 0, 0, 1, ' ', tabwriter.AlignRight|tabwriter.Debug)
	dash := make([]string, len(header))
	for i := range dash {
		dash[i] = "----"
	}
	t.Row(header...)
	t.Row(dash...)
	return t
}

func (t *Table) Row(entries ...string) {
	t.line.Reset()
	for i := range entries {
		if i != 0 {
			t.line.WriteRune('\t')
		}
		t.line.WriteString(entries[i])
	}
	t.line.WriteByte('\n')
	io.Copy(&t.tw, &t.line)
}

func (t *Table) Flush() {
	t.tw.Flush()
	t.Buffer.Reset()
	s := bufio.NewScanner(&t.b)
	for s.Scan() {
		fmt.Fprintf(&t.Buffer, "|%s|\n", s.Text())
	}
}
