package enum

import "fmt"

// Enum represents an enum.
type Enum interface {
	~int8 | ~int16 | ~int32 | ~int64 | ~int | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uint
	Descriptor() Descriptor
}

// Internals represents the blanket implementations derived from the provided
// values of an enum.
type Internals[T interface{ Descriptor() Descriptor }] struct {
	MarshalText   func(T) ([]byte, error)
	UnmarshalText func(*T, []byte) error
	Descriptor    Descriptor
}

// New returns the Internals for an enum of type T with the given values.
func New[T Enum](values []Value[T]) Internals[T] {
	toBytes := map[T][]byte{}
	fromString := map[string]T{}
	for _, v := range values {
		toBytes[v.Value] = []byte(v.String)
		fromString[v.String] = v.Value
	}

	marshalText := func(v T) ([]byte, error) {
		if b, ok := toBytes[v]; ok {
			return b, nil
		}
		return nil, fmt.Errorf("unknown enum value: %##v", v)
	}

	unmarshalText := func(p *T, data []byte) error {
		if v, ok := fromString[string(data)]; ok {
			*p = v
			return nil
		}
		return fmt.Errorf("unknown enum string representation: %q", string(data))
	}

	descriptor := &internalDescriptor[T]{values}

	return Internals[T]{
		MarshalText:   marshalText,
		UnmarshalText: unmarshalText,
		Descriptor:    descriptor,
	}
}

// Value represents a single value of an enum.
type Value[T Enum] struct {
	Value  T
	String string
}

// Number implements ValueDescriptor.
func (v Value[T]) Number() uint64 {
	return uint64(v.Value)
}

// Descriptor can be used for reflection of enums.
type Descriptor interface {
	Values() Values
}

// Values represents the values of an enum.
type Values interface {
	Len() int
	Get(index int) ValueDescriptor
}

// ValueDescriptor represents a single value of an enum abstracted.
type ValueDescriptor interface {
	Number() uint64
}

type internalDescriptor[T Enum] struct {
	values []Value[T]
}

func (d *internalDescriptor[T]) Values() Values {
	return d
}

func (d *internalDescriptor[T]) Get(index int) ValueDescriptor {
	return d.values[index]
}

func (d *internalDescriptor[T]) Len() int {
	return len(d.values)
}
