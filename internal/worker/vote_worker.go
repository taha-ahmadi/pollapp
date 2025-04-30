package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"pollapp/internal/adapter/broker"
	"pollapp/internal/repository/model"
	"pollapp/internal/repository/postgresql/pollrepo"
	"pollapp/internal/service/domain"
	"sync"
	"time"
)

const (
	VoteQueueName = "vote_queue"
)

type VoteWorker struct {
	broker      broker.MessageBroker
	pollRepo    pollrepo.IPollRepository
	queueName   string
	workerCount int
	shutdownCh  chan struct{}
	wg          sync.WaitGroup
}

func NewVoteWorker(broker broker.MessageBroker, pollRepo pollrepo.IPollRepository, queueName string, workerCount int) *VoteWorker {
	return &VoteWorker{
		broker:      broker,
		pollRepo:    pollRepo,
		queueName:   queueName,
		workerCount: workerCount,
		shutdownCh:  make(chan struct{}),
	}
}

func (w *VoteWorker) Start(ctx context.Context) error {
	msgCh, err := w.broker.Consume(w.queueName)
	if err != nil {
		return fmt.Errorf("failed to start consuming from queue %s: %w", w.queueName, err)
	}

	for i := 0; i < w.workerCount; i++ {
		w.wg.Add(1)
		go w.processMessages(ctx, msgCh, i)
	}

	log.Printf("Started %d vote workers", w.workerCount)
	return nil
}

func (w *VoteWorker) Stop() {
	close(w.shutdownCh)
	w.wg.Wait()
	log.Println("Vote workers stopped")
}

func (w *VoteWorker) processMessages(ctx context.Context, msgCh <-chan broker.Message, workerID int) {
	defer w.wg.Done()

	for {
		select {
		case <-w.shutdownCh:
			log.Printf("Worker %d shutting down", workerID)
			return
		case <-ctx.Done():
			log.Printf("Worker %d context canceled", workerID)
			return
		case msg, ok := <-msgCh:
			if !ok {
				log.Printf("Worker %d message channel closed", workerID)
				return
			}

			w.handleMessage(ctx, msg, workerID)
		}
	}
}

func (w *VoteWorker) handleMessage(ctx context.Context, msg broker.Message, workerID int) {
	startTime := time.Now()
	log.Printf("Worker %d processing message: %s", workerID, msg.ID)

	var voteJob domain.VoteJob
	err := json.Unmarshal(msg.Body, &voteJob)
	if err != nil {
		log.Printf("Worker %d failed to unmarshal vote job: %v", workerID, err)
		w.broker.Nack(msg.DeliveryTag, false)
		return
	}

	vote := model.Vote{
		PollID:      int(voteJob.PollID),
		UserID:      int(voteJob.UserID),
		OptionIndex: voteJob.OptionIndex,
		CreatedAt:   voteJob.CreatedAt,
	}

	err = w.pollRepo.Vote(ctx, voteJob.PollID, vote)
	if err != nil {
		log.Printf("Worker %d failed to process vote: %v", workerID, err)
		w.broker.Nack(msg.DeliveryTag, true)
		return
	}

	err = w.broker.Ack(msg.DeliveryTag)
	if err != nil {
		log.Printf("Worker %d failed to acknowledge message: %v", workerID, err)
	}

	processingTime := time.Since(startTime)
	log.Printf("Worker %d completed message %s in %v", workerID, msg.ID, processingTime)
}
