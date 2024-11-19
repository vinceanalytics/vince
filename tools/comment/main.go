package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
)

func main() {
	flag.Parse()
	switch flag.Arg(0) {
	case "code":
		code(os.Stdout, os.Stdin)
	}
}

func code(out io.Writer, in io.Reader) {
	r := bufio.NewScanner(in)
	for r.Scan() {
		fmt.Fprintf(out, "\t%s\n", r.Text())
	}
}
