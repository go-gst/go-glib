package convert

import (
	"fmt"
	"strings"

	"github.com/go-gst/go-glib/gir/girgen/file"
	"github.com/go-gst/go-glib/gir/girgen/typesystem"
)

// UnimplementedConverter is used for any conversion that is not implemented.
//
// the conversion will panic, which means that this will generate fine, but crash at runtime
type UnimplementedConverter struct {
	Param  *typesystem.Param
	Reason string
}

// Metadata implements Converter.
func (n *UnimplementedConverter) Metadata() string {
	var contextParts []string

	contextParts = append(contextParts,
		n.Param.Direction,
		fmt.Sprintf("transfer: %s", n.Param.TransferOwnership),
		fmt.Sprintf("C Pointers: %d", n.Param.CTypePointers),
		fmt.Sprintf("Name: %s", n.Param.Type.Type.GIRName()),
	)

	if n.Param.Scope != typesystem.CallbackParamScopeCall {
		contextParts = append(contextParts, fmt.Sprintf("scope: %s", n.Param.Scope))
	}

	if n.Param.Optional {
		contextParts = append(contextParts, "optional")
	}

	if n.Param.Nullable {
		contextParts = append(contextParts, "nullable")
	}

	if n.Param.CallerAllocates {
		contextParts = append(contextParts, "caller-allocates")
	}

	if n.Param.Closure != nil {
		contextParts = append(contextParts, fmt.Sprintf("closure: %s", n.Param.Closure.CName))
	}

	if n.Param.Destroy != nil {
		contextParts = append(contextParts, fmt.Sprintf("destroy: %s", n.Param.Destroy.CName))
	}

	if arr, ok := n.Param.Type.Type.(*typesystem.Array); ok {
		var arrParts []string

		if arr.Inner.Type != nil {
			arrParts = append(arrParts, fmt.Sprintf("inner %s (%T)", arr.Inner.Type.CType(arr.InnerPointers), arr.Inner.Type))
		} else {
			arrParts = append(arrParts, "inner unknown")
		}

		if arr.FixedSize > 0 {
			arrParts = append(arrParts, fmt.Sprintf("fixed-size: %d", arr.FixedSize))
		}

		if arr.ZeroTerminated {
			arrParts = append(arrParts, "zero-terminated")
		}

		if arr.Length != nil {
			arrParts = append(arrParts, fmt.Sprintf("length-by: %s", arr.Length.CName))
		}

		contextParts = append(contextParts, fmt.Sprintf("array (%s)", strings.Join(arrParts, ", ")))
	}

	return strings.Join(contextParts, ", ")
}

// Convert implements Converter.
func (n *UnimplementedConverter) Convert(w file.File) {
	// use the params to prevent not used errors:
	fmt.Fprintf(w.Go(), "_ = %s\n", n.Param.GoName)
	fmt.Fprintf(w.Go(), "_ = %s\n", n.Param.CName)

	// also use referenced implicit params to prevent not used errors:
	if n.Param.Closure != nil {
		fmt.Fprintf(w.Go(), "_ = %s\n", n.Param.Closure.CName)
	}
	if n.Param.Destroy != nil {
		fmt.Fprintf(w.Go(), "_ = %s\n", n.Param.Destroy.CName)
	}
	if arr, ok := n.Param.Type.Type.(*typesystem.Array); ok {
		if arr.Length != nil {
			// doesn't need goname because the go name is never declared
			fmt.Fprintf(w.Go(), "_ = %s\n", arr.Length.CName)
		}
	}

	reason := n.Reason

	if reason == "" {
		reason = "unknown reason"
	}

	fmt.Fprintf(w.Go(), "panic(\"unimplemented conversion of %s (%s) because of %s\")\n", n.Param.GoType(), n.Param.CType(), reason)
}

var _ Converter = (*UnimplementedConverter)(nil)
