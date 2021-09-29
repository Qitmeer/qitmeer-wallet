package util

type ExplicitString struct {
	Value         string
	explicitlySet bool
}

// ExplicitlySet returns whether the flag was explicitly set through the
// flags.Unmarshaler interface.
func (e *ExplicitString) ExplicitlySet() bool { return e.explicitlySet }

// MarshalFlag implements the flags.Marshaler interface.
func (e *ExplicitString) MarshalFlag() (string, error) { return e.Value, nil }

// UnmarshalFlag implements the flags.Unmarshaler interface.
func (e *ExplicitString) UnmarshalFlag(value string) error {
	e.Value = value
	e.explicitlySet = true
	return nil
}
