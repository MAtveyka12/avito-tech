package teams

import "errors"

var (
	ErrTeamNotFound   = errors.New("team not found")
	ErrTeamExists     = errors.New("team already exists")
)

