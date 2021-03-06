package download

import (
	"log"
	"time"
)

type SpeedMonitor interface {
	SetSpeed(float64)
}

type speedFilter struct {
	basicFilter
	sm SpeedMonitor
}

func (sf *speedFilter) active() {
	defer sf.closeOutput()

	timer := time.NewTicker(time.Second)
	sr := newSegRing(40)
	for {
		select {
		case b, ok := <-sf.input:
			if !ok {
				if sf.sm != nil {
					sf.sm.SetSpeed(0)
				}
				return
			}
			// trace(fmt.Sprint("speed filter input:", b.from, b.to))

			sf.writeOutput(b)
			// trace(fmt.Sprint("speed filter output:", b.from, b.to))

			sr.add(int64(len(b.Data)))
			break
		case <-timer.C:
			// println("speed filter tick")
			sr.add(0)
			sr.add(0)
			sr.add(0)
			if sf.sm != nil {
				sf.sm.SetSpeed(sr.calcSpeed())
			} else {
				println("speed monitor is nil")
			}
		case <-sf.quit:
			if sf.sm != nil {
				sf.sm.SetSpeed(0)
			}

			log.Print("speed quit")
			return
		}
	}
}
