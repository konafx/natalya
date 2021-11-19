package repository

type GameStatus uint8

const (
	GameStatusStart = GameStatus(iota + 1)
	GameStatusPlaying
	GameStatusSuspend
	GamestatusGraceful
)
