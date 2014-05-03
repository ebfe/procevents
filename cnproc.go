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
