package subtitle

import (
	// "io/ioutil"
	"log"
	. "player/clock"
	. "player/shared"
	"player/subtitle/ass"
	"player/subtitle/srt"
	// "runtime"
	"time"
)

type durationArg struct {
	d   time.Duration
	res chan time.Duration
}
type subTimeArg struct {
	t      time.Duration
	offset int
	res    chan time.Duration
}

type Subtitle struct {
	r     SubRender
	items []*SubItem

	c *Clock

	offset time.Duration
	quit   chan bool

	ChanSeek chan time.Duration

	ChanOffset chan durationArg

	chanGetSubTime chan subTimeArg

	// ChanPause        chan time.Duration
	// ChanPauseSeeking chan time.Duration

	chanStop chan bool

	Name string

	IsMainOrSecondSub bool

	// DisplayingMap map[int]uintptr
}

type displayingItem struct {
	pos    int
	handle uintptr
}

func (s *Subtitle) printSub(pos int) {
	item := s.items[pos]
	content := ""
	if len(item.Content) > 0 {
		content = item.Content[0].Content
	}

	log.Printf("%d:%s", pos, content)
}

func (s *Subtitle) Play() {
	for i, _ := range s.items {
		s.items[i].Id = i
	}

	chRes := make(chan SubItemExtra)
	for {
		select {
		case arg := <-chRes:
			println("res from show:", arg.Id, arg.Handle)
			s.items[arg.Id].Handle = arg.Handle
			break
		case arg := <-s.ChanOffset:
			s.offset += arg.d
			arg.res <- s.offset
			break
		case arg := <-s.chanGetSubTime:
			pos, _ := s.findPos(arg.t)

			s.printSub(pos)
			if arg.offset < 0 && s.calcFrom(pos) > arg.t {
				pos--
			}

			s.printSub(pos)

			// pos += arg.offset
			// println("pos:", pos)
			for s.checkPos(pos, arg.t) {
				pos += arg.offset
			}

			s.printSub(pos)

			for {
				if pos < 0 {
					pos = 0
					break
				}
				if pos >= len(s.items) {
					pos = len(s.items) - 1
					break
				}

				item := s.items[pos]
				if !item.IsInDefaultPosition() {
					pos += arg.offset

				} else {
					break
				}
			}

			arg.res <- s.calcFrom(pos)
			break
		case t := <-s.ChanSeek:
			s.render(t, chRes)
			break
		case <-s.chanStop:
			for _, item := range s.items {
				if item.Handle != 0 {
					s.hideSubItem(*item)
					item.Handle = 0
				}
			}
			break
		}
	}
}

func (s *Subtitle) render(t time.Duration, chRes chan SubItemExtra) {
	for i, item := range s.items {
		if !s.checkPos(i, t) {
			if item.Handle != 0 && item.Handle != 1 {
				// println(t.String(), "hide sub: ", item.Id, item.Content[0].Content)
				s.hideSubItem(*item)
				item.Handle = 0
			}
		}
	}

	for i, item := range s.items {
		if s.checkPos(i, t) {
			if item.Handle == 0 {
				// println(t.String(), "show sub: ", item.Id, item.Content[0].Content)
				s.showSubitem(*item, chRes)
				item.Handle = 1
			}
		}
	}
}

func (s *Subtitle) Stop() {
	s.chanStop <- true
}

func (s *Subtitle) calcFromTo(pos int) (time.Duration, time.Duration) {
	item := s.items[pos]
	return item.From + s.offset, item.To + s.offset
}

func (s *Subtitle) calcFrom(pos int) time.Duration {
	item := s.items[pos]
	return item.From + s.offset
}

func (s *Subtitle) checkPos(pos int, t time.Duration) bool {
	if pos >= len(s.items) || pos < 0 {
		return false
	}

	from, to := s.calcFromTo(pos)
	// println("check pos:", pos, t.String(), from.String(), to.String())
	return t >= from && t < to
}

func (s *Subtitle) showSubitem(item SubItem, chRes chan SubItemExtra) {
	if !s.IsMainOrSecondSub && item.PositionType != 10 {
		if (item.PositionType != 2) || (item.X >= 0) || (item.Y >= 0) {
			return
		} else {
			item.PositionType = 10
		}
	}
	arg := SubItemArg{item, chRes}
	go s.r.SendShowText(arg)
}
func (s *Subtitle) hideSubItem(item SubItem) {
	if item.Handle != 0 && item.Handle != 1 {
		go s.r.SendHideText(SubItemArg{item, nil})
	}
}

func (s *Subtitle) findPos(t time.Duration) (int, *SubItem) {
	for i := 0; i < len(s.items); i++ {
		from, to := s.calcFromTo(i)
		if t < to {
			if t >= from {
				return i, s.items[i]
			} else {
				return i, nil
			}
		}
	}

	return 1 << 31, nil
}

func NewSubtitle(file string, r SubRender, c *Clock, width, height float64) *Subtitle {
	var err error
	sub := GetSubtitle(file)
	if sub == nil {
		log.Print("can't find subtitle: ", file)
		return nil
	}
	s := &Subtitle{}
	s.c = c
	s.offset = sub.Offset
	s.ChanSeek = make(chan time.Duration)
	s.ChanOffset = make(chan durationArg)
	s.chanStop = make(chan bool)
	s.chanGetSubTime = make(chan subTimeArg)
	s.Name = file
	s.IsMainOrSecondSub = true

	if sub.Type == "ass" {
		s.items, err = ass.Parse(sub.Content)
	} else {
		s.items, err = srt.Parse(sub.Content, width, height)
	}

	if err != nil {
		log.Print(err)
		return nil
	}

	s.quit = make(chan bool)
	s.r = r

	log.Print("sub items:", len(s.items))
	return s
}

func (s *Subtitle) Seek(t time.Duration) {
	// println("subtitle seek:", t.String())
	s.ChanSeek <- t
	// println("subtitle seeked:", t.String())
}

func (s *Subtitle) AddOffset(d time.Duration) time.Duration {
	res := make(chan time.Duration)
	s.ChanOffset <- durationArg{d, res}
	return <-res
}

func (s *Subtitle) GetSubtime(t time.Duration, offset int) time.Duration {
	res := make(chan time.Duration)
	arg := subTimeArg{t, offset, res}
	s.chanGetSubTime <- arg
	return <-arg.res
}