package main

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"gitlab.com/nikonov1101/go-for-milk/rtm"
)

func main() {
	key, secret := loadKeys()

	cli, err := rtm.New(key, secret)
	if err != nil {
		panic(err)
	}

	// invoked with some args, lets make a task from them,
	// keep in mind that priority is set with "!n", where n is 1..3,
	// but zsh requires the escape of "!", so it will looks like this:
	// go run main.go this is important stuff \!2
	if len(os.Args) > 1 {
		name := strings.Join(os.Args[1:], " ")
		if err := cli.AddTask(name); err != nil {
			panic(err)
		}

		fmt.Printf("OK: task %q added.", name)
		os.Exit(0)
	}

	// no args given, just list all tasks
	tasks, err := cli.ListTasks()
	if err != nil {
		panic(err)
	}

	sort.Slice(tasks, func(i, j int) bool {
		if tasks[i].Priority != tasks[j].Priority {
			// sort by priority first
			return tasks[i].Priority > tasks[j].Priority
		}
		// then by created_at, if priorities are equal
		return tasks[i].CreatedAt.After(tasks[j].CreatedAt)
	})

	for _, task := range tasks {
		if task.Visible() { // task is not visible if it's completed OR deleted
			fmt.Println(colorizeTask(task))
		}
	}
}

func loadKeys() (string, string) {
	// TODO(nikonov): or pass them via -flags, or introduce the config file...
	key := os.Getenv("RTM_APIKEY")
	secret := os.Getenv("RTM_SECRET")
	if len(key) == 0 {
		fmt.Println("no RTM_APIKEY env variable set")
		os.Exit(1)
	}
	if len(secret) == 0 {
		fmt.Println("no RTM_SECRET env variable set")
		os.Exit(1)
	}

	return key, secret
}

func colorizeTask(t rtm.Task) string {
	prioNum := ""
	text := t.Name
	switch t.Priority {
	case 0:
		prioNum = white(" ")
	case 1:
		prioNum = green("!")
	case 2:
		prioNum = yellow("!")
	case 3:
		prioNum = red("!")
		text = white(text)
	default:
		return strconv.Itoa(t.Priority)
	}

	tags := ""
	if len(t.Tags) > 0 {
		tmp := make([]string, len(t.Tags))
		for i := range t.Tags {
			tmp[i] = "#" + t.Tags[i]
		}
		tags = " " + gray(strings.Join(tmp, ", "))
	}
	return fmt.Sprintf("[%s] %s%s", prioNum, text, tags)
}

const (
	RED = "\033[1;31m"
	YEL = "\033[1;33m"
	GRE = "\033[1;32m"
	WHT = "\033[1;37m"
	GRY = "\033[1;30m"
	NC  = "\033[0m" // No Color
)

func green(s string) string {
	return GRE + s + NC
}

func yellow(s string) string {
	return YEL + s + NC
}

func red(s string) string {
	return RED + s + NC
}

func white(s string) string {
	return WHT + s + NC
}

func gray(s string) string {
	return GRY + s + NC
}
