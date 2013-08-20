package download

import (
	// "errors"
	"fmt"
	"io"
	"task"
	"time"
	// "bytes"
	// "log"
	// "bytes"
)

var play_quit chan bool

func Play(t *task.Task, w io.Writer, from, to int64) {
	fmt.Println("playing download from ", from, " to ", to)
	if play_quit != nil {
		select {
		case <-play_quit:
			// Since no one write to quit channel,
			// the channel must be closed when pass through receive operation.
			break
		case <-time.After(time.Millisecond):
			close(play_quit)
		}
	}

	t.Status = "Playing"
	task.SaveTask(t)

	play_quit = make(chan bool)

	control := make(chan int)
	progress := doDownload(t.URL, w, from, to, 0, control, play_quit)
	handleProgress(progress, t, play_quit)
}
