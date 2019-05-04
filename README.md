#DeepCopy

`deepcopy` is a golang library that provides an ability to deeply duplicate specific fields of an object.

## Examples

Copy fields.

```go
import "github.com/kitt1987/deepcopy"

func main() {
  // copied will be false if no specific fields found in src.
  copied := deepcopy.Partial(&dst, &src, "FieldA", "FieldB.SubFieldC")
}
```

Copy fields when they are different in the source from in destination.

```go
import "github.com/kitt1987/deepcopy"

func main() {
  // copied will be false if no specific fields found in src.
  copied := deepcopy.OnChange(&dst, &src, "FieldA", "FieldB.SubFieldC")
}
```

## Thanks

This library uses [github.com/mohae/deepcopy](github.com/mohae/deepcopy) to copy arbitrary fields.