package deepcopy

import "reflect"

type PartialReplicator interface {
	Copy(dst, src interface{}) (copied bool)
}

func NewPartialReplicator(fieldsSelected ...string) PartialReplicator {
	return &partialReplicator{
		hierarchy: fieldsToTree(fieldsSelected),
	}
}

type partialReplicator struct {
	hierarchy tree
}

func (r partialReplicator) Copy(dst, src interface{}) (copied bool) {
	if src == nil {
		return
	}

	if dst == nil {
		panic("the destination must not be nil")
	}

	if reflect.TypeOf(src) != reflect.TypeOf(dst) {
		panic("both src and dst must have the same type")
	}

	mimic, copied := inspectObject(reflect.ValueOf(src), &r.hierarchy)
	if copied {
		reflect.ValueOf(dst).Elem().Set(mimic.Elem())
	}
	return
}
