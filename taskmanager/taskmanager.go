package taskmanager

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/nsf/termbox-go"
)

type TaskStatus string

const (
	NotCompleted TaskStatus = "Not Completed"
	InProgress   TaskStatus = "In Progress"
	Completed    TaskStatus = "Completed"
)

type TaskData struct {
	ID          int        `json:"id"`
	Description string     `json:"description"`
	Status      TaskStatus `json:"status"`
	CreatedAt   time.Time  `json:"createdAt"`
}

type Config struct {
	Filename string
	Port     int
}

type TaskManager struct {
	filename string
	port     int
}

func NewTaskManager(config Config) *TaskManager {
	if config.Filename == "" {
		config.Filename = "tasks.json"
	}
	if config.Port == 0 {
		config.Port = 9000
	}
	return &TaskManager{
		filename: config.Filename,
		port:     config.Port,
	}
}

func (tm *TaskManager) Run() error {
	go tm.startWebServer()

	err := termbox.Init()
	if err != nil {
		return err
	}
	defer termbox.Close()

	for {
		tm.drawMenu()
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Ch {
			case '1':
				tm.addTask()
			case '2':
				tm.updateTask()
			case '3':
				tm.showAllTasks()
			case '4':
				tm.showTasksByStatus(Completed)
			case '5':
				tm.showTasksByStatus(InProgress)
			case '6':
				tm.showTasksByStatus(NotCompleted)
			case '7':
				tm.deleteTask()
			case '8':
				return nil
			}
		case termbox.EventError:
			return ev.Err
		}
	}
}

func (tm *TaskManager) startWebServer() {
	http.HandleFunc("/", tm.handleHome)
	http.HandleFunc("/tasks", tm.handleTasks)
	http.HandleFunc("/add", tm.handleAddTask)
	http.HandleFunc("/update", tm.handleUpdateTask)
	http.HandleFunc("/delete", tm.handleDeleteTask)

	fmt.Printf("Web server running on http://localhost:%d\n", tm.port)
	http.ListenAndServe(fmt.Sprintf(":%d", tm.port), nil)
}

func (tm *TaskManager) handleHome(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to Task Manager. Available endpoints: /tasks, /add, /update, /delete")
}

func (tm *TaskManager) handleTasks(w http.ResponseWriter, r *http.Request) {
	tasks := tm.loadTasks()
	json.NewEncoder(w).Encode(tasks)
}

func (tm *TaskManager) handleAddTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var task TaskData
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tasks := tm.loadTasks()
	task.ID = len(tasks) + 1
	task.CreatedAt = time.Now()
	tasks = append(tasks, task)
	tm.saveTasks(tasks)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

func (tm *TaskManager) handleUpdateTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var updatedTask TaskData
	err := json.NewDecoder(r.Body).Decode(&updatedTask)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tasks := tm.loadTasks()
	for i, task := range tasks {
		if task.ID == updatedTask.ID {
			tasks[i] = updatedTask
			tm.saveTasks(tasks)
			json.NewEncoder(w).Encode(updatedTask)
			return
		}
	}

	http.Error(w, "Task not found", http.StatusNotFound)
}

func (tm *TaskManager) handleDeleteTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	tasks := tm.loadTasks()
	for i, task := range tasks {
		if task.ID == id {
			tasks = append(tasks[:i], tasks[i+1:]...)
			tm.saveTasks(tasks)
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}

	http.Error(w, "Task not found", http.StatusNotFound)
}

func (tm *TaskManager) drawMenu() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	tm.drawCenteredText(0, "--- Task Manager ---", termbox.ColorCyan, termbox.ColorDefault)
	tm.drawCenteredText(2, "1. Add Task", termbox.ColorYellow, termbox.ColorDefault)
	tm.drawCenteredText(3, "2. Update Task", termbox.ColorYellow, termbox.ColorDefault)
	tm.drawCenteredText(4, "3. Show All Tasks", termbox.ColorYellow, termbox.ColorDefault)
	tm.drawCenteredText(5, "4. Show Completed Tasks", termbox.ColorYellow, termbox.ColorDefault)
	tm.drawCenteredText(6, "5. Show In Progress Tasks", termbox.ColorYellow, termbox.ColorDefault)
	tm.drawCenteredText(7, "6. Show Not Completed Tasks", termbox.ColorYellow, termbox.ColorDefault)
	tm.drawCenteredText(8, "7. Delete Task", termbox.ColorYellow, termbox.ColorDefault)
	tm.drawCenteredText(9, "8. Exit", termbox.ColorRed, termbox.ColorDefault)
	termbox.Flush()
}

func (tm *TaskManager) drawCenteredText(y int, text string, fg, bg termbox.Attribute) {
	width, _ := termbox.Size()
	x := (width - len(text)) / 2
	for i, ch := range text {
		termbox.SetCell(x+i, y, ch, fg, bg)
	}
}

func (tm *TaskManager) getUserInput(prompt string) string {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	tm.drawCenteredText(0, prompt, termbox.ColorWhite, termbox.ColorDefault)
	termbox.Flush()

	var input []rune
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			if ev.Key == termbox.KeyEnter {
				return string(input)
			} else if ev.Key == termbox.KeyBackspace || ev.Key == termbox.KeyBackspace2 {
				if len(input) > 0 {
					input = input[:len(input)-1]
				}
			} else if ev.Key == termbox.KeySpace {
				input = append(input, ' ')
			} else if ev.Ch != 0 {
				input = append(input, ev.Ch)
			}
			termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
			tm.drawCenteredText(0, prompt, termbox.ColorWhite, termbox.ColorDefault)
			tm.drawCenteredText(2, string(input), termbox.ColorYellow, termbox.ColorDefault)
			termbox.Flush()
		}
	}
}

func (tm *TaskManager) showLoadingAnimation(done chan bool) {
	frames := []string{"|", "/", "-", "\\"}
	i := 0
	for {
		select {
		case <-done:
			return
		default:
			termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
			tm.drawCenteredText(0, "Loading "+frames[i], termbox.ColorCyan, termbox.ColorDefault)
			termbox.Flush()
			time.Sleep(100 * time.Millisecond)
			i = (i + 1) % len(frames)
		}
	}
}

func (tm *TaskManager) loadTasks() []TaskData {
	var tasks []TaskData
	file, err := os.ReadFile(tm.filename)
	if err != nil {
		return tasks
	}
	json.Unmarshal(file, &tasks)
	return tasks
}

func (tm *TaskManager) saveTasks(tasks []TaskData) {
	file, _ := json.MarshalIndent(tasks, "", "  ")
	os.WriteFile(tm.filename, file, 0644)
}

func (tm *TaskManager) addTask() {
	description := tm.getUserInput("Enter task description: ")
	done := make(chan bool)
	go tm.showLoadingAnimation(done)

	tasks := tm.loadTasks()
	newTask := TaskData{
		ID:          len(tasks) + 1,
		Description: description,
		Status:      NotCompleted,
		CreatedAt:   time.Now(),
	}
	tasks = append(tasks, newTask)
	tm.saveTasks(tasks)

	done <- true
	tm.drawCenteredText(0, "Task added successfully.", termbox.ColorGreen, termbox.ColorDefault)
	termbox.Flush()
	time.Sleep(2 * time.Second)
}

func (tm *TaskManager) updateTask() {
	tasks := tm.loadTasks()
	tm.showAllTasks()
	idStr := tm.getUserInput("Enter task ID to update: ")
	id, _ := strconv.Atoi(idStr)

	for i, task := range tasks {
		if task.ID == id {
			termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
			tm.drawCenteredText(0, "1. Not Completed", termbox.ColorYellow, termbox.ColorDefault)
			tm.drawCenteredText(1, "2. In Progress", termbox.ColorYellow, termbox.ColorDefault)
			tm.drawCenteredText(2, "3. Completed", termbox.ColorYellow, termbox.ColorDefault)
			termbox.Flush()

			statusChoice := tm.getUserInput("Enter new status (1-3): ")
			done := make(chan bool)
			go tm.showLoadingAnimation(done)

			switch statusChoice {
			case "1":
				tasks[i].Status = NotCompleted
			case "2":
				tasks[i].Status = InProgress
			case "3":
				tasks[i].Status = Completed
			default:
				done <- true
				tm.drawCenteredText(0, "Invalid choice. Status not updated.", termbox.ColorRed, termbox.ColorDefault)
				termbox.Flush()
				time.Sleep(2 * time.Second)
				return
			}
			tm.saveTasks(tasks)
			done <- true
			tm.drawCenteredText(0, "Task updated successfully.", termbox.ColorGreen, termbox.ColorDefault)
			termbox.Flush()
			time.Sleep(2 * time.Second)
			return
		}
	}
	tm.drawCenteredText(0, "Task not found.", termbox.ColorRed, termbox.ColorDefault)
	termbox.Flush()
	time.Sleep(2 * time.Second)
}

func (tm *TaskManager) showAllTasks() {
	tasks := tm.loadTasks()
	tm.showTasks(tasks)
}

func (tm *TaskManager) showTasksByStatus(status TaskStatus) {
	tasks := tm.loadTasks()
	filteredTasks := []TaskData{}
	for _, task := range tasks {
		if task.Status == status {
			filteredTasks = append(filteredTasks, task)
		}
	}
	tm.showTasks(filteredTasks)
}

func (tm *TaskManager) showTasks(tasks []TaskData) {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	if len(tasks) == 0 {
		tm.drawCenteredText(0, "No tasks found.", termbox.ColorYellow, termbox.ColorDefault)
		termbox.Flush()
		time.Sleep(2 * time.Second)
		return
	}

	for i, task := range tasks {
		var color termbox.Attribute
		switch task.Status {
		case Completed:
			color = termbox.ColorGreen
		case InProgress:
			color = termbox.ColorYellow
		case NotCompleted:
			color = termbox.ColorRed
		}
		taskStr := fmt.Sprintf("ID: %d, Description: %s, Status: %s, Created At: %s",
			task.ID, task.Description, task.Status, task.CreatedAt.Format(time.RFC822))
		tm.drawCenteredText(i, taskStr, color, termbox.ColorDefault)
	}
	termbox.Flush()
	termbox.PollEvent()
}

func (tm *TaskManager) deleteTask() {
	tasks := tm.loadTasks()
	tm.showAllTasks()
	idStr := tm.getUserInput("Enter task ID to delete: ")
	id, _ := strconv.Atoi(idStr)

	done := make(chan bool)
	go tm.showLoadingAnimation(done)

	for i, task := range tasks {
		if task.ID == id {
			tasks = append(tasks[:i], tasks[i+1:]...)
			tm.saveTasks(tasks)
			done <- true
			tm.drawCenteredText(0, "Task deleted successfully.", termbox.ColorGreen, termbox.ColorDefault)
			termbox.Flush()
			time.Sleep(2 * time.Second)
			return
		}
	}
	done <- true
	tm.drawCenteredText(0, "Task not found.", termbox.ColorRed, termbox.ColorDefault)
	termbox.Flush()
	time.Sleep(2 * time.Second)
}