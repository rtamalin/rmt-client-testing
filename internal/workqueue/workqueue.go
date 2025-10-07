package workqueue

import (
	"fmt"
	"log/slog"
	"math"
	"strings"
	"sync"
	"time"
)

type TaskFunc func() error

type Job struct {
	Id        int64
	Name      string
	Task      TaskFunc
	CreatedAt time.Time
	StartedAt time.Time
	Duration  time.Duration
	Error     error
}

func NewJob(id int64, prefix string, task TaskFunc) *Job {
	j := new(Job)

	j.Init(id, prefix, task)

	return j
}

func (j *Job) setName(id int64, prefix string) {
	j.Id = id
	j.Name = fmt.Sprintf("%s_%08d", prefix, j.Id)
}

func (j *Job) Init(id int64, prefix string, task TaskFunc) {
	j.setName(id, prefix)
	j.Task = task
	j.CreatedAt = time.Now()
}

func (j *Job) Start() {
	j.StartedAt = time.Now()
}

func (j *Job) Finished() {
	j.Duration = time.Since(j.StartedAt)
}

func (j *Job) FinishedAt() time.Time {
	return j.StartedAt.Add(j.Duration)
}

type StatBlock struct {
	name    string
	unitSfx string
	count   int64
	min     float64
	max     float64
	mean    float64
	m2      float64
}

func NewStatBlock(name, unitSfx string) *StatBlock {
	s := new(StatBlock)
	s.Init(name, unitSfx)
	return s
}

func (s *StatBlock) Init(name, unitSfx string) {
	s.name = name
	s.unitSfx = unitSfx
	s.min = math.MaxFloat64
}

func (s *StatBlock) Update(value float64) {
	s.count++

	// update min if needed
	if value < s.min {
		s.min = value
	}

	// update max if needed
	if value > s.max {
		s.max = value
	}

	// using Welford's Online Algorithm
	// https://en.wikipedia.org/wiki/Algorithms_for_calculating_variance#Welford's_online_algorithm
	delta1 := value - s.mean
	s.mean += delta1 / float64(s.count)
	delta2 := value - s.mean
	s.m2 += delta1 * delta2
}

func (s *StatBlock) Name() string {
	return s.name
}

func (s *StatBlock) UnitSuffix() string {
	return s.unitSfx
}

func formatInt64(value int64, prefix, valueFmt, name, unit string) string {
	return fmt.Sprintf("%s%-16s "+valueFmt+" %s", prefix, name+":", value, unit)
}

func formatFloat64(value float64, prefix, valueFmt, name, unit string) string {
	return fmt.Sprintf("%s%-16s "+valueFmt+" %s", prefix, name+":", value, unit)
}

func (s *StatBlock) Summary(name string) string {
	// use default name if no override provided
	if name == "" {
		name = s.name
	}

	result := []string{
		fmt.Sprintf("%s Stats:", name),
		formatInt64(s.Count(), "  ", "%13d", "Total", ""),
		formatFloat64(s.Min(), "  ", "%13.6f", "Min", s.unitSfx),
		formatFloat64(s.Max(), "  ", "%13.6f", "Max", s.unitSfx),
		formatFloat64(s.Average(), "  ", "%13.6f", "Average", s.unitSfx),
		formatFloat64(s.Variance(), "  ", "%13.6f", "Variance", s.unitSfx),
		formatFloat64(s.StandardDeviation(), "  ", "%13.6f", "StdDev", s.unitSfx),
		formatFloat64(s.RootMeanSquare(), "  ", "%13.6f", "RMS", s.unitSfx),
	}

	return strings.Join(result, "\n")
}

func (s *StatBlock) Count() int64 {
	return s.count
}

func (s *StatBlock) Min() float64 {
	return s.min
}

func (s *StatBlock) Max() float64 {
	return s.max
}

func (s *StatBlock) Average() float64 {
	return s.mean
}

func (s *StatBlock) Variance() float64 {
	return s.m2 / float64(s.count)
}

func (s *StatBlock) SampleVariance() float64 {
	return s.m2 / float64(s.count-1)
}

func (s *StatBlock) StandardDeviation() float64 {
	return float64(math.Sqrt(s.Variance()))
}

func (s *StatBlock) SampleStandardDeviation() float64 {
	return float64(math.Sqrt(s.SampleVariance()))
}

func (s *StatBlock) RootMeanSquare() float64 {
	return math.Sqrt((s.m2 / float64(s.count)) + (s.mean * s.mean))
}

type WorkQueueStats struct {
	// private attributes
	jobStats  *StatBlock
	poolStats *StatBlock
}

func NewWorkQueueStats() *WorkQueueStats {
	s := new(WorkQueueStats)
	s.Init()
	return s
}

func (s *WorkQueueStats) Init() {
	// job durations will be converted to milliseconds
	s.jobStats = NewStatBlock("Job", "ms")

	// pool counts will be plain integers
	s.poolStats = NewStatBlock("Pool", "")
}

func (s *WorkQueueStats) JobStats() *StatBlock {
	return s.jobStats
}

func (s *WorkQueueStats) PoolStats() *StatBlock {
	return s.poolStats
}

func (s *WorkQueueStats) JobUpdate(job *Job) {
	s.jobStats.Update(float64(job.Duration.Milliseconds()))
}

func (s *WorkQueueStats) PoolUpdate(processedJobs int64) {
	s.poolStats.Update(float64(processedJobs))
}

type WorkQueue struct {
	// public attributes
	Stats      *WorkQueueStats
	StartTime  time.Time
	FinishTime time.Time
	Errors     []error

	// private attributes
	name         string
	numPools     int64
	jobs         chan *Job
	results      chan *Job
	pools        chan int64
	poolGroup    *sync.WaitGroup
	resultsGroup *sync.WaitGroup
}

func NewWorkQueue(name string, numPools int64) *WorkQueue {
	q := new(WorkQueue)

	q.name = name
	q.numPools = numPools

	q.Stats = NewWorkQueueStats()
	q.jobs = make(chan *Job)
	q.results = make(chan *Job)
	q.pools = make(chan int64, numPools)
	q.poolGroup = new(sync.WaitGroup)
	q.resultsGroup = new(sync.WaitGroup)

	return q
}

func (q *WorkQueue) poolHandler(id int64) {
	defer q.poolGroup.Done()

	slog.Debug(
		"Worker started",
		slog.Int64("id", id),
		slog.Time("start", time.Now()),
	)

	var processedJobs int64 = 0
	for job := range q.jobs {
		job.Start()
		err := job.Task()
		job.Finished()

		// if the job failed, updated the error to include the job name
		if err != nil {
			err = fmt.Errorf("job %q failed: %w", job.Name, err)
		}
		job.Error = err

		// submit the results
		q.results <- job

		// increment the processed jobs count
		processedJobs++
	}

	slog.Debug(
		"Worker finished",
		slog.Int64("id", id),
		slog.Time("finish", time.Now()),
		slog.Int64("processedJobs", processedJobs),
	)

	q.pools <- processedJobs
}

func (q *WorkQueue) startPoolHandlers() {
	var i int64
	for i = 0; i < q.numPools; i++ {
		q.poolGroup.Add(1)
		go q.poolHandler(i)
	}
}

func (q *WorkQueue) jobResultsHandler() {
	defer q.resultsGroup.Done()

	for job := range q.results {
		if job.Error != nil {
			q.Errors = append(q.Errors, job.Error)
		}
		q.Stats.JobUpdate(job)
	}
}

func (q *WorkQueue) poolResultsHandler() {
	defer q.resultsGroup.Done()

	for processedJobs := range q.pools {
		q.Stats.PoolUpdate(processedJobs)
	}
}

func (q *WorkQueue) startResultsHandlers() {
	// job results handler
	q.resultsGroup.Add(1)
	go q.jobResultsHandler()

	// pool results handler
	q.resultsGroup.Add(1)
	go q.poolResultsHandler()
}

func (q *WorkQueue) NewJob(id int64, task TaskFunc) *Job {
	return NewJob(id, q.name, task)
}

func (q *WorkQueue) Start() {
	q.startPoolHandlers()
	q.startResultsHandlers()
}

func (q *WorkQueue) Add(job *Job) {
	q.jobs <- job
}

func (q *WorkQueue) WaitForCompletion() {
	close(q.jobs)
	q.poolGroup.Wait()

	close(q.results)
	close(q.pools)
	q.resultsGroup.Wait()
}
