package utils

// Ptr returns a pointer to the given value.
func Ptr[T any](in T) *T {
	return &in
}

// Deref returns the dereferenced value of the given pointer, or the zero value of the type if the pointer is nil.
func Deref[T any](in *T) T {
	return DerefOr(in, ZeroValue[T]())
}

// DerefOr returns the dereferenced value of the given pointer, or the given value if the pointer is nil.
func DerefOr[T any](in *T, valueIfNil T) T {
	if in == nil {
		return valueIfNil
	}
	return *in
}

// ZeroValue returns the zero value of the given type.
func ZeroValue[T any]() (out T) {
	return
}
