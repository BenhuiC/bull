package template

func Logger() []byte {
	return []byte(`
package log

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger = zap.SugaredLogger

var level = zap.InfoLevel

func init() {
	if os.Getenv("DEBUG") == "1" {
		level = zap.DebugLevel
	}
}

func NewLogger(opts ...zap.Option) *Logger {
	c := zap.NewProductionConfig()
	// c := zap.NewDevelopmentConfig()
	c.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	c.Level = zap.NewAtomicLevelAt(level)
	c.Encoding = "console"
	c.DisableStacktrace = true
	l, err := c.Build(opts...)
	if err != nil {
		panic(err)
	}
	return l.Sugar()
}
`)
}

func ExtendLogger() []byte {
	return []byte(`
package log

import "go.uber.org/zap"

type ExtendLogger struct {
	zap.SugaredLogger
}

func (l *ExtendLogger) Printf(format string, args ...interface{}) {
	l.Infof(format, args...)
}
`)
}

func Util() []byte {
	return []byte(`
package util

// todo
`)
}

func WorkerCtx() []byte {
	return []byte(`
package worker

import "context"

type Context interface {
	context.Context
	Meta() *Meta
}

type ContextKey string

const (
	ContextKeyMeta ContextKey = "meta"
)

type workContext struct {
	context.Context
}

func NewContext(ctx context.Context, m *Meta) Context {
	ctx = context.WithValue(ctx, ContextKeyMeta, m)
	return &workContext{ctx}
}

func (ctx workContext) Meta() *Meta {
	m, _ := ctx.Value(ContextKeyMeta).(*Meta)
	return m
}
`)
}

func WorkerMeta() []byte {
	return []byte(`
package worker

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Queue = string

const (
	QueueHigh Queue = "High"
	QueueLow  Queue = "Low"

	DefaultRetryCount = 13
)

type Meta struct {
	ID        string
	Name      string
	PerformAt *time.Time
	Retry     int
	Queue     Queue

	CreatedAt  time.Time
	RetryCount int
	Success    bool
	Error      string
	Raw        []byte
}

func NewMetaByWorker(w Worker, opts ...Option) (*Meta, error) {
	raw, err := json.Marshal(w)
	if err != nil {
		return nil, err
	}
	m := Meta{
		ID:    uuid.NewString(),
		Name:  w.WorkerName(),
		Raw:   raw,
		Retry: DefaultRetryCount,
		Queue: QueueLow,

		CreatedAt: time.Now(),
	}
	for _, opt := range opts {
		opt(&m)
	}
	return &m, nil
}

type Option func(c *Meta)

// retry < 0 means retry count unlimits
// default WorkerDefaultRetryCount
func WithRetry(retry int) Option {
	return func(c *Meta) {
		c.Retry = retry
	}
}

// set worker execute time
func WithPerformAt(performAt time.Time) Option {
	return func(c *Meta) {
		c.PerformAt = &performAt
	}
}

// set worker Queue
// default low
func WithQueue(q Queue) Option {
	return func(c *Meta) {
		c.Queue = q
	}
}
`)
}

func WorkerRunner() []byte {
	return []byte(`
package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strings"
	"sync"
	"time"

	"{{ .ProjectName}}/pkg/log"
	"{{ .ProjectName}}/pkg/util"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	antsv2 "github.com/panjf2000/ants/v2"
	"go.uber.org/zap"
)

const (
	RunnerAliveStatusTTL = 30 // runner????????????????????????
	RedisTimeout         = time.Second * 3

	Prefix = "worker:"

	KeyWorkers           = Prefix + "workers" // ??????work??????
	KeyRunnerAlivePrefix = Prefix + "alive#"  // ??????runner????????????
	KeyWaitingQueue      = Prefix + "waiting" // ????????????
	ReadyQueueLockTerm   = 60 * time.Second   // ????????????????????????

	KeyReadyQueueLocker   = Prefix + "readyQueueLocker" // ???????????????
	KeyWaitingQueueLocker = Prefix + "waitingLocker"    // ???????????????
	KeyWorkingCheckLocker = Prefix + "workingLocker"    // ???????????????????????????

	KeyWorking                      = Prefix + "working" // ????????????
	WorkingCheckLockerTerm          = 6 * time.Minute    // ???????????????????????????????????????
	WaitingQueueCatchMissingWaiting = 30 * time.Second   // ????????????????????????????????????????????????
	WaitingQueueCatchEmptyWaiting   = 1 * time.Second    // ??????????????????
	WaitingQueueLockTerm            = 60 * time.Second   // ????????????????????????
	WaitingQueueCatchBatchSize      = 100                // ??????????????????????????????
	WaitingQueueDataIDSeparator     = "#"                // ????????????????????????????????????ID????????????????????????

	ReadyQueuePullBatchSize = 30 // ??????????????????????????????
	NeedPullThresholdRatio  = 3  // ????????????????????????NeedPullThresholdRatio * Threads ????????????????????????????????????
)

var (
	ErrWorkerNotRegistry = errors.New("unregistry worker")
	logger               = log.NewLogger()

	KeyReadyQueueHigh = QueueKey(QueueHigh) // ????????????????????????
	KeyReadyQueueLow  = QueueKey(QueueLow)  // ????????????????????????
)

// ??????????????????????????????????????????????????????????????????, ????????????????????????????????????????????????
// ??????????????????????????????worker????????????????????????????????????????????????????????????
// ???????????????????????????worker???????????????????????????????????????????????????????????????????????????
// ????????????????????????????????????????????????????????????worker??????????????????????????????????????????,??????????????????????????????????????????????????????????????????????????????????????????
// ????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????
// ??????????????????????????????????????????????????????
type RedisRunner struct {
	ID              string
	redisCli        *redis.Client
	RegistryWorkers map[string]reflect.Type

	threads uint
	wg      sync.WaitGroup

	// ????????????
	execChan chan *Meta

	// ??????????????????
	execResult chan *Meta

	// ???????????????
	execPool *antsv2.PoolWithFunc

	// ????????????????????????????????????pull??????
	needPull chan bool

	// ??????????????????worker??????
	batchPull chan int

	status *RunnerStatus
}

func NewRunner(redisCli *redis.Client, threads uint) (*RedisRunner, error) {
	r := RedisRunner{
		ID:              uuid.NewString() + time.Now().Format("#2006-01-02T15:04:05"),
		redisCli:        redisCli,
		RegistryWorkers: make(map[string]reflect.Type),
		wg:              sync.WaitGroup{},
		threads:         threads,
		execChan:        make(chan *Meta),
		execResult:      make(chan *Meta),
		needPull:        make(chan bool),
		batchPull:       make(chan int, 1),
		status:          new(RunnerStatus),
	}
	var err error
	// ?????????1??????????????????ants????????????
	r.execPool, err = antsv2.NewPoolWithFunc(
		int(threads),
		r.newExecWorkerFunc(),
		antsv2.WithExpiryDuration(time.Second*10), // ????????????????????????
		antsv2.WithLogger(&log.ExtendLogger{SugaredLogger: *logger}),
	)
	return &r, err
}

// Declare should used before worker Registry
func (r *RedisRunner) Declare(work Worker, opts ...Option) (*Meta, error) {
	c, err := NewMetaByWorker(work, opts...)
	if err != nil {
		return nil, err
	}
	return c, r.doSubmit(c)
}

// worker should registry before worker loop lanch
func (r *RedisRunner) RegistryWorker(work Worker) error {
	if _, exist := r.RegistryWorkers[work.WorkerName()]; exist {
		return fmt.Errorf("worker %s has already registry", work.WorkerName())
	}
	r.RegistryWorkers[work.WorkerName()] = reflect.TypeOf(work).Elem()
	return nil
}

func (r *RedisRunner) Run(ctx context.Context) error {
	// ????????????????????????????????????worker, ???????????????worker????????????
	r.checkWorkingWorkers(ctx)

	r.wg.Add(5)
	// ????????????runner????????????
	go r.startRunnerAlive(ctx)
	// ???????????????????????????????????????????????????????????????
	go r.startLoopTransWaitingQueue(ctx)
	// ?????????????????????????????????
	go r.startLoopPullWorker(ctx)
	// ????????????
	go r.startLoopExecWorker(ctx)
	// ?????????????????????????????????
	go r.startLoopCollect(ctx)

	logger.Infof("workerRunner start %v", r.ID)
	<-ctx.Done()
	r.wg.Wait()
	return nil
}

func (r *RedisRunner) startRunnerAlive(ctx context.Context) {
	defer r.wg.Done()
	tk := time.NewTicker(time.Second * RunnerAliveStatusTTL / 3)
	defer tk.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tk.C:
			r.doSetAlive()
		}
	}
}

// ???????????????
func (r *RedisRunner) checkWorkingWorkers(ctx context.Context) {
	unlocker, err := util.RedisLockV(r.redisCli, KeyWorkingCheckLocker, r.ID, time.Second*10)
	if err != nil {
		if err == util.ErrLockerAlreadySet {
			logger.Info("checkWorkingWorkers already set")
			return
		}
	}
	defer unlocker()
	workers, err := r.getAllWorkingWorkers()
	if err != nil {
		logger.Errorf("checkWorkingWorkers error %v", err)
		return
	}
	cache := map[string]bool{r.ID: true}
	for workerID, runnerID := range workers {
		alive, ok := cache[runnerID]
		if !ok {
			alive, err = r.checkRunnerAlive(runnerID)
			if err != nil {
				logger.Errorf("checkWorkingWorkers: %v", err)
				continue
			}
			cache[runnerID] = alive
		}
		if alive {
			continue
		}
		logger.Warn("recover worker", zap.String("worker_id", workerID))
		if err := r.recoverWorker(workerID); err != nil {
			logger.Errorf("checkWorkingWorkers: %v", err)
		}
	}
	logger.Debug("workers checked")
}

func (r *RedisRunner) startLoopTransWaitingQueue(ctx context.Context) {
	defer r.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			r.transWaitingWorkers(ctx)
		}
	}
}

func (r *RedisRunner) transWaitingWorkers(ctx context.Context) {
	defer func() {
		if err := recover(); err != nil {
			logger.Errorf("transWaitingWorkers panic: %v", err)
			time.Sleep(WaitingQueueCatchEmptyWaiting)
		}
	}()
	// ??????WaitingQueue?????????
	unlocker, err := util.RedisLockV(r.redisCli, KeyWaitingQueueLocker, r.ID, WaitingQueueLockTerm)
	if err != nil {
		if err == util.ErrLockerAlreadySet {
			time.Sleep(WaitingQueueCatchMissingWaiting)
			return
		}
		logger.Errorf("transWaitingWorkers: %v", err)
		time.Sleep(WaitingQueueCatchEmptyWaiting)
		return
	}
	defer unlocker()

	now := time.Now().Unix()
	ws, err := r.loadWaitingWorkers(now)
	if err != nil {
		logger.Errorf("transWaitingWorkers: %v", err)
		time.Sleep(WaitingQueueCatchEmptyWaiting)
		return
	}
	if len(ws) > 0 {
		err = r.transWaitingToReady(ws)
		if err != nil {
			logger.Errorf("transWaitingWorkers: %v", err)
			time.Sleep(WaitingQueueCatchEmptyWaiting)
			return
		}
	} else {
		logger.Debug("transWaitingWorkers: no worker")
		time.Sleep(WaitingQueueCatchEmptyWaiting)
	}
}

func (r *RedisRunner) startLoopPullWorker(ctx context.Context) {
	defer r.wg.Done()
	for {
		ws, err := r.loadReadyWorkers()
		if err != nil {
			logger.Errorf("loadReadyWorkers: %v", err)
			time.Sleep(time.Second)
			continue
		}
		if len(ws) > 0 {
			r.toExec(ws)
		}

		select {
		case <-ctx.Done():
			close(r.execChan)
			return
		case <-r.needPull:
			continue
		}
	}
}

// ?????????????????????????????????????????????????????????
func (r *RedisRunner) loadReadyWorkers() (ws []string, err error) {
	unlocker, err := util.RedisLockV(r.redisCli, KeyReadyQueueLocker, r.ID, ReadyQueueLockTerm)
	if err != nil {
		if errors.Is(err, util.ErrLockerAlreadySet) {
			logger.Debug("loadReadyWorkers conflict")
			return nil, nil
		}
		return
	}
	defer unlocker()

	// get should pull len
	highCount, lowCount, err := r.shouldBachPullReadyCount()
	if err != nil {
		return
	}
	if highCount == 0 && lowCount == 0 {
		return
	}

	var w []string
	if w, err = r.loadAndTransReadyToWorking(KeyReadyQueueHigh, highCount); err != nil {
		return
	}
	ws = append(ws, w...)
	if w, err = r.loadAndTransReadyToWorking(KeyReadyQueueLow, lowCount); err != nil {
		return
	}
	ws = append(ws, w...)
	r.batchPull <- len(ws)
	return ws, nil
}

func (r *RedisRunner) loadAndTransReadyToWorking(queue string, count int64) (ws []string, err error) {
	if count == 0 {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), RedisTimeout)
	defer cancel()
	ws, err = r.redisCli.LRange(ctx, queue, 0, count-1).Result()
	if err != nil {
		return
	}

	// save to working
	ctx, cancel = context.WithTimeout(context.Background(), RedisTimeout)
	defer cancel()
	workingVal := map[string]interface{}{}
	for _, workerID := range ws {
		workingVal[workerID] = r.ID
	}
	_, err = r.redisCli.HSet(ctx, KeyWorking, workingVal).Result()
	if err != nil {
		return
	}

	// remove from ready
	ctx, cancel = context.WithTimeout(context.Background(), RedisTimeout)
	defer cancel()
	_, err = r.redisCli.LTrim(ctx, queue, count, -1).Result()
	return
}

func (r *RedisRunner) toExec(ws []string) {
	// load works content
	var faildCount int
	defer func() {
		if faildCount > 0 {
			r.batchPull <- -1 * faildCount
		}
	}()
	ctx, cancel := context.WithTimeout(context.Background(), RedisTimeout)
	defer cancel()
	pip := r.redisCli.Pipeline()
	workContents := []string{}
	for _, workerID := range ws {
		pip.HGet(ctx, KeyWorkers, workerID)
	}
	rs, err := pip.Exec(ctx)
	if err != nil {
		logger.Errorf("toExec: %v", err)
		faildCount = len(ws)
		return
	}
	for _, r := range rs {
		if r.Err() != nil {
			logger.Errorf("toExec: %v", r.Err())
			faildCount++
			continue
		}
		workContents = append(workContents, r.(*redis.StringCmd).Val())
	}

	// execute
	for _, workContent := range workContents {
		w := &Meta{}
		err := json.Unmarshal([]byte(workContent), &w)
		if err != nil {
			logger.Errorf("transReadyToWorking: %v", err)
			faildCount++
			continue
		}
		r.execChan <- w
	}
}

// ???execChan??????worker????????????
func (r *RedisRunner) startLoopExecWorker(ctx context.Context) {
	defer r.wg.Done()
	for wc := range r.execChan {
		r.execPool.Invoke(wc)
	}
	defer r.execPool.Release()
	t := time.After(time.Second * 30)
	for {
		select {
		case <-t:
			return
		default:
			if r.execPool.Running() > 0 {
				time.Sleep(time.Second)
			} else {
				return
			}
		}
	}
}

// ????????????????????????????????????????????????
func (r *RedisRunner) startLoopCollect(ctx context.Context) {
	defer r.wg.Done()

	var left int
	var threshold = int(r.threads * NeedPullThresholdRatio)

	status := r.status

	notice := time.NewTimer(time.Second * 3)
	defer notice.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-notice.C:
			if left <= threshold {
				select {
				case r.needPull <- true:
				default:
				}
			}
			logger.Debugf("execCount: %d, failCount: %d, left: %d", status.ExecCount, status.FailCount, left)
			notice.Reset(time.Second)
		case ws := <-r.execResult:
			left--
			if left <= threshold {
				select {
				case r.needPull <- true:
				default:
				}
			}

			// ????????????
			status.ExecCount++
			if !ws.Success {
				status.FailCount++
			}
		case count := <-r.batchPull:
			left += count
		}
	}
}

func (r *RedisRunner) shouldBachPullReadyCount() (queueHighCount, queueLowCount int64, err error) {
	queueHighLen, err := r.GetQueueLen(KeyReadyQueueHigh)
	if err != nil {
		return
	}
	queueLowLen, err := r.GetQueueLen(KeyReadyQueueLow)
	if err != nil {
		return
	}

	queueLowCount = ReadyQueuePullBatchSize / 3
	if queueLowLen < queueLowCount {
		queueLowCount = queueLowLen
	}
	queueHighCount = ReadyQueuePullBatchSize - queueLowCount
	if queueHighLen < queueHighCount {
		queueHighCount = queueHighLen
	}
	if queueHighCount+queueLowCount < ReadyQueuePullBatchSize && queueLowLen > queueLowCount {
		queueLowCount = ReadyQueuePullBatchSize - queueHighCount
		if queueLowLen < queueLowCount {
			queueLowCount = queueLowLen
		}
	}
	return
}

// ??????worker
func (r *RedisRunner) recoverWorker(workerID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), RedisTimeout)
	defer cancel()
	_, err := r.redisCli.RPush(ctx, KeyReadyQueueLow, workerID).Result()
	if err != nil {
		return err
	}
	_, err = r.redisCli.HDel(ctx, KeyWorking, workerID).Result()
	if err != nil {
		return err
	}
	return nil
}

func (r *RedisRunner) retryWorker(wc *Meta) {
	wc.RetryCount++
	delay := wc.RetryCount
	if delay > 10 {
		delay = 10
	}
	delay = int(math.Pow(2, float64(delay)))
	retryAt := time.Now().Add(time.Duration(delay) * time.Minute)
	wc.PerformAt = &retryAt
	var err error
	if err = r.doSubmit(wc); err == nil {
		ctx, cancel := context.WithTimeout(context.Background(), RedisTimeout)
		defer cancel()
		r.redisCli.HDel(ctx, KeyWorking, wc.ID)
	}
	if err != nil {
		withWorkerLogger(wc).Errorf("unable to remove worker, err: %v", err)
	}
}

func (r *RedisRunner) removeWorker(wc *Meta) {
	ctx, cancel := context.WithTimeout(context.Background(), RedisTimeout)
	defer cancel()
	if err := r.redisCli.HDel(ctx, KeyWorking, wc.ID).Err(); err != nil {
		withWorkerLogger(wc).Errorf("unable to remove worker, err: %v", err)
	}
	if err := r.redisCli.HDel(ctx, KeyWorkers, wc.ID).Err(); err != nil {
		withWorkerLogger(wc).Errorf("unable to remove worker, err: %v", err)
	}
}

func (r *RedisRunner) newExecWorkerFunc() func(item interface{}) {
	return func(item interface{}) {
		wc := item.(*Meta)
		l := withWorkerLogger(wc)
		l.Info("START")

		// ??????????????????
		defer func() {
			if e := recover(); e != nil {
				wc.Error = fmt.Sprintf("%v", e)
				l.Errorf("panic: %s", wc.Error)
			}
			r.execResult <- wc
			if !wc.Success && wc.RetryCount < wc.Retry {
				r.retryWorker(wc)
			} else {
				if !wc.Success {
					withWorkerLogger(wc).Warn("retry times over, remove")
				}
				r.removeWorker(wc)
			}
		}()

		// ??????worker??????
		wt, ok := r.RegistryWorkers[wc.Name]
		if !ok {
			l.Errorf("unregistry worker: %s", wc.Name)
			wc.Error = "unknow worker type"
			return
		}
		worker := reflect.New(wt).Interface().(Worker)
		err := json.Unmarshal(wc.Raw, worker)
		if err != nil {
			l.Errorf("unable to unmarshal worker, err: %v", err)
			wc.Error = err.Error()
			return
		}

		// ??????worker
		ctx := NewContext(context.Background(), wc)
		if err = worker.Perform(ctx, l); err != nil {
			l.Errorf("perform worker err: %v", err)
			wc.Error = err.Error()
			return
		}
		wc.Success = true
		l.Info("DONE")
	}
}

func (r *RedisRunner) doSubmit(c *Meta) error {
	raw, err := json.Marshal(c)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), RedisTimeout)
	defer cancel()
	pip := r.redisCli.TxPipeline()

	pip.HSet(ctx, KeyWorkers, c.ID, raw)
	queue := QueueKey(c.Queue)
	if c.PerformAt != nil && c.PerformAt.After(time.Now()) {
		val := queue + WaitingQueueDataIDSeparator + c.ID
		pip.ZAdd(ctx, KeyWaitingQueue, &redis.Z{
			Score:  float64(c.PerformAt.Unix()),
			Member: val,
		})
	} else {
		pip.RPush(ctx, queue, c.ID).Err()
	}
	if _, err = pip.Exec(ctx); err != nil {
		return err
	}
	return nil
}

func (r *RedisRunner) doSetAlive() {
	c, cancel := context.WithTimeout(context.Background(), RedisTimeout)
	defer cancel()
	err := r.redisCli.Set(c, KeyRunnerAlivePrefix+r.ID, "alive", RunnerAliveStatusTTL*time.Second).Err()
	if err != nil {
		logger.Errorf("DoSetAlive: %v", err)
	}
}

func (r *RedisRunner) getAllWorkingWorkers() (workers map[string]string, err error) {
	c, cancel := context.WithTimeout(context.Background(), RedisTimeout)
	defer cancel()
	workers, err = r.redisCli.HGetAll(c, KeyWorking).Result()
	if err != nil {
		return nil, fmt.Errorf("GetAllWorkingWorkers: %v", err)
	}
	return
}

func (r *RedisRunner) checkRunnerAlive(runnerID string) (bool, error) {
	c, cancel := context.WithTimeout(context.Background(), RedisTimeout)
	defer cancel()
	val, err := r.redisCli.Get(c, KeyRunnerAlivePrefix+runnerID).Result()
	if err != nil {
		return false, fmt.Errorf("CheckRunnerAlive: %v", err)
	}
	if val == "alive" {
		return true, nil
	}
	return false, nil
}

func (r *RedisRunner) loadWaitingWorkers(endAt int64) (ws []string, err error) {
	c, cancel := context.WithTimeout(context.Background(), RedisTimeout)
	defer cancel()
	ws, err = r.redisCli.ZRangeByScore(c, KeyWaitingQueue, &redis.ZRangeBy{
		Min:    "0",
		Max:    fmt.Sprintf("%d", endAt),
		Offset: 0,
		Count:  WaitingQueueCatchBatchSize,
	}).Result()
	if err != nil {
		err = fmt.Errorf("loadWaitingWorkers: %v", err)
		return
	}
	return
}

func (r *RedisRunner) transWaitingToReady(ws []string) error {
	c, cancel := context.WithTimeout(context.Background(), RedisTimeout)
	defer cancel()
	pip := r.redisCli.Pipeline()
	for _, w := range ws {
		vals := strings.Split(w, WaitingQueueDataIDSeparator)
		if len(vals) != 2 {
			logger.Errorf("invalid val: %v", w)
			continue
		}
		pip.RPush(c, vals[0], vals[1])
		pip.ZRem(c, KeyWaitingQueue, w)
	}
	_, err := pip.Exec(c)
	return err
}

func (r *RedisRunner) GetQueueLen(queue string) (int64, error) {
	c, cancel := context.WithTimeout(context.Background(), RedisTimeout)
	defer cancel()
	return r.redisCli.LLen(c, queue).Result()
}

func withWorkerLogger(wc *Meta) *log.Logger {
	return logger.With(
		zap.String("name", string(wc.Name)),
		zap.String("id", wc.ID),
		zap.String("retry", fmt.Sprintf("%d/%d", wc.RetryCount, wc.Retry)),
	)
}

func QueueKey(queue string) string {
	return Prefix + "Queue_" + queue
}
`)
}

func WorkerStatus() []byte {
	return []byte(`
package worker

type RunnerStatus struct {
	ExecCount int64
	FailCount int64
}
`)
}

func Worker() []byte {
	return []byte(`
package worker

// todo api
// todo queue attributes


import (
	"{{ .ProjectName }}/pkg/log"
)

type Worker interface {
	WorkerName() string
	Perform(ctx Context, logger *log.Logger) error
}
`)
}
