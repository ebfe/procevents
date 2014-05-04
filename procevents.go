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

type Event interface {
	What() uint32
	Cpu() uint32
	Timestamp() uint64
	Pid() int32
	Tgid() int32
}

type Header struct {
	what      uint32
	cpu       uint32
	timestamp uint64
	pid       int32
	tgid      int32
}

func (h Header) What() uint32      { return h.what }
func (h Header) Cpu() uint32       { return h.cpu }
func (h Header) Timestamp() uint64 { return h.timestamp }
func (h Header) Pid() int32        { return h.pid }
func (h Header) Tgid() int32       { return h.tgid }

type Fork struct {
	Header
	ChildPid  int32
	ChildTgid int32
}

type Exec struct {
	Header
}

type Uid struct {
	Header
	Ruid uint32
	Euid uint32
}

type Gid struct {
	Header
	Rgid uint32
	Egid uint32
}

type Sid struct {
	Header
}

type Ptrace struct {
	Header
	TracerPid  int32
	TracerTgid int32
}

type Comm struct {
	Header
	Comm [16]byte
}

type Coredump struct {
	Header
}

type Exit struct {
	Header
	Code   uint32
	Signal uint32
}
