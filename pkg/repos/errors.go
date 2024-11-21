package repos

import "errors"

var (
	ErrNoSuchEntity  = errors.New("Entity does not exist")
	ErrEntityDeleted = errors.New("Entity has been deleted")
)
