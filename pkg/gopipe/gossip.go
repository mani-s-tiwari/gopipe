package gopipe

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// GossipState represents worker pool status shared in gossip
type GossipState struct {
	WorkerID  string
	QueueSize int
	InFlight  int
	Failures  int
	Timestamp time.Time
}

// GossipManager handles gossip between worker pools
type GossipManager struct {
	peers []*WorkerPool
	mu    sync.RWMutex
	stop  chan struct{}
}

// NewGossipManager creates a gossip manager
func NewGossipManager() *GossipManager {
	return &GossipManager{
		peers: []*WorkerPool{},
		stop:  make(chan struct{}),
	}
}

// AddPeer adds a worker pool as a gossip participant
func (gm *GossipManager) AddPeer(wp *WorkerPool) {
	gm.mu.Lock()
	defer gm.mu.Unlock()
	gm.peers = append(gm.peers, wp)
}

// Start begins gossip loop
func (gm *GossipManager) Start() {
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-gm.stop:
				return
			case <-ticker.C:
				gm.mu.RLock()
				if len(gm.peers) > 1 {
					from := gm.peers[rand.Intn(len(gm.peers))]
					to := gm.peers[rand.Intn(len(gm.peers))]
					if from != to {
						state := from.CollectState()
						to.HandleGossip(state)
					}
				}
				gm.mu.RUnlock()
			}
		}
	}()
}

// Stop ends gossiping
func (gm *GossipManager) Stop() {
	close(gm.stop)
}

// --- WorkerPool Gossip Hooks ---

// CollectState gathers current worker pool stats
func (wp *WorkerPool) CollectState() GossipState {
	return GossipState{
		WorkerID:  fmt.Sprintf("wp-%p", wp),
		QueueSize: len(wp.taskQueue),
		InFlight:  int(wp.metrics.TasksInFlight.Get()),
		Failures:  int(wp.metrics.TasksFailed.Get()),
		Timestamp: time.Now(),
	}
}

// HandleGossip processes gossip info
func (wp *WorkerPool) HandleGossip(state GossipState) {
	fmt.Printf("[gossip] %s heard: worker=%s queue=%d inFlight=%d failures=%d\n",
		time.Now().Format(time.RFC3339),
		state.WorkerID,
		state.QueueSize,
		state.InFlight,
		state.Failures,
	)
	// future: adaptively forward tasks to less loaded peers
}

// --- GossipCluster on top ---

type GossipCluster struct {
	manager   *GossipManager
	pipelines []*Pipeline
}

func NewGossipCluster(pipes ...*Pipeline) *GossipCluster {
	manager := NewGossipManager()
	// register all worker pools of each pipeline as peers
	for _, p := range pipes {
		for _, wp := range p.workerPools {
			manager.AddPeer(wp)
		}
	}
	manager.Start()

	return &GossipCluster{
		manager:   manager,
		pipelines: pipes,
	}
}

func (gc *GossipCluster) Submit(task *Task) error {
	idx := rand.Intn(len(gc.pipelines))
	return gc.pipelines[idx].Submit(task)
}

func (gc *GossipCluster) SubmitAndWait(task *Task) TaskResult {
	idx := rand.Intn(len(gc.pipelines))
	return gc.pipelines[idx].SubmitAndWait(task)
}

func (gc *GossipCluster) Stop() {
	gc.manager.Stop()
	for _, p := range gc.pipelines {
		p.Stop()
	}
}
