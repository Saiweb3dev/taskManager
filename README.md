# TaskManager

TaskManager is a simple, terminal-based task management application with a web interface. It allows you to manage tasks with features like adding, updating, deleting, and viewing tasks.

## Features

- Terminal-based user interface
- Web interface for remote access
- Add, update, delete, and view tasks
- Filter tasks by status (Completed, In Progress, Not Completed)
- Persistent storage using JSON file

## Installation

To install TaskManager, you need to have Go installed on your system. Then, you can use the following command:

```bash
go get github.com/yourusername/taskmanager
```

## Usage

### As a Library

To use TaskManager in your Go project:


```go
package main
import (
"log"
"github.com/yourusername/taskmanager/internal/taskmanager"
)
func main() {
config := taskmanager.Config{
Filename: "tasks.json",
Port: 9000,
}
tm := taskmanager.NewTaskManager(config)
err := tm.Run()
if err != nil {
log.Fatal("Error running task manager:", err)
}
}
```


### As a Standalone Application

To run TaskManager as a standalone application:


```bash
go run main.go
```


## Terminal Interface

Once running, you'll see a menu with the following options:

1. Add Task
2. Update Task
3. Show All Tasks
4. Show Completed Tasks
5. Show In Progress Tasks
6. Show Not Completed Tasks
7. Delete Task
8. Exit

Use the number keys to select an option.

## Web Interface

The web interface runs on `http://localhost:9000` (or the port you specified in the configuration). You can use the following endpoints:

- GET `/tasks`: List all tasks
- POST `/add`: Add a new task
- POST `/update`: Update an existing task
- POST `/delete?id=<task_id>`: Delete a task

### Example API Usage

Add a task:

```bash
curl -X POST http://localhost:9000/add -H "Content-Type: application/json" -d '{"description":"New task", "status":"Not Completed"}'
```

## Configuration

You can configure the TaskManager with the following options:

- `Filename`: The name of the JSON file to store tasks (default: "tasks.json")
- `Port`: The port number to run the web interface on (default: 9000)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.