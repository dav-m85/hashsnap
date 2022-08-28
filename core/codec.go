package core

import (
	"time"

	"github.com/google/uuid"
)

// const MaxInt = int(^uint(0) >> 1)

// Info header
type Info struct {
	RootPath  string
	CreatedAt time.Time
	Version   int
	Nonce     uuid.UUID
}
