package gui

// Opt is an optional value. The zero value means "not set";
// use Some(v) to set explicitly.
type Opt[T any] struct {
	val T
	set bool
}

// Some returns an Opt with v explicitly set.
func Some[T any](v T) Opt[T] { return Opt[T]{val: v, set: true} }

// Get returns val if set, otherwise def.
func (o Opt[T]) Get(def T) T {
	if o.set {
		return o.val
	}
	return def
}

// IsSet reports whether a value was explicitly set.
func (o Opt[T]) IsSet() bool { return o.set }

// Value returns the value and whether it was set.
func (o Opt[T]) Value() (T, bool) { return o.val, o.set }
