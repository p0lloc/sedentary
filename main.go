package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/robotn/gohook"
)

const BreakInactivityDurationMs int64 = 30 * 1000

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Missing work and break duration arguments")
		return
	}

	s := Sedentary{
		workDurationSecs:  parseInt(os.Args[1]),
		breakDurationSecs: parseInt(os.Args[2]),
	}

	go s.listenForEvents()
	go s.runTimer()
	fmt.Scanln()
}

func parseInt(str string) int {
	v, _ := strconv.Atoi(str)
	return v
}

type Sedentary struct {
	breakDurationSecs int
	workDurationSecs  int

	timerDuration *int
	breakMode     bool
	lastActivity  int64

	currentNotificationId *string
}

func (s *Sedentary) runTimer() {
	duration := s.timerDuration
	if duration != nil {
		var typeName string
		if s.breakMode {
			typeName = "Break"
		} else {
			typeName = "Work"
		}

		fmt.Printf("%s (%d)\n", typeName, *duration)
		if *duration == 0 {
			// Timer duration is set to nil when user is in the intermediate period between break and work.
			// We're waiting for them to either stop using keyboard/mouse or start doing it again.
			s.timerDuration = nil

			if s.breakMode {
				s.deleteCurrentNotification()
				s.notify("Time to focus!")
			} else {
				s.notify("Break up sedentary time!")
			}

			s.breakMode = !s.breakMode
		} else {
			s.setDuration(*duration - 1)
		}
	} else {
		msSinceLastActivity := time.Now().UnixMilli() - s.lastActivity
		fmt.Printf("Time since last activity: %d ms\n", msSinceLastActivity)
		if s.breakMode && msSinceLastActivity > BreakInactivityDurationMs {
			// If user has been away from computer for the threshold time, count it as them taking a break
			s.deleteCurrentNotification()
			s.notify("Break in progress")
			s.setDuration(s.breakDurationSecs)
		}
	}

	time.Sleep(time.Second)
	s.runTimer()
}

func (s *Sedentary) notify(msg string) {
	res, err := exec.Command("notify-send", "-p", "--urgency=critical", msg).Output()
	if err != nil {
		panic(err)
	}

	str := strings.TrimSuffix(string(res), "\n")
	s.currentNotificationId = &str
}

func (s *Sedentary) deleteCurrentNotification() {
	if s.currentNotificationId == nil {
		return
	}

	if err := s.deleteNotification(*s.currentNotificationId); err != nil {
		panic(err)
	}
}

func (s *Sedentary) deleteNotification(notificationId string) error {
	return exec.Command("notify-send", fmt.Sprintf("--replace-id=%s", notificationId), "-p", "\" \"", "--expire-time=1").Run()
}

func (s *Sedentary) setDuration(duration int) {
	s.timerDuration = &duration
}

// Called when user is actively working on their computer.
func (s *Sedentary) onActivity() {
	s.lastActivity = time.Now().UnixMilli()
	if !s.breakMode && s.timerDuration == nil {
		// Wait until user comes back from break before starting a new work session
		s.deleteCurrentNotification()
		s.setDuration(s.workDurationSecs)
	}
}

// Listens for keyboard and mouse events and counts them as activity from the user.
func (s *Sedentary) listenForEvents() {

	events := hook.Start()
	defer hook.End()

	for event := range events {
		if event.Kind != hook.KeyDown && event.Kind != hook.MouseMove {
			continue
		}

		s.onActivity()
	}
}
