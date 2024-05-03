package glib

// #include <gio/gio.h>
import "C"
import (
	"errors"
	"syscall"
	"unsafe"

	"golang.org/x/exp/constraints"
)

type Socket struct {
	*Object
}

func SocketNew(domain, typ, proto int) (*Socket, error) {
	fd, err := syscall.Socket(domain, typ, proto)
	if err != nil {
		return nil, err
	}
	return SocketNewFromFd(fd)
}

func SocketNewFromFd[T constraints.Integer](fd T) (*Socket, error) {
	var gerr *C.GError
	socket := C.g_socket_new_from_fd((C.gint)(fd), (**C.GError)(unsafe.Pointer(&gerr)))
	if gerr != nil {
		defer C.g_error_free(gerr)
		return nil, errors.New(C.GoString(gerr.message))
	}

	return &Socket{Take(unsafe.Pointer(socket))}, nil
}

func (s *Socket) ToGValue() (*Value, error) {
	val, err := ValueInit(TYPE_SOCKET)
	if err != nil {
		return nil, err
	}
	val.SetInstance(unsafe.Pointer(s.GObject))
	return val, nil
}
