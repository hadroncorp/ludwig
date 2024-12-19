package identifier

// Factory is a system component in charge of generating unique identifiers.
type Factory interface {
	// NewID generates a unique identifier.
	NewID() (string, error)
}
