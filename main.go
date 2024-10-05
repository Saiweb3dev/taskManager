package main

import (
	"log"

	"taskManager/taskmanager"
)

func main() {
	config := taskmanager.Config{
		Filename: "tasks.json",
		Port:     9000,
	}
	tm := taskmanager.NewTaskManager(config)
	err := tm.Run()
	if err != nil {
		log.Fatal("Error running task manager:", err)
	}
}