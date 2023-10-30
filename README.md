# 2do

A simple command line task manager. Made for tasks with specific start and end dates.

## Usage

```sh
2do <command> [arguments]
```

```txt
The commands are:
    list      list all tasks
              [-l <limit_number>] [-t <theme1>[,<theme2>[,...]]]
              -l --limit <limit>: Only list <limit> number of tasks
              -t --theme <theme1,theme2,...>: Only list tasks with specific themes
    add       add a new task
              <start time> <stop time> <description> <theme> \
              [-r <repeat days> <repeat times>]
              -r, --repeat: Repeat the task every <repeat days>, <repeat times> 
    delete    delete one or more tasks 
              <id> [<id> ...]
    done      mark one or more tasks as done
              [-u] <id> [<id> ...]
              -u, --undo: mark tasks as not done
```