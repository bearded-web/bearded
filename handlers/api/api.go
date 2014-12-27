package api

import (
	"net/http"

	"github.com/bearded-web/bearded/models/task"
	"github.com/bearded-web/bearded/modules/dispatcher"
	"github.com/gin-gonic/gin"
	"github.com/bearded-web/bearded/modules/worker"
	"encoding/json"
	"bytes"
	"io/ioutil"
)

const CallbackHeader = "X-Callback-Addr"

type DispatcherApi struct {
	dispatcher *dispatcher.Dispatcher
	workers	   *worker.Manager
}

func New(dispatcher *dispatcher.Dispatcher, workers *worker.Manager) *DispatcherApi {
	return &DispatcherApi{
		dispatcher: dispatcher,
		workers: workers,
	}
}

func (a *DispatcherApi) TaskCreate(ctx *gin.Context) {
	t := task.New()
	if ctx.Bind(t) == false {
		return // 400 response
	}
	a.dispatcher.TaskManager.Add(t)
	if callbackUrl := ctx.Request.Header.Get(CallbackHeader); callbackUrl != "" {
		// TODO(m0sth8): rewrite this hack
		t.OnStateChange(func(t *task.Task){
			println("callback, task.status:", t.State.Status)
			data, err := json.Marshal(t)
			if err != nil {
				println(err)
			}
			resp, err := http.Post(callbackUrl, "application/json", bytes.NewBuffer(data))
			if err != nil {
				println(err)
				return
			}
			println("callback: ", callbackUrl, resp.StatusCode)
			responseData, _ := ioutil.ReadAll(resp.Body)
			println("response: ", string(responseData))
		})

	}
	a.workers.Queue() <- t
	ctx.JSON(http.StatusCreated, t)
}

func (a *DispatcherApi) TaskList(ctx *gin.Context) {
	tasks, err := a.dispatcher.TaskManager.All()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, NewApiError(0, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, tasks)
	return
}

func (a *DispatcherApi) TaskGet(ctx *gin.Context, task *task.Task) {
	ctx.JSON(http.StatusOK, task)
}

func (a *DispatcherApi) TaskDelete(ctx *gin.Context, t *task.Task) {
	if err := a.dispatcher.TaskManager.Delete(t.Id); err != nil {
		ctx.JSON(http.StatusInternalServerError, NewApiError(0, err.Error()))
		return
	}
	ctx.String(http.StatusOK, "")
}

func (a *DispatcherApi) TaskStateGet(ctx *gin.Context, t *task.Task) {
	ctx.JSON(http.StatusOK, t.State)
}

func (a *DispatcherApi) TaskReportGet(ctx *gin.Context, t *task.Task) {
	if t.Report == nil {
		ctx.JSON(http.StatusNotFound, NewApiError(1, "Report was not found"))
		return
	}
	ctx.JSON(http.StatusOK, t.Report)
}

// Middleware which is trying to find task with :taskId param.
func (a *DispatcherApi) TakeTask(handler func(*gin.Context, *task.Task)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Params.ByName("taskId")
		task, err := a.dispatcher.TaskManager.Get(id)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, NewApiError(0, err.Error()))
			return
		}
		if task == nil {
			ctx.JSON(http.StatusNotFound, NewApiError(1, "Task was not found"))
			return
		}
		handler(ctx, task)
	}
}
