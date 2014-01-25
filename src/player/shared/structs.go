package shared

import (
	"time"
)

type AttributedString struct {
	Content string
	Style   int //0 -normal, 1 -italic, 2 -bold, 3 italic and bold
	Color   uint
}

type SubItem struct {
	From, To time.Duration
	Content  []AttributedString

	PositionType int
	Position

	SubItemExtra
}

type Sub struct {
	Movie   string
	Name    string
	Offset  time.Duration
	Content string
	Type    string
}

type Playing struct {
	Movie       string
	LastPos     time.Duration
	SoundStream int
	Sub1        string
	Sub2        string
}

func (item *SubItem) IsInDefaultPosition() bool {
	return item.PositionType == 2 && item.X < 0 && item.Y < 0
}

type SubItemExtra struct {
	Id     int
	Handle uintptr
}
type SubItemArg struct {
	SubItem
	Result chan SubItemExtra
}
type Position struct {
	X, Y float64
}

type PlayProgressInfo struct {
	Left     string
	Right    string
	Percent  float64
	Percent2 float64
}