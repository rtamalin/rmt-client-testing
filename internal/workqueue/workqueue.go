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
	Id         int64
	Name       string
	Task       TaskFunc
	CreatedAt  time.Time
	StartedAt  time.Time
	FinishedAt time.Time
	Error      error
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

func (j *Job) Finish() {
	j.FinishedAt = time.Now()
}

func (j *Job) Duration() time.Duration {
	return j.FinishedAt.Sub(j.StartedAt)
}

type StatBlock struct {
	// initialised
	name    string
	unitSfx string

	// calculated
	count int64
	min   float64
	max   float64
	mean  float64
	m2    float64
	start time.Time
	end   time.Time
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

func (s *StatBlock) Update(sample float64, start, end time.Time) {
	s.count++

	// update start and end timers
	if !start.IsZero() && (start.Before(s.start) || s.start.IsZero()) {
		s.start = start
	}

	if !end.IsZero() && (end.After(s.end) || s.end.IsZero()) {
		s.end = end
	}

	// update min if needed
	if sample < s.min {
		s.min = sample
	}

	// update max if needed
	if sample > s.max {
		s.max = sample
	}

	// using Welford's Online Algorithm
	// https://en.wikipedia.org/wiki/Algorithms_for_calculating_variance#Welford's_online_algorithm
	delta1 := sample - s.mean
	s.mean += delta1 / float64(s.count)
	delta2 := sample - s.mean
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

type SummaryOpts map[string]any

const (
	OPT_HEADER      = "header"
	OPT_FOOTER      = "footer"
	OPT_NAME        = "name"
	OPT_RATE        = "rate"
	OPT_MIN_MAX     = "min_max"
	OPT_EXTRA_STATS = "extra_stats"
)

func DefaultSummaryOpts() SummaryOpts {
	return SummaryOpts{
		OPT_MIN_MAX:     true,
		OPT_EXTRA_STATS: true,
	}
}

const (
	INT64_FMT = "%13d"
	FLT64_FMT = "%13.6f"
)

func (s *StatBlock) Summary(opts SummaryOpts) string {
	result := []string{}

	// if provided, start with the header value
	if value, found := opts[OPT_HEADER]; found {
		result = append(result,
			value.(string),
		)
	}

	// use default name if no override provided
	name := s.name
	if value, found := opts[OPT_NAME]; found {
		name = value.(string)
	}

	// common initial entries
	result = append(result,
		fmt.Sprintf("%s Stats:", name),
		formatInt64(s.Count(), "  ", INT64_FMT, "Total", ""),
	)

	// if requested, include elapsed and rate time
	if _, found := opts[OPT_RATE]; found {
		if !s.start.IsZero() {
			result = append(result,
				formatFloat64(s.Elapsed(), "  ", FLT64_FMT, "Elapsed", "s"),
				formatFloat64(s.Rate(), "  ", FLT64_FMT, "Rate", "/s"),
			)
		}
	}

	// append standard entries
	result = append(result,
		formatFloat64(s.Average(), "  ", FLT64_FMT, "Average", s.unitSfx),
	)

	// if requested, include min & max
	if _, found := opts[OPT_MIN_MAX]; found {
		result = append(result,
			formatFloat64(s.Min(), "  ", FLT64_FMT, "Min", s.unitSfx),
			formatFloat64(s.Max(), "  ", FLT64_FMT, "Max", s.unitSfx),
		)
	}

	// if requested, include extra_stats
	if _, found := opts[OPT_EXTRA_STATS]; found {
		result = append(result,
			formatFloat64(s.Variance(), "  ", FLT64_FMT, "Variance", s.unitSfx),
			formatFloat64(s.StandardDeviation(), "  ", FLT64_FMT, "StdDev", s.unitSfx),
			formatFloat64(s.RootMeanSquare(), "  ", FLT64_FMT, "RMS", s.unitSfx),
		)
	}

	// if provided, finish with the footer value
	if value, found := opts[OPT_FOOTER]; found {
		result = append(result,
			value.(string),
		)
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

func (s *StatBlock) Elapsed() float64 {
	return s.end.Sub(s.start).Seconds()
}

func (s *StatBlock) Rate() float64 {
	elapsed := s.Elapsed()
	// return 0 if no elapsed time is available
	if elapsed == 0 {
		return 0
	}
	return float64(s.count) / elapsed
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
	s.jobStats = NewStatBlock("Job", "s")

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
	s.jobStats.Update(
		job.Duration().Seconds(),
		job.StartedAt,
		job.FinishedAt,
	)
}

func (s *WorkQueueStats) PoolUpdate(processedJobs int64) {
	s.poolStats.Update(
		float64(processedJobs),
		time.Time{},
		time.Time{},
	)
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
		job.Finish()

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
	if q.StartTime.IsZero() {
		q.StartTime = time.Now()
	}
	q.jobs <- job
}

func (q *WorkQueue) WaitForCompletion() {
	close(q.jobs)
	q.poolGroup.Wait()
	q.FinishTime = time.Now()

	close(q.results)
	close(q.pools)
	q.resultsGroup.Wait()
}
