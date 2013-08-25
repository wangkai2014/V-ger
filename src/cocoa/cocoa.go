package cocoa

import (
	"fmt"
	"github.com/mkrautz/objc"
	. "github.com/mkrautz/objc/AppKit"
	. "github.com/mkrautz/objc/Foundation"
	"log"
	"os/exec"
	"path"
	"runtime"
	"task"
	"time"
	"util"
)

var config map[string]string

func init() {
	c := objc.NewClass(AppDelegate{})
	c.AddMethod("menuClick:", (*AppDelegate).MenuClick)
	// c.AddMethod("didActivateNotification:", (*AppDelegate).DidActivateNotification)

	objc.RegisterClass(c)
}

type AppDelegate struct {
	objc.Object `objc:"GOAppDelegate : NSObject"`
}

func (delegate *AppDelegate) MenuClick(sender uintptr) {
	if t, ok := task.GetDownloadingTask(); ok {
		cmd := exec.Command("open", path.Join(util.ReadConfig("dir"), t.Name))
		cmd.Start()
	} else {
		cmd := exec.Command("open", "/Applications/V'ger.app")
		cmd.Start()
	}
}

// func (delegate *AppDelegate) DidActivateNotification(notification objc.Object) {
// 	log.Print("DidActivateNotification")
// }

type uiCommand struct {
	name      string
	arguments interface{}
}

func timerStart(chUI chan uiCommand) {
	watch := make(chan *task.Task)

	log.Println("status bar watch task change: ", watch)
	task.WatchChange(watch)

	for t := range watch {
		var properties []string
		if t.Status == "Downloading" {
			properties = []string{fmt.Sprintf("%s %.1f%%", util.CleanMovieName(t.Name),
				float64(t.DownloadedSize)/float64(t.Size)*100.0),
				fmt.Sprintf("%.2f KB/s %s", t.Speed, t.Est)}
		} else if t.Status == "Playing" {
			properties = []string{fmt.Sprintf("%s %.1f KB/s", util.CleanMovieName(t.Name), t.Speed), ""}
		} else {
			if !task.HasDownloadingOrPlaying() {
				properties = []string{"V'ger"}
			}
		}

		chUI <- uiCommand{"statusItem", properties}
	}
}

var chUI chan uiCommand

// func SendNotification(title string, infoText string) {
// 	chUI <- uiCommand{"sendNotification", []string{title, infoText}}
// }
// func TrashFile(dir string, name string) {
// 	chUI <- uiCommand{"trashFile", []string{dir, name}}
// }

func Start() {
	runtime.LockOSThread()

	pool := NewNSAutoreleasePool()

	// InstallNSBundleHook()

	delegate := objc.GetClass("GOAppDelegate").Alloc().Init()

	app := NSSharedApplication()

	NSDefaultUserNotificationCenter().SetDelegate(delegate)

	statusItem := NSStatusItem{NSSystemStatusBar().StatusItemWithLength(-1).Retain()}
	statusItem.SetHighlightMode(true)
	statusItem.SetTarget(delegate.Pointer())
	statusItem.SetAction(objc.GetSelector("menuClick:"))
	statusItem.SetTitle("V'ger")

	chUI = make(chan uiCommand)

	go timerStart(chUI)

	for {
		pool.Release()
		pool = NewNSAutoreleasePool()

		event := app.NextEventMatchingMask(0xffffff, NSDateWithTimeIntervalSinceNow(0.05),
			"kCFRunLoopDefaultMode", true)

		app.SendEvent(event)
		// app.UpdateWindows()

		t := time.After(time.Millisecond * 100)
		select {
		case cmd := <-chUI:
			switch cmd.name {
			case "statusItem":
				prop := cmd.arguments.([]string)
				if len(prop) == 0 {
					break
				}

				statusItem.SetTitle(prop[0])

				if len(prop) > 1 {
					statusItem.SetToolTip(prop[1])
				} else {
					statusItem.SetToolTip("")
				}
				break
			// case "sendNotification":
			// 	args := cmd.arguments.([]string)
			// 	title := args[0]
			// 	infoText := args[1]

			// 	notification := NSUserNotification{objc.GetClass("NSUserNotification").Alloc().Init()}
			// 	notification.SetTitle(title)
			// 	notification.SetInformativeText(infoText)
			// 	notification.SetSoundName(NSUserNotificationDefaultSoundName)
			// 	notification.SetHasActionButton(true)
			// 	notification.SetActionButtonTitle("Open")

			// 	center := NSDefaultUserNotificationCenter()
			// 	center.DeliverNotification(notification)

			// 	break
			// case "trashFile":
			// 	prop := cmd.arguments.([]string)
			// 	NSTrashFile(prop[0], prop[1])
			default:
				log.Printf("unknown cmd %v", cmd)
				break
			}
			break
		case <-t:
			break
		}
	}

	statusItem.Release()
	pool.Release()
}
