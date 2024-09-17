package main

import (
	"log"
	"os"
	"sort"

	"gitlab.com/nikonov1101/go-for-milk/rtm"
)

func main() {
	key, secret := loadKeys()

	cli, err := rtm.New(key, secret)
	if err != nil {
		panic(err)
	}

	// add a task to the Inbox, note that smart parsing works here, so #^! are turned into the task meta-data
	if err := cli.AddTask("this is my new task using my new client #random !2 ^tomorrow"); err != nil {
		panic(err)
	}

	// query back your tasks, returns all items from all list
	// as a single, flat array, for details see implementation in rtm/types.go @ intoTasks()
	tasks, err := cli.ListTasks()
	if err != nil {
		panic(err)
	}

	// client does not sort tasks, but we can do so by any criteria we want
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].Priority > tasks[j].Priority
	})

	for _, task := range tasks {
		if task.Visible() { // task is not visible if it's completed OR deleted
			log.Printf("[%d] %s :: #%v", task.Priority, task.Name, task.Tags)
		}
	}
}

func loadKeys() (string, string) {
	// TODO(nikonov): or pass them via -flags, or introduce the config file...
	key := os.Getenv("RTM_APIKEY")
	secret := os.Getenv("RTM_SECRET")
	if len(key) == 0 {
		log.Printf("no RTM_APIKEY env variable set")
		os.Exit(1)
	}
	if len(secret) == 0 {
		log.Printf("no RTM_SECRET env variable set")
		os.Exit(1)
	}

	return key, secret
}
