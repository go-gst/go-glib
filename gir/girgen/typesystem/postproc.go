package typesystem

import (
	"fmt"
)

type PostProcessor func(*Registry) error

func MarkAsManuallyExtended(namespace string, typ string) PostProcessor {
	return func(r *Registry) error {
		ns := r.FindNamespaceByName(namespace)
		if ns == nil {
			return fmt.Errorf("could not find namespace %s", namespace)
		}

		t := ns.FindLocalTypeByGIRName(typ)

		if t == nil {
			return fmt.Errorf("could not find type %s in namespace %s", typ, namespace)
		}

		switch t := t.(type) {
		case *Class:
			t.ManuallyExtended = true
		case *Interface:
			t.ManuallyExtended = true
		default:
			return fmt.Errorf("type %s is not a class or interface, but instead %T", typ, t)
		}

		return nil
	}

}
