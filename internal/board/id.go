package board

import "github.com/oklog/ulid/v2"

func NewID() string {
	return "board_" + ulid.Make().String()
}
