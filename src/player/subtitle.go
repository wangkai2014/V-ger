package main

import (
	// "fmt"
	"io/ioutil"
	"log"
	. "player/clock"
	// "player/glfw"
	"player/gui"
	"player/srt"
	// "strings"
	"sync"
	"time"
)

type subtitle struct {
	sync.Locker

	w *gui.Window

	items []*srt.SubItem
	pos   int

	c *Clock

	offset time.Duration
	quit   chan bool
}

func NewSubtitle(file string, w *gui.Window) *subtitle {
	var err error
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		log.Print(err)
		return nil
	}
	s := &subtitle{}
	s.Locker = &sync.Mutex{}

	s.items = srt.Parse(string(bytes))
	if err != nil {
		log.Print(err)
		return nil
	}

	s.quit = make(chan bool)
	s.w = w

	log.Print("sub items:", len(s.items))
	w.FuncKeyDown = append(w.FuncKeyDown, func(keycode int) { //run in main thread, safe to operate ui elements
		switch keycode {
		case gui.KEY_MINUS:
			println("key minus pressed")
			s.addOffset(-1000 * time.Millisecond)
			break
		case gui.KEY_EQUAL:
			println("key equal pressed")
			s.addOffset(1000 * time.Millisecond)
			break
		case gui.KEY_LEFT_BRACKET:
			println("left bracket pressed")
			s.addOffset(-200 * time.Millisecond)
			break
		case gui.KEY_RIGHT_BRACKET:
			println("right bracket pressed")
			s.addOffset(200 * time.Millisecond)
			break
		}
	})
	return s
}

func (s *subtitle) setPosition(pos int) {
	// atomic.StoreInt32(&s.pos, int32(pos))
	s.Lock()
	defer s.Unlock()

	s.pos = pos
}
func (s *subtitle) position() int {
	s.Lock()
	defer s.Unlock()

	return s.pos
}
func (s *subtitle) increasePosition() {
	s.Lock()
	defer s.Unlock()

	s.pos += 1
}
func (s *subtitle) seek(t time.Duration) {
	s.Lock()
	defer s.Unlock()

	close(s.quit)
	for i, item := range s.items {
		to := item.To + time.Duration(s.offset)*time.Second
		if to > t {
			log.Print("seek to ", to.String(), " i: ", i, " Content:", item.Content)
			s.pos = i
			return
		}
	}
}

func (s *subtitle) addOffset(delta time.Duration) {
	s.Lock()
	s.offset += delta
	close(s.quit)
	s.pos = 0
	s.Unlock()

	go s.play()
}

func (s *subtitle) getOffset() time.Duration {
	s.Lock()
	s.Unlock()

	return s.offset
}
func (s *subtitle) playWithQuit(quit chan bool) {
	for s.position() < len(s.items) {
		item := s.items[s.position()]
		s.increasePosition()

		offset := s.getOffset()
		from := item.From + offset
		to := item.To + offset
		if to < s.c.GetTime() {
			continue
		}
		if s.c.WaitUtilWithQuit(from, quit) {
			s.w.PostEvent(gui.Event{gui.DrawSub, &srt.SubItem{}})
			return
		}

		s.w.PostEvent(gui.Event{gui.DrawSub, item})

		nextFrom := to
		nextPos := s.position()
		if nextPos < len(s.items) {
			nextFrom = s.items[nextPos].From + offset
		}

		go func(to, nextFrom time.Duration) { //overlap time, it's really nice with goroutine.
			if to > nextFrom {
				return
			}

			s.c.WaitUtilWithQuit(to-50*time.Millisecond, quit)

			s.w.PostEvent(gui.Event{gui.DrawSub, &srt.SubItem{}})
		}(to, nextFrom)
	}
}
func (s *subtitle) play() {
	s.quit = make(chan bool)
	s.playWithQuit(s.quit)
}
