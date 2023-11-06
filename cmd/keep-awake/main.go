package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/getlantern/systray"
	"github.com/getlantern/systray/example/icon"
)

const (
	EsSystemRequired = 0x00000001
	EsContinuous     = 0x80000000
)

var kernel32 *syscall.LazyDLL = syscall.NewLazyDLL("kernel32.dll")
var setThreadExecStateProc *syscall.LazyProc = kernel32.NewProc("SetThreadExecutionState")

// TODO: Less awful hack
func completed_successfully(err error) bool {
	return strings.Contains(err.Error(), "The operation completed successfully")
}

func Keep_awake() error {
	if _, _, err := setThreadExecStateProc.Call(uintptr(EsSystemRequired | EsContinuous)); err != nil && !completed_successfully(err) {
		return err
	}
	return nil
}

var INTERVALS = []time.Duration{
	time.Minute * 15,
	time.Minute * 30,
	time.Hour,
	time.Hour * 2,
	time.Hour * 4,
	time.Hour * 8,
	time.Hour * 12,
}

func main() {
	wg := &sync.WaitGroup{}
	awakemsg := make(chan time.Time, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	waitdone := &sync.WaitGroup{}

	quit := make(chan struct{}, 1)

	go func() {
		<-quit
		fmt.Printf("quit received\n")
		cancel()
		waitdone.Wait()
	}()

	systray.Run(
		func() {
			systray.SetTemplateIcon(icon.Data, icon.Data)
			running := systray.AddMenuItem("Not Running", "state of things")
			wg.Add(1)
			go func() {
				defer wg.Done()
				defer log.Printf("awakeloop done")
				var keepingAwakeUntil *time.Time
				for {
					select {
					case <-ctx.Done():
						return
					case t := <-awakemsg:
						fmt.Printf("got awakemsg %v\n", t)
						keepingAwakeUntil = &t
					case <-time.After(time.Second * 1):
					}
					if keepingAwakeUntil != nil {
						if time.Now().After(*keepingAwakeUntil) {
							fmt.Printf("clear ?\n")
							keepingAwakeUntil = nil
						} else {
							running.SetTitle(fmt.Sprintf("Keeping awake until %v", keepingAwakeUntil))
							if err := Keep_awake(); err != nil {
								fmt.Printf("failed to keep awake %v\n", err)
							}
						}
					} else {
						running.SetTitle("Not running")
					}
				}
			}()

			stop := systray.AddMenuItem("Stop", "Stop keeping awake")
			wg.Add(1)
			go func() {
				defer wg.Done()
				defer log.Printf("done with stop")
				for {
					select {
					case <-ctx.Done():
						return
					case <-stop.ClickedCh:
						awakemsg <- time.Now().Add(time.Second * -1)
					}
				}
			}()

			ka := systray.AddMenuItem("Keep Awake", "Keep the computer awake")

			for _, intvl := range INTERVALS {
				subitem := ka.AddSubMenuItem(fmt.Sprintf("keep awake for %v", intvl), fmt.Sprintf("keep awake for %v", intvl))
				wg.Add(1)
				iv := intvl
				go func() {
					defer wg.Done()
					defer log.Printf("done with %v", iv)
					for {
						select {
						case <-stop.ClickedCh:
							return
						case <-ctx.Done():
							return
						case <-subitem.ClickedCh:
						}
						awakemsg <- time.Now().Add(iv)
					}
				}()

			}

			quitItem := systray.AddMenuItem("Quit", "Shut down completely")
			wg.Add(1)
			go func() {
				defer wg.Done()
				select {
				case <-quitItem.ClickedCh:
					cancel()
				case <-ctx.Done():
				}
				systray.Quit()
			}()

		},
		func() {
			cancel()
			wg.Wait()
		},
	)
}
