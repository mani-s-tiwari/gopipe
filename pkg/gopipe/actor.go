package gopipe

import (
	"context"
	"sync"
)

type ActorSystem struct {
	actors map[string]chan *Task
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewActorSystem() *ActorSystem {
	ctx, cancel := context.WithCancel(context.Background())
	return &ActorSystem{
		actors: make(map[string]chan *Task),
		ctx:    ctx,
		cancel: cancel,
	}
}

func (a *ActorSystem) RegisterActor(name string, handler Handler) {
	ch := make(chan *Task, 100)
	a.actors[name] = ch
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		for {
			select {
			case <-a.ctx.Done():
				return
			case t := <-ch:
				_ = handler(a.ctx, t)
			}
		}
	}()
}

func (a *ActorSystem) Submit(task *Task) error {
	if ch, ok := a.actors[task.Name]; ok {
		ch <- task
		return nil
	}
	return ErrHandlerNotFound
}

func (a *ActorSystem) SubmitAndWait(task *Task) TaskResult {
	if err := a.Submit(task); err != nil {
		return TaskResult{TaskID: task.ID, Error: err, Success: false}
	}
	return task.WaitForResult()
}

func (a *ActorSystem) Stop() {
	a.cancel()
	a.wg.Wait()
}
