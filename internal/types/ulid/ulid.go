package ulid

import (
	"crypto/rand"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/oklog/ulid/v2"
)

func NewUlid() uuid.UUID {
	t := time.Now()
	v, err := ulid.New(ulid.Timestamp(t), rand.Reader)
	if err != nil {
		fmt.Printf("Error generating ULID: %s", err.Error())
	}

	return uuid.UUID(v)
}

func ULIDFromString(str string) (uuid.UUID, error) {
	return uuid.Parse(str)
}
