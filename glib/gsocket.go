package glib

import "C"
import (
	"errors"
	"syscall"
	"unsafe"
)

func SocketNew(domain, typ, proto int) (*Object, error) {
	fd, err := syscall.Socket(domain, typ, proto)
	if err != nil {
		return nil, err
	}
	return SocketNewFromFd(fd)
}

func SocketNewFromFd(fd int) (*Object, error) {
	var gerr *C.GError
	var socket *C.GSocket
	socket = C.g_socket_new_from_fd((C.gint)(fd), (**C.GError)(unsafe.Pointer(&gerr)))
	if gerr != nil {
		defer C.g_error_free(gerr)
		return nil, errors.New(C.GoString(gerr.message))
	}

	return Take(unsafe.Pointer(socket)), nil
}
