package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 54320
	user     = "user"
	password = "password"
	dbname   = "2do"
)

const (
	timeFormat = "200601021504"
)

var HELPMSG string = `2do is a simple command line task manager.
Usage:
	2do <command> [arguments]

The commands are:
    list      list all tasks        
    add       add a new task
              <start time> <stop time> <description> <theme>
    delete    delete one or more tasks 
              <id> [<id> ...]
    done      mark one or more tasks as done
              <id> [<id> ...]
`

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println(HELPMSG)
		os.Exit(1)
	}

	// Connect to the PostgreSQL database and create the tasks table
	// it it does not already exist
	dbinfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", dbinfo)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create a table for tasks
	createTable := `
        CREATE TABLE IF NOT EXISTS tasks (
            id SERIAL PRIMARY KEY,
            start_time TIMESTAMP,
            stop_time TIMESTAMP,
            description TEXT,
            theme TEXT,
            is_done BOOLEAN
        )
    `

	_, err = db.Exec(createTable)
	if err != nil {
		log.Fatal(err)
	}

	// Handle the command line arguments
	switch args[0] {
	case "help":
		fmt.Println(HELPMSG)
		os.Exit(0)
	case "list":
		listTasks(db)
		os.Exit(0)
	case "delete":
		// Delete an event from the tasks table
		deleteEvent := `
			DELETE FROM tasks WHERE id = $1
		`
		ids := args[1:]
		if len(ids) == 0 {
			fmt.Println("No ids provided to delete.")
			os.Exit(1)
		}
		for _, id := range ids {
			_, err := db.Exec(deleteEvent, id)
			if err != nil {
				fmt.Printf("Error deleting event with id %s: %v\n", id, err)
			}
		}
		os.Exit(0)
	case "done":
		doneEvent := `
			UPDATE tasks SET is_done = true WHERE id = $1
		`
		ids := args[1:]
		if len(ids) == 0 {
			fmt.Printf("No ids provided to mark as done.")
			os.Exit(1)
		}
		for _, id := range ids {
			_, err := db.Exec(doneEvent, id)
			if err != nil {
				fmt.Printf("Error marking event with id %s as done: %v\n", id, err)
			}
		}
		os.Exit(0)
	case "add":
		addTask(db, args[1:])
		os.Exit(0)
	}

	os.Exit(0)

}

// parseTime parses a time string on the format YYMMDDHHMM
// and returns a time.Time object in UTC
func parseTime(timeString string) (time.Time, error) {

	timeNowString := time.Now().Local().Format(timeFormat)
	if len(timeString) > len(timeNowString) {
		return time.Time{}, fmt.Errorf("Time has invalid format")
	}
	l := len(timeNowString)
	m := len(timeString)
	for len(timeString) < len(timeNowString) {
		timeString = " " + timeString
	}
	joinedTimeString := ""
	for i := 0; i < len(timeNowString); i++ {
		if i < l-m+1 {
			joinedTimeString += string(timeNowString[i])
		} else {
			joinedTimeString += string(timeString[i])
		}
	}
	t, err := time.ParseInLocation(timeFormat, joinedTimeString, time.Now().Location())
	if err != nil {
		return time.Time{}, err
	}
	return t.UTC(), nil
}

func addTask(db *sql.DB, args []string) {
	i := 0
	popArg := func() (string, bool) {
		if i >= len(args) {
			return "", true
		}
		i++
		return args[i-1], false
	}
	// Insert an event into the tasks table
	insertEvent := `
        INSERT INTO tasks (start_time, stop_time, description, theme, is_done)
        VALUES ($1, $2, $3, $4, $5)
    `

	var startTime, stopTime time.Time
	var description, theme string
	// Command is of the form:
	// 2do add <start time> <stop time> <description> <theme>
	// Time is on the format YYMMDDHHMM

	// Start time
	startArg, done := popArg()
	if done {
		fmt.Printf("No start time provided.\n")
		os.Exit(1)
	}
	startTime, err := parseTime(startArg)
	if err != nil {
		fmt.Printf("Error parsing start time: %v\n", err)
		os.Exit(1)
	}

	// Stop time
	StopArg, done := popArg()
	if done {
		fmt.Printf("No stop time provided.\n")
		os.Exit(1)
	}
	stopTime, err = parseTime(StopArg)
	if err != nil {
		fmt.Printf("Error parsing stop time: %v\n", err)
		os.Exit(1)
	}

	description, done = popArg()
	if done {
		fmt.Printf("No description provided.\n")
		os.Exit(1)
	}

	theme, done = popArg()
	if done {
		fmt.Printf("No theme provided.\n")
		os.Exit(1)
	}

	_, err = db.Exec(insertEvent, startTime, stopTime, description, theme, false)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Event inserted successfully.")

}

func listTasks(db *sql.DB) {
	// only list tasks that are not done
	listTasks := `
		SELECT * FROM tasks WHERE is_done = false ORDER BY stop_time ASC
	`

	rows, err := db.Query(listTasks)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	// Iterate over the rows and print the results
	for rows.Next() {
		var (
			id          int
			startTime   time.Time
			stopTime    time.Time
			description string
			theme       string
			isDone      bool
		)
		if err := rows.Scan(&id, &startTime, &stopTime, &description, &theme, &isDone); err != nil {
			log.Fatal(err)
		}

		// format color of start time
		startString := startTime.Local().Format("Jan02 1504")
		if time.Now().UTC().After(startTime) {
			startString = fmt.Sprintf("\x1b[31m%s\x1b[0m", startString)
		} else {
			startString = fmt.Sprintf("\x1b[32m%s\x1b[0m", startString)
		}

		// color of stop time
		stopString := stopTime.Local().Format("Jan02 1504")
		if time.Now().UTC().After(stopTime) {
			stopString = fmt.Sprintf("\x1b[31m%s\x1b[0m", stopString)
		} else {
			stopString = fmt.Sprintf("\x1b[32m%s\x1b[0m", stopString)
		}

		// done marker
		doneMarker := " "
		if isDone {
			doneMarker = "\x1b[32m✓\x1b[0m"
		}

		// format color of end time
		fmt.Printf("\x1b[34m%04d\x1b[0m: %s - %s [%s] \x1b[38;5;%vm%s\x1b[0m, \x1b[38;5;8m%s\x1b[0m\n",
			id, startString, stopString, doneMarker, 8 /*rand.Intn(256)*/, theme, description)
	}

	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

}
