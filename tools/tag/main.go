package main

import (
	"log"
	"os"
	"os/exec"
)

func main() {
	v := os.Getenv("VERSION")
	if v == "" {
		return
	}
	err := os.WriteFile("version/VERSION", []byte(v), 0600)
	if err != nil {
		log.Fatal(err)
	}
	run("git", "commit", "-am", "release "+v)
	run("git", "tag", "-a", v, "-m", "release "+v)
}

func run(name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Env = os.Environ()
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

}
