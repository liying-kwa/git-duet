package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/pborman/getopt"
)

const preCommit = "pre-commit"
const prepareCommitMsg = "prepare-commit-msg"
const preCommitHook = `#!/usr/bin/env bash
exec git duet-pre-commit "$@"
`
const prepareCommitMsgHook = `#!/usr/bin/env bash
exec git duet-prepare-commit-msg "$1"
`

func main() {
	var (
		quiet = getopt.BoolLong("quiet", 'q', "Silence output")
		help  = getopt.BoolLong("help", 'h', "Help")
	)

	getopt.Parse()
	getopt.SetParameters(fmt.Sprintf("{ %s | %s }", preCommit, prepareCommitMsg))

	if *help {
		getopt.Usage()
		os.Exit(0)
	}

	args := getopt.Args()
	if len(args) != 1 {
		getopt.Usage()
		os.Exit(1)
	}
	hookFileName := args[0]

	var hook string
	if hookFileName == preCommit {
		hook = preCommitHook
	} else if hookFileName == prepareCommitMsg {
		hook = prepareCommitMsgHook
	} else {
		getopt.Usage()
		os.Exit(1)
	}

	output := new(bytes.Buffer)
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Stdout = output
	if err := cmd.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	hookPath := path.Join(strings.TrimSpace(output.String()), ".git", "hooks", hookFileName)

	hookFile, err := os.OpenFile(hookPath, os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer hookFile.Close()

	b, err := ioutil.ReadAll(hookFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	contents := strings.TrimSpace(string(b))
	if contents != "" {
		if hook == preCommitHook && contents != strings.TrimSpace(preCommitHook) ||
			hook == prepareCommitMsgHook && contents != strings.TrimSpace(prepareCommitMsgHook) {
			fmt.Printf("can't install hook: file %s already exists\n", hookPath)
			os.Exit(1)
		}
		os.Exit(0) // hook file with the desired content already exists
	}

	if _, err = hookFile.WriteString(hook); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if !*quiet {
		fmt.Printf("git-duet-install-hook: Installed hook to %s\n", hookPath)
	}
}
