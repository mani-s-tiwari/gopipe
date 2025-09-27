package gopipe

import (
	"context"
	"sync"
)

type ManagerWorker struct {
	managerCh chan *Task
	workers   []chan *Task
	handler   Handler
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
}

func NewManagerWorker(numWorkers int, handler Handler) *ManagerWorker {
	ctx, cancel := context.WithCancel(context.Background())
	mw := &ManagerWorker{
		managerCh: make(chan *Task, 1000),
		workers:   make([]chan *Task, numWorkers),
		handler:   handler,
		ctx:       ctx,
		cancel:    cancel,
	}
	for i := range mw.workers {
		mw.workers[i] = make(chan *Task, 100)
		mw.wg.Add(1)
		go mw.worker(i, mw.workers[i])
	}
	mw.wg.Add(1)
	go mw.manager()
	return mw
}

func (mw *ManagerWorker) Submit(task *Task) error {
	select {
	case mw.managerCh <- task:
		return nil
	case <-mw.ctx.Done():
		return context.Canceled
	}
}

func (mw *ManagerWorker) SubmitAndWait(task *Task) TaskResult {
	if err := mw.Submit(task); err != nil {
		return TaskResult{TaskID: task.ID, Error: err, Success: false}
	}
	return task.WaitForResult()
}

func (mw *ManagerWorker) manager() {
	defer mw.wg.Done()
	idx := 0
	for {
		select {
		case <-mw.ctx.Done():
			return
		case t := <-mw.managerCh:
			mw.workers[idx%len(mw.workers)] <- t
			idx++
		}
	}
}

func (mw *ManagerWorker) worker(id int, ch chan *Task) {
	defer mw.wg.Done()
	for {
		select {
		case <-mw.ctx.Done():
			return
		case t := <-ch:
			_ = mw.handler(mw.ctx, t)
		}
	}
}

func (mw *ManagerWorker) Stop() {
	mw.cancel()
	mw.wg.Wait()
}
