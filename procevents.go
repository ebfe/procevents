package procevents

import (
	"syscall"
)

type Conn struct {
	sock  int
	evbuf []interface{}
}

func Dial() (*Conn, error) {

	sock, err := cnSocket()
	if err != nil {
		return nil, err
	}

	if err := cnBind(sock, cnIdxProc); err != nil {
		return nil, err
	}

	if err := cnProcMcastListen(sock); err != nil {
		return nil, err
	}

	return &Conn{sock: sock}, nil
}

func (c *Conn) Read() (interface{}, error) {

	for len(c.evbuf) == 0 {
		buf := make([]byte, 1<<16)
		n, _, err := syscall.Recvfrom(c.sock, buf, 0)
		if err != nil {
			return nil, err
		}
		if n < syscall.NLMSG_HDRLEN {
			continue
		}

		msgs, err := syscall.ParseNetlinkMessage(buf[:n])
		if err != nil {
			return nil, err
		}

		for i := range msgs {
			ev, err := parseProcEvent(&msgs[i])
			if err != nil {
				return nil, err
			}
			c.evbuf = append(c.evbuf, ev)
		}
	}

	ev := c.evbuf[0]
	c.evbuf = c.evbuf[1:]

	return ev, nil
}

func (c *Conn) Close() error {
	cnProcMcastIgnore(c.sock)
	return syscall.Close(c.sock)
}

type Event struct {
	What      uint32
	Cpu       uint32
	Timestamp uint64
	Pid       int
	Tgid      int
}

type Fork Event
type Exec Event
type Uid Event
type Gid Event
type Sid Event
type Ptrace Event
type Comm Event
type Coredump Event
type Exit Event
