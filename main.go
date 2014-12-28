package main

import (
	"bytes"
	"fmt"
	"github.com/deckarep/gosx-notifier"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const (
	THRESHOLD = 100
	DELAY     = 10 * time.Second
)

type processInfo struct {
	Name string
	PID  int64
	CPU  float64
}

type processList []processInfo

func getProcessList() processList {
	results := make(processList, 0)
	cmd := exec.Command("ps", "-o", "%cpu,pid,command", "-er")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	processes := strings.Split(out.String(), "\n")

	for i, process := range processes {
		if i == 0 {
			// Skip header
			continue
		}
		p := strings.TrimSpace(process)
		if len(p) <= 0 {
			continue
		}
		cpuEnd := strings.Index(p, " ")
		cpu, err := strconv.ParseFloat(p[:cpuEnd], 64)
		if err != nil {
			panic(fmt.Sprintf("%s", err))
		}

		p = p[cpuEnd:]
		p = strings.TrimSpace(p)

		pidEnd := strings.Index(p, " ")
		pid, err := strconv.ParseInt(strings.TrimSpace(p[:pidEnd]), 10, 64)
		if err != nil {
			panic(fmt.Sprintf("%s", err))
		}

		p = p[pidEnd:]
		name := strings.TrimSpace(p)

		results = append(results, processInfo{name, pid, cpu})
	}
	return results
}

func fetchName(p processInfo) string {
	idx := strings.LastIndex(p.Name, "/")
	if idx == -1 {
		return p.Name
	} else {
		return p.Name[idx+1:]
	}
}

func notify(p processInfo) {
	note := gosxnotifier.NewNotification("High CPU usage")
	note.Title = "High CPU usage detected"
	note.Subtitle = fmt.Sprintf("%s is using a lot of CPU", fetchName(p))
	note.Sound = gosxnotifier.Basso
	err := note.Push()
	if err != nil {
		panic(fmt.Sprintf("%s", err))
	}
}

func (x *processList) isBlacklisted(p processInfo) bool {
	for _, i := range *x {
		if i.PID == p.PID && i.Name == p.Name {
			return true
		}
	}
	return false
}

func (x *processList) getBlacklisting(p processInfo) int {
	for n, i := range *x {
		if i.PID == p.PID && i.Name == p.Name {
			return n
		}
	}
	return -1
}

func main() {
	blacklist := make(processList, 0)
	for {
		res := getProcessList()
		for _, i := range res {
			if i.CPU > THRESHOLD && !blacklist.isBlacklisted(i) {
				notify(i)
				blacklist = append(blacklist, i)
			} else {
				if x := blacklist.getBlacklisting(i); x != -1 {
					blacklist = append(blacklist[:x], blacklist[x+1:]...)
				}
			}
		}
		time.Sleep(DELAY)
	}
}
