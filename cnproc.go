// Created by cgo -godefs - DO NOT EDIT
// cgo -godefs types_linux.go

package procevents

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

/*
#include <linux/netlink.h>
#include <linux/connector.h>
#include <linux/cn_proc.h>
*/
import "C"

const (
	cnIdxProc = C.CN_IDX_PROC
	cnValProc = C.CN_VAL_PROC
)

type procCnMcastOp C.enum_proc_cn_mcast_op

const (
	procCnMcastListen = C.PROC_CN_MCAST_LISTEN
	procCnMcastIgnore = C.PROC_CN_MCAST_IGNORE
)

type procEventMcastMsg struct {
	nlhdr C.struct_nlmsghdr
	cnhdr C.struct_cn_msg
	op    C.enum_proc_cn_mcast_op
}

func cnSocket() (int, error) {
	sock, err := syscall.Socket(syscall.AF_NETLINK, syscall.SOCK_DGRAM, syscall.NETLINK_CONNECTOR)
	if err != nil {
		return -1, fmt.Errorf("procevents: socket: %s", err)
	}
	return sock, nil
}

func cnBind(sock int, groups uint32) error {
	sa := syscall.SockaddrNetlink{Family: syscall.AF_NETLINK, Groups: groups, Pid: uint32(os.Getpid())}
	if err := syscall.Bind(sock, &sa); err != nil {
		syscall.Close(sock)
		return fmt.Errorf("procevents: bind: %s", err)
	}
	return nil
}

func cnProcMcastOp(sock int, op procCnMcastOp) error {
	msg := procEventMcastMsg{}
	msg.nlhdr.nlmsg_type = syscall.NLMSG_DONE
	msg.nlhdr.nlmsg_flags = 0
	msg.nlhdr.nlmsg_seq = 0
	msg.nlhdr.nlmsg_pid = C.__u32(os.Getpid())

	msg.cnhdr.id.idx = cnIdxProc
	msg.cnhdr.id.val = cnValProc
	msg.cnhdr.seq = 0
	msg.cnhdr.ack = 0
	msg.cnhdr.len = C.__u16(unsafe.Sizeof(msg.op))
	msg.cnhdr.flags = 0

	msg.op = uint32(op)

	msg.nlhdr.nlmsg_len = C.__u32(unsafe.Sizeof(msg.cnhdr)) + C.__u32(msg.cnhdr.len)

	raw := C.GoBytes(unsafe.Pointer(&msg), C.int(unsafe.Sizeof(msg)))
	_, err := syscall.Write(sock, raw)
	if err != nil {
		syscall.Close(sock)
		return fmt.Errorf("procevents: write: %s", err)
	}
	return nil
}

func cnProcMcastListen(sock int) error {
	return cnProcMcastOp(sock, procCnMcastListen)
}

func cnProcMcastIgnore(sock int) error {
	return cnProcMcastOp(sock, procCnMcastIgnore)
}

func parseProcEvent(msg *syscall.NetlinkMessage) (interface{}, error) {
	cnmsg := (*C.struct_cn_msg)(unsafe.Pointer(&msg.Data[0]))
	pe := (*C.struct_proc_event)(unsafe.Pointer(&cnmsg.data))

	ev := Event{
		What:      uint32(pe.what),
		Cpu:       uint32(pe.cpu),
		Timestamp: uint64(pe.timestamp_ns),
	}

	switch pe.what {
	case C.PROC_EVENT_FORK:
		return Fork(ev), nil
	case C.PROC_EVENT_EXEC:
		return Exec(ev), nil
	case C.PROC_EVENT_UID:
		return Uid(ev), nil
	case C.PROC_EVENT_GID:
		return Gid(ev), nil
	case C.PROC_EVENT_SID:
		return Sid(ev), nil
	case C.PROC_EVENT_PTRACE:
		return Ptrace(ev), nil
	case C.PROC_EVENT_COMM:
		return Comm(ev), nil
	case C.PROC_EVENT_COREDUMP:
		return Coredump(ev), nil
	case C.PROC_EVENT_EXIT:
		return Exit(ev), nil
	default:
		return nil, fmt.Errorf("procevents: unknown event type (%x)", pe.what)
	}
}
