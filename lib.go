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
	return OnChangeD(&traceNothing{}, dst, src, fieldsSelected...)
}

func OnChangeD(tracer Tracer, dst, src interface{}, fieldsSelected ...string) (copied bool) {
	hierarchy := fieldsToTree(fieldsSelected)
	_, copied = copyPieceChanges(reflect.ValueOf(dst), reflect.ValueOf(src), &hierarchy, &stackTracer{
		HierarchyStack: HierarchyStack(""),
		Tracer:         tracer,
	})
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

func (t tree) Trace(tr *stackTracer) {
	brs := ""
	for br := range t.branches {
		brs += "\n            +--" + tr.Prefix() + "." + br
	}

	tr.Println(brs)
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

func copyPieceChanges(dst, src reflect.Value, hierarchy *tree, tr *stackTracer) (mimic reflect.Value, copied bool) {
	if src.Kind() != reflect.Ptr && src.Kind() != reflect.Slice && src.Kind() != reflect.Struct {
		panic(fmt.Sprintf("the object should be a pointer, structure or slice but %s", src.Kind()))
	}

	hierarchy.Trace(tr)
	defer tr.Pop()
	tr.PrintfLn("Source: %#v", src)
	tr.PrintfLn("Destination: %#v", dst)

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
			tr.PrintfLn("Source field【%s】is a %s! Go through the %dth element!", tr.Prefix(),
				reflect.Slice.String(), j)
			elem, elemCopied := copyPieceChanges(out.Index(j), src.Index(j), hierarchy, tr)
			if elem.IsValid() {
				slice = reflect.Append(slice, elem)
			}
			copied = copied || elemCopied
		}
		return
	}

	for value, branch := range hierarchy.branches {
		tr.PrintfLn("=======================Detect branch【%s.%s】======================", tr.Prefix(), value)
		nextIn := src.FieldByName(value)
		nextOut := out.FieldByName(value)
		var elemCopied bool

		if len(branch.branches) == 0 {
			if !nextIn.IsValid() {
				tr.PrintfLn("Can't found field %s.%s in Source. Skip!", tr.Prefix(), value)
				tr.PrintfLn("========================End branch【%s.%s】========================", tr.Prefix(), value)
				continue
			}

			switch nextIn.Kind() {
			case reflect.Map,
				reflect.Struct,
				reflect.Slice:
				elemCopied = !reflect.DeepEqual(nextIn.Interface(), nextOut.Interface())
			default:
				elemCopied = nextIn.Interface() != nextOut.Interface()
			}

			tr.PrintfLn("Source: %#v", nextIn.Interface())
			tr.PrintfLn("Destination: %#v", nextOut.Interface())
			tr.PrintfLn("Source field【%s.%s】is a %s! Copied? %t", tr.Prefix(), value, nextIn.Kind().String(), copied)
			if elemCopied {
				copyRecursive(nextIn, nextOut)
			}
		} else {
			tr.PrintfLn("Source field【%s.%s】has branches. Go through!", tr.Prefix(), value)
			tr.Push(value)
			_, elemCopied = copyPieceChanges(nextOut, nextIn, &branch, tr)
			tr.PrintfLn("Source field【%s.%s】Copied? %t", tr.Prefix(), value, copied)
		}

		copied = copied || elemCopied
		tr.PrintfLn("========================End branch【%s.%s】========================", tr.Prefix(), value)
	}

	return
}
