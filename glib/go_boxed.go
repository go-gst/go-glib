package glib

/*
#include "glib.go.h"

extern gpointer goCopyGoPointer (gpointer handle);
extern void     goFreeGoPointer (gpointer handle);

G_DEFINE_BOXED_TYPE(GlibGoArbitraryData, glib_go_arbitrary_data,
                    glib_go_arbitrary_data_copy,
                    glib_go_arbitrary_data_free)

static void glib_go_arbitrary_data_free (GlibGoArbitraryData * d)
{
	goFreeGoPointer(d->data);

	g_free(d);
}

static GlibGoArbitraryData *glib_go_arbitrary_data_copy (GlibGoArbitraryData * orig)
{
    GlibGoArbitraryData *copy;

    if (!orig)
        return NULL;

    copy = g_new0 (GlibGoArbitraryData, 1);
    copy->data = goCopyGoPointer(orig->data);

    return copy;
}

static GlibGoArbitraryData *glib_go_arbitrary_data_new(gpointer data)
{
    GlibGoArbitraryData *gdata;

    gdata = g_new0(GlibGoArbitraryData, 1);
    gdata->data = data;

    return gdata;
}
*/
import "C"

import (
	"unsafe"

	gopointer "github.com/go-gst/go-pointer"
)

var TYPE_ARBITRARY_DATA Type = Type(C.GLIB_GO_TYPE_ARBITRARY_DATA)

func init() {
	tm := []TypeMarshaler{
		{TYPE_ARBITRARY_DATA, marshalArbitraryValue},
	}

	RegisterGValueMarshalers(tm)
}

// ArbitraryValue allows to pass any value into a glib property or signal.
//
// it is helpful when you want to pass a go value into a custom element that is
// also defined in go.
type ArbitraryValue struct {
	Data any
}

var _ ValueTransformer = ArbitraryValue{}

func (v ArbitraryValue) ToGValue() (*Value, error) {
	handle := gopointer.Save(v.Data)

	gv, err := ValueInit(TYPE_ARBITRARY_DATA)

	if err != nil {
		return nil, err
	}

	cv := C.glib_go_arbitrary_data_new(C.gpointer(handle))

	// TakeBoxed lets the GValue take ownership of the boxed struct
	// the gvalue will free the data when it is freed
	gv.TakeBoxed(unsafe.Pointer(cv))

	return gv, nil
}

func marshalArbitraryValue(p unsafe.Pointer) (interface{}, error) {
	cp := C.g_value_get_boxed((*C.GValue)(p))

	cv := (*C.GlibGoArbitraryData)(cp)

	data := gopointer.Restore(unsafe.Pointer(cv.data))

	arb := ArbitraryValue{
		Data: data,
	}

	return arb, nil
}
