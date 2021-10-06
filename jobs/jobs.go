package jobs

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/flow-hydraulics/flow-wallet-api/errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WorkerPool struct {
	logger      *log.Logger
	wg          *sync.WaitGroup
	store       Store
	jobChan     chan *Job
	capacity    uint
	workerCount uint
}

type Result struct {
	Result        string
	TransactionID string
}

type Process func(result *Result) error

// State is a type for Job state.
type State string

const (
	Init               State = "INIT"
	Accepted           State = "ACCEPTED"
	NoAvailableWorkers State = "NO_AVAILABLE_WORKERS"
	Error              State = "ERROR"
	Complete           State = "COMPLETE"
	Failed             State = "FAILED"
)

// Job database model
type Job struct {
	ID            uuid.UUID      `gorm:"column:id;primary_key;type:uuid;"`
	Type          string         `gorm:"column:type"`
	State         State          `gorm:"column:state;default:INIT"`
	Error         string         `gorm:"column:error"`
	Result        string         `gorm:"column:result"`
	TransactionID string         `gorm:"column:transaction_id"`
	RetryCount    int            `gorm:"column:retry_count;default:0"`
	CreatedAt     time.Time      `gorm:"column:created_at"`
	UpdatedAt     time.Time      `gorm:"column:updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"column:deleted_at;index"`
	Do            Process        `gorm:"-"`
}

// Job HTTP response
type JSONResponse struct {
	ID            uuid.UUID `json:"jobId"`
	State         State     `json:"state"`
	Error         string    `json:"error"`
	Result        string    `json:"result"`
	TransactionID string    `json:"transactionId"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

func (j Job) ToJSONResponse() JSONResponse {
	return JSONResponse{
		ID:            j.ID,
		State:         j.State,
		Error:         j.Error,
		Result:        j.Result,
		TransactionID: j.TransactionID,
		CreatedAt:     j.CreatedAt,
		UpdatedAt:     j.UpdatedAt,
	}
}

func (j *Job) BeforeCreate(tx *gorm.DB) (err error) {
	j.ID = uuid.New()
	return nil
}

func (j *Job) Wait(wait bool) error {
	if wait {
		// Wait for the job to have finished
		for j.State == Accepted {
			time.Sleep(10 * time.Millisecond)
		}
		if j.State == Error {
			return fmt.Errorf(j.Error)
		}
	}
	return nil
}

func NewWorkerPool(logger *log.Logger, db Store, capacity uint, workerCount uint) *WorkerPool {
	if logger == nil {
		// Make sure we always have a logger
		logger = log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)
	}
	wg := &sync.WaitGroup{}
	jobChan := make(chan *Job, capacity)
	pool := &WorkerPool{logger, wg, db, jobChan, capacity, workerCount}
	pool.startWorkers()
	return pool
}

func (wp *WorkerPool) startWorkers() {
	for i := uint(0); i < wp.workerCount; i++ {
		wp.wg.Add(1)
		go func() {
			defer wp.wg.Done()
			for job := range wp.jobChan {
				if job == nil {
					break
				}
				wp.process(job)
			}
		}()
	}
}

// AddJob will try to add a job to the workerpool
func (wp *WorkerPool) AddJob(do Process) (*Job, error) {
	// Init job
	job := &Job{Do: do, State: Init}

	// Insert job into database
	if err := wp.store.InsertJob(job); err != nil {
		return job, err
	}

	// Trying queueing the job
	if !wp.tryEnqueue(job) {
		job.State = NoAvailableWorkers

		// Update database
		if err := wp.store.UpdateJob(job); err != nil {
			wp.logger.Println("WARNING: Could not update DB entry for Job", job.ID)
		}

		return job, &errors.JobQueueFull{Err: fmt.Errorf("%s", job.State)}
	}

	job.State = Accepted

	// Update database
	if err := wp.store.UpdateJob(job); err != nil {
		wp.logger.Println("WARNING: Could not update DB entry for Job", job.ID)
	}

	return job, nil
}

func (wp *WorkerPool) Stop() {
	close(wp.jobChan)
	wp.wg.Wait()
}

func (wp *WorkerPool) tryEnqueue(job *Job) bool {
	select {
	case wp.jobChan <- job:
		return true
	default:
		return false
	}
}

func (wp *WorkerPool) process(job *Job) {
	result := &Result{}
	err := job.Do(result)

	job.Result = result.Result
	job.TransactionID = result.TransactionID

	if err != nil {
		wp.logger.Printf("[Job %s] Error while processing job: %s\n", job.ID, err)
		job.State = Error
		job.Error = err.Error()
		if err := wp.store.UpdateJob(job); err != nil {
			wp.logger.Println("WARNING: Could not update DB entry for Job", job.ID)
		}
		return
	}

	job.State = Complete
	if err := wp.store.UpdateJob(job); err != nil {
		wp.logger.Println("WARNING: Could not update DB entry for Job", job.ID)
	}
}
