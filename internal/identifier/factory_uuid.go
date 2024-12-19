package identifier

import "github.com/google/uuid"

// FactoryUUID is the concrete implementation of the [Factory] interface.
//
// It uses [github.com/google/uuid] library to perform internal operations.
type FactoryUUID struct{}

var _ Factory = (*FactoryUUID)(nil)

func (f FactoryUUID) NewID() (string, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return "", err
	}
	return id.String(), nil
}
