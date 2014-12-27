package dispatcher

import "github.com/bearded-web/bearded/models/task"

type Dispatcher struct {
	TaskManager task.Manager
}
