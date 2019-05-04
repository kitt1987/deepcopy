package deepcopy

import (
	"fmt"
	"reflect"
	"strings"
)

func Partial(dst, src interface{}, fieldsSelected ...string) (copied bool) {
	return NewPartialReplicator(fieldsSelected...).Copy(dst, src)
}

func OnChange(dst, src interface{}, fieldsSelected ...string) (copied bool) {
	hierarchy := fieldsToTree(fieldsSelected)
	_, copied = copyPieceChanges(reflect.ValueOf(dst), reflect.ValueOf(src), &hierarchy)
	return copied
}

type tree struct {
	branches map[string]tree
	layer    int
}

func (t tree) FindBranch(branchValue string) (branch *tree) {
	if b, found := t.branches[branchValue]; found {
		branch = &b
	}

	return
}

func (t *tree) AddBranch(branchValue string) (branch *tree) {
	b := newTree(t.layer + 1)
	t.branches[branchValue] = b
	return &b
}

func newTree(layerId int) tree {
	return tree{
		branches: make(map[string]tree),
		layer:    layerId,
	}
}

func fieldsToTree(fields []string) (t tree) {
	t = newTree(0)
	for i, field := range fields {
		if len(field) == 0 {
			panic(fmt.Sprintf("the %dth field in %#v is empty", i, fields))
		}

		hierarchies := strings.Split(field, ".")
		cur := &t
		for h, hierarchy := range hierarchies {
			if len(hierarchy) == 0 {
				panic(fmt.Sprintf("field %s contains a blank path at index %d", field, h))
			}

			if branch := cur.FindBranch(hierarchy); branch != nil {
				cur = branch
			} else {
				cur = cur.AddBranch(hierarchy)
			}
		}
	}

	return
}

func inspectObject(in reflect.Value, hierarchy *tree) (mimic reflect.Value, copied bool) {
	if in.Kind() != reflect.Ptr && in.Kind() != reflect.Slice && in.Kind() != reflect.Struct {
		panic(fmt.Sprintf("the object should be a pointer, structure or slice but %s", in.Kind()))
	}

	mimic = reflect.New(in.Type()).Elem()
	out := mimic

	if in.Kind() == reflect.Ptr {
		in = in.Elem()
		if !out.Elem().IsValid() {
			out.Set(reflect.New(in.Type()))
		}

		out = out.Elem()
	}

	if in.Kind() == reflect.Slice {
		slice := reflect.Zero(in.Type())
		for j := 0; j < in.Len(); j++ {
			var elem reflect.Value
			elem, copied = inspectObject(in.Index(j), hierarchy)
			if elem.IsValid() {
				slice = reflect.Append(slice, elem)
			}
		}

		out.Set(slice)
		return
	}

	for value, branch := range hierarchy.branches {
		nextIn := in.FieldByName(value)
		nextOut := out.FieldByName(value)

		if len(branch.branches) == 0 {
			if !nextIn.IsValid() {
				continue
			}

			if nextIn.Kind() == reflect.Map && !nextIn.IsNil() ||
				reflect.Zero(nextIn.Type()).Interface() != nextIn.Interface() {
				copied = true
				copyRecursive(nextIn, nextOut)
			}
		} else {
			var v reflect.Value
			v, copied = inspectObject(nextIn, &branch)
			nextOut.Set(v)
		}
	}

	return
}

func copyPieceChanges(dst, src reflect.Value, hierarchy *tree) (mimic reflect.Value, copied bool) {
	if src.Kind() != reflect.Ptr && src.Kind() != reflect.Slice && src.Kind() != reflect.Struct {
		panic(fmt.Sprintf("the object should be a pointer, structure or slice but %s", src.Kind()))
	}

	if dst.IsValid() {
		mimic = dst
	} else {
		mimic = reflect.New(src.Type()).Elem()
	}

	out := mimic

	if src.Kind() == reflect.Ptr {
		src = src.Elem()
		if !out.Elem().IsValid() {
			out.Set(reflect.New(src.Type()))
		}

		out = out.Elem()
	}

	if src.Kind() == reflect.Slice {
		slice := reflect.Zero(src.Type())
		for j := 0; j < src.Len(); j++ {
			var elem reflect.Value
			elem, copied = copyPieceChanges(out.Index(j), src.Index(j), hierarchy)
			if elem.IsValid() {
				slice = reflect.Append(slice, elem)
			}
		}
		return
	}

	for value, branch := range hierarchy.branches {
		nextIn := src.FieldByName(value)
		nextOut := out.FieldByName(value)

		if len(branch.branches) == 0 {
			if !nextIn.IsValid() {
				continue
			}

			switch nextIn.Kind() {
			case reflect.Map,
				reflect.Struct,
				reflect.Slice:
				copied = !reflect.DeepEqual(nextIn.Interface(), nextOut.Interface())
			default:
				copied = nextIn.Interface() != nextOut.Interface()
			}

			if copied {
				copyRecursive(nextIn, nextOut)
			}
		} else {
			_, copied = copyPieceChanges(nextOut, nextIn, &branch)
		}
	}

	return
}
