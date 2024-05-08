package token

import (
	"github.com/google/uuid"
	"time"
)

type Maker interface {
	// CreateToken creates a new token for a specific username and duration.
	CreateToken(username string, userId uuid.UUID, duration time.Duration) (string, *Payload, error)

	// VerifyToken checks if the token is valid or not.
	VerifyToken(token string) (*Payload, error)
}
