// Created by cgo -godefs - DO NOT EDIT
// cgo -godefs types_linux.go

package procevents

import (
	"fmt"
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
	sa := &syscall.SockaddrNetlink{Family: syscall.AF_NETLINK, Groups: groups, Pid: 0}
	if err := syscall.Bind(sock, sa); err != nil {
		syscall.Close(sock)
		return fmt.Errorf("procevents: bind: %s", err)
	}
	return nil
}

func cnProcMcastOp(sock int, op procCnMcastOp) error {
	s, err := syscall.Getsockname(sock)
	if err != nil {
		return err
	}
	portid := s.(*syscall.SockaddrNetlink).Pid

	msg := procEventMcastMsg{}
	msg.nlhdr.nlmsg_type = syscall.NLMSG_DONE
	msg.nlhdr.nlmsg_flags = 0
	msg.nlhdr.nlmsg_seq = 0
	msg.nlhdr.nlmsg_pid = C.__u32(portid)

	msg.cnhdr.id.idx = cnIdxProc
	msg.cnhdr.id.val = cnValProc
	msg.cnhdr.seq = 0
	msg.cnhdr.ack = 0
	msg.cnhdr.len = C.__u16(C.sizeof_enum_proc_cn_mcast_op)
	msg.cnhdr.flags = 0

	msg.op = uint32(op)

	msg.nlhdr.nlmsg_len = C.__u32(C.sizeof_struct_nlmsghdr + C.sizeof_struct_cn_msg + C.sizeof_enum_proc_cn_mcast_op)

	raw := C.GoBytes(unsafe.Pointer(&msg), C.int(msg.nlhdr.nlmsg_len))
	_, err = syscall.Write(sock, raw)
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

func parseProcEvent(msg *syscall.NetlinkMessage) (Event, error) {
	cnmsg := (*C.struct_cn_msg)(unsafe.Pointer(&msg.Data[0]))
	pe := (*C.struct_proc_event)(unsafe.Pointer(&cnmsg.data))

	switch pe.what {
	case C.PROC_EVENT_NONE:
		return *(*None)(unsafe.Pointer(pe)), nil
	case C.PROC_EVENT_FORK:
		return *(*Fork)(unsafe.Pointer(pe)), nil
	case C.PROC_EVENT_EXEC:
		return *(*Exec)(unsafe.Pointer(pe)), nil
	case C.PROC_EVENT_UID:
		return *(*Uid)(unsafe.Pointer(pe)), nil
	case C.PROC_EVENT_GID:
		return *(*Gid)(unsafe.Pointer(pe)), nil
	case C.PROC_EVENT_SID:
		return *(*Sid)(unsafe.Pointer(pe)), nil
	case C.PROC_EVENT_PTRACE:
		return *(*Ptrace)(unsafe.Pointer(pe)), nil
	case C.PROC_EVENT_COMM:
		return *(*Comm)(unsafe.Pointer(pe)), nil
	case C.PROC_EVENT_COREDUMP:
		return *(*Coredump)(unsafe.Pointer(pe)), nil
	case C.PROC_EVENT_EXIT:
		return *(*Exit)(unsafe.Pointer(pe)), nil
	default:
		return nil, fmt.Errorf("procevents: unknown event type (%x)", pe.what)
	}
}
