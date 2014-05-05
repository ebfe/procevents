package procevents_test

import (
	"bytes"
	"fmt"
	"io/ioutil"

	"github.com/ebfe/procevents"
)

func getCommandLine(pid int32) []string {
	buf, err := ioutil.ReadFile(fmt.Sprintf("/proc/%d/cmdline", pid))
	if err != nil {
		return []string{"<unknown>"}
	}
	var cmdline []string
	tok := bytes.Split(buf, []byte{0})
	for _, t := range tok {
		cmdline = append(cmdline, string(t))
	}
	return cmdline
}

func Example() {
	conn, err := procevents.Dial()
	if err != nil {
		fmt.Printf("err: %s\n", err)
		return
	}
	defer conn.Close()

	for {
		events, err := conn.Read()
		if err != nil {
			fmt.Printf("err: %s\n", err)
			return
		}
		for _, ev := range events {
			switch ev := ev.(type) {
			case procevents.Exec:
				fmt.Printf("exec: %d %s\n", ev.Pid(), getCommandLine(ev.Pid()))
			case procevents.Exit:
				fmt.Printf("exit: %d (%d)\n", ev.Pid(), ev.Code)
			default:
				/* ignore */
			}
		}
	}
}
