package procevents 

import (
	"os"
	"syscall"
	"testing"
)

func TestMultipleSockets(t *testing.T) {
	var s [3]int

	if os.Getuid() != 0 {
		t.Skip("uid != 0")
		return
	}

	for i := range s {
		var err error
		s[i], err = cnSocket()
		if err != nil {
			t.Error(err)
			return
		}
		defer syscall.Close(s[i])

		if err := cnBind(s[i], cnIdxProc); err != nil {
			t.Error(err)
			return
		}
	}
}

