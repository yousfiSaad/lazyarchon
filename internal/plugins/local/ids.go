package local

import "github.com/google/uuid"

// newID returns a random UUIDv4 string used for project and task IDs.
func newID() string {
	return uuid.NewString()
}
