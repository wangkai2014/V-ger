package subtitle

import (
	"io/ioutil"
	"log"
	. "player/clock"
	. "player/shared"
	"player/srt"
	"runtime"
	"time"
)

type Subtitle struct {
	r     SubRender
	items []*SubItem

	c *Clock

	offset time.Duration
	quit   chan bool

	ChanSeek chan time.Duration

	ChanOffset    chan time.Duration
	ChanOffsetRes chan time.Duration
}

func (s *Subtitle) Play() {
	var pos int

	for {
		select {
		case d := <-s.ChanOffset:
			s.offset += d
			s.ChanOffsetRes <- s.offset
			break
		case t := <-s.ChanSeek:
			close(s.quit)
			s.quit = make(chan bool)

			pos, _ = s.FindPos(t)
			runtime.Gosched()
			break
		case <-time.After(20 * time.Millisecond):
			if s.checkPos(pos, s.c.GetSeekTime()) {
				go func(pos int) {
					s.playOneItem(pos)
				}(pos)
				pos++
			}
			break
		}
	}
}

func (s *Subtitle) calcFromTo(i int) (time.Duration, time.Duration) {
	item := s.items[i]
	return item.From + s.offset, item.To + s.offset
}

func (s *Subtitle) checkPos(pos int, t time.Duration) bool {
	if pos >= len(s.items) {
		return false
	}

	from, to := s.calcFromTo(pos)
	// println("check pos:", pos, t.String(), from.String(), to.String())
	return t >= from && t < to
}

func (s *Subtitle) playOneItem(pos int) {
	_, to := s.calcFromTo(pos)
	tId := s.r.SendShowText(s.items[pos])
	s.c.WaitUtilWithQuit(to-20*time.Millisecond, s.quit)
	s.r.SendHideText(tId)
	println("play one item:", pos, s.items[pos].Content[0].Content)
}

func (s *Subtitle) FindPos(t time.Duration) (int, *SubItem) {
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

func NewSubtitle(file string, r SubRender, c *Clock) *Subtitle {
	var err error
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		log.Print(err)
		return nil
	}
	s := &Subtitle{}
	s.c = c
	s.ChanSeek = make(chan time.Duration)
	s.ChanOffset = make(chan time.Duration)
	s.ChanOffsetRes = make(chan time.Duration)

	s.items = srt.Parse(string(bytes))
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
	println("subtitle seek:", t.String())
	s.ChanSeek <- t
}

func (s *Subtitle) AddOffset(d time.Duration) time.Duration {
	s.ChanOffset <- d
	return <-s.ChanOffsetRes
}
