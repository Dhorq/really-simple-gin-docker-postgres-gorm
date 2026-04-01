package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"encoding/xml"
	"errors"
	"fmt"
	"hash"
	"io"
	"log"
	"math"
	"math/big"
	"math/bits"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode"
	"unsafe"
)

// ============================================================
// CONSTANTS & ENUMS
// ============================================================

const (
	MaxRetries     = 3
	DefaultTimeout = 30 * time.Second
	BufferSize     = 4096
	MaxWorkers     = 100
	Version        = "1.0.0"
	AppName        = "MegaGoApp"
)

type Status int

const (
	StatusPending Status = iota
	StatusRunning
	StatusDone
	StatusFailed
	StatusCancelled
)

func (s Status) String() string {
	switch s {
	case StatusPending:
		return "pending"
	case StatusRunning:
		return "running"
	case StatusDone:
		return "done"
	case StatusFailed:
		return "failed"
	case StatusCancelled:
		return "cancelled"
	default:
		return "unknown"
	}
}

type LogLevel int

const (
	LogDebug LogLevel = iota
	LogInfo
	LogWarn
	LogError
	LogFatal
)

// ============================================================
// INTERFACES
// ============================================================

type Reader interface {
	Read(p []byte) (n int, err error)
}

type Writer interface {
	Write(p []byte) (n int, err error)
}

type Closer interface {
	Close() error
}

type ReadWriter interface {
	Reader
	Writer
}

type Storage interface {
	Get(key string) ([]byte, error)
	Set(key string, value []byte) error
	Delete(key string) error
	Keys() ([]string, error)
}

type Cache interface {
	Storage
	TTL(key string) (time.Duration, error)
	SetWithTTL(key string, value []byte, ttl time.Duration) error
	Flush() error
}

type Encoder interface {
	Encode(v interface{}) error
}

type Decoder interface {
	Decode(v interface{}) error
}

type Serializer interface {
	Encoder
	Decoder
}

type EventHandler interface {
	Handle(event Event) error
}

type Middleware func(http.Handler) http.Handler

type Validator interface {
	Validate() error
}

type Transformer interface {
	Transform(input interface{}) (interface{}, error)
}

type Observable interface {
	Subscribe(observer Observer) error
	Unsubscribe(observer Observer) error
	Notify(event Event)
}

type Observer interface {
	Update(event Event)
}

type Repository[T any] interface {
	FindByID(ctx context.Context, id string) (*T, error)
	FindAll(ctx context.Context) ([]*T, error)
	Save(ctx context.Context, entity *T) error
	Delete(ctx context.Context, id string) error
}

type Service[T any] interface {
	Create(ctx context.Context, input T) (*T, error)
	Update(ctx context.Context, id string, input T) (*T, error)
	Delete(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (*T, error)
	List(ctx context.Context, filter Filter1) ([]*T, error)
}

// ============================================================
// STRUCTS & DATA MODELS
// ============================================================

type User struct {
	ID        string                 `json:"id" xml:"id" db:"id"`
	Username  string                 `json:"username" xml:"username" db:"username"`
	Email     string                 `json:"email" xml:"email" db:"email"`
	Password  string                 `json:"-" xml:"-" db:"password"`
	FirstName string                 `json:"first_name" xml:"first_name" db:"first_name"`
	LastName  string                 `json:"last_name" xml:"last_name" db:"last_name"`
	Age       int                    `json:"age" xml:"age" db:"age"`
	IsActive  bool                   `json:"is_active" xml:"is_active" db:"is_active"`
	Role      string                 `json:"role" xml:"role" db:"role"`
	CreatedAt time.Time              `json:"created_at" xml:"created_at" db:"created_at"`
	UpdatedAt time.Time              `json:"updated_at" xml:"updated_at" db:"updated_at"`
	DeletedAt *time.Time             `json:"deleted_at,omitempty" xml:"deleted_at,omitempty" db:"deleted_at"`
	Meta      map[string]interface{} `json:"meta,omitempty"`
}

func (u *User) Validate() error {
	if u.Username == "" {
		return errors.New("username is required")
	}
	if u.Email == "" {
		return errors.New("email is required")
	}
	if !isValidEmail(u.Email) {
		return errors.New("email is invalid")
	}
	if u.Age < 0 || u.Age > 150 {
		return errors.New("age must be between 0 and 150")
	}
	return nil
}

func (u *User) FullName() string {
	return strings.TrimSpace(u.FirstName + " " + u.LastName)
}

func (u *User) IsAdmin() bool {
	return u.Role == "admin"
}

type Product struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Price       float64           `json:"price"`
	Currency    string            `json:"currency"`
	Stock       int               `json:"stock"`
	CategoryID  string            `json:"category_id"`
	Tags        []string          `json:"tags"`
	Images      []Image           `json:"images"`
	Attributes  map[string]string `json:"attributes"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

type Image struct {
	URL    string `json:"url"`
	Alt    string `json:"alt"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Format string `json:"format"`
}

type Order struct {
	ID         string      `json:"id"`
	UserID     string      `json:"user_id"`
	Items      []OrderItem `json:"items"`
	Status     Status      `json:"status"`
	TotalPrice float64     `json:"total_price"`
	Currency   string      `json:"currency"`
	Address    Address     `json:"address"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
}

type OrderItem struct {
	ProductID string  `json:"product_id"`
	Name      string  `json:"name"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
	Total     float64 `json:"total"`
}

type Address struct {
	Street     string `json:"street"`
	City       string `json:"city"`
	State      string `json:"state"`
	Country    string `json:"country"`
	PostalCode string `json:"postal_code"`
}

type Event struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Source    string                 `json:"source"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

type Filter1 struct {
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
	Sort     string            `json:"sort"`
	Order    string            `json:"order"`
	Search   string            `json:"search"`
	Filters  map[string]string `json:"filters"`
}

type Pagination struct {
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
	Pages    int   `json:"pages"`
}

type APIResponse struct {
	Success    bool        `json:"success"`
	Message    string      `json:"message,omitempty"`
	Data       interface{} `json:"data,omitempty"`
	Error      string      `json:"error,omitempty"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

type Config struct {
	Server   ServerConfig   `json:"server" yaml:"server"`
	Database DatabaseConfig `json:"database" yaml:"database"`
	Redis    RedisConfig    `json:"redis" yaml:"redis"`
	JWT      JWTConfig      `json:"jwt" yaml:"jwt"`
	Log      LogConfig      `json:"log" yaml:"log"`
}

type ServerConfig struct {
	Host         string        `json:"host" yaml:"host"`
	Port         int           `json:"port" yaml:"port"`
	ReadTimeout  time.Duration `json:"read_timeout" yaml:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout" yaml:"write_timeout"`
	TLS          TLSConfig     `json:"tls" yaml:"tls"`
}

type TLSConfig struct {
	Enabled  bool   `json:"enabled" yaml:"enabled"`
	CertFile string `json:"cert_file" yaml:"cert_file"`
	KeyFile  string `json:"key_file" yaml:"key_file"`
}

type DatabaseConfig struct {
	Host            string        `json:"host" yaml:"host"`
	Port            int           `json:"port" yaml:"port"`
	Name            string        `json:"name" yaml:"name"`
	User            string        `json:"user" yaml:"user"`
	Password        string        `json:"password" yaml:"password"`
	SSLMode         string        `json:"ssl_mode" yaml:"ssl_mode"`
	MaxOpenConns    int           `json:"max_open_conns" yaml:"max_open_conns"`
	MaxIdleConns    int           `json:"max_idle_conns" yaml:"max_idle_conns"`
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime" yaml:"conn_max_lifetime"`
}

type RedisConfig struct {
	Host     string `json:"host" yaml:"host"`
	Port     int    `json:"port" yaml:"port"`
	Password string `json:"password" yaml:"password"`
	DB       int    `json:"db" yaml:"db"`
}

type JWTConfig struct {
	Secret     string        `json:"secret" yaml:"secret"`
	Expiration time.Duration `json:"expiration" yaml:"expiration"`
	Issuer     string        `json:"issuer" yaml:"issuer"`
}

type LogConfig struct {
	Level  string `json:"level" yaml:"level"`
	Format string `json:"format" yaml:"format"`
	Output string `json:"output" yaml:"output"`
}

// ============================================================
// GENERIC DATA STRUCTURES
// ============================================================

type Stack[T any] struct {
	items []T
	mu    sync.Mutex
}

func NewStack[T any]() *Stack[T] {
	return &Stack[T]{}
}

func (s *Stack[T]) Push(item T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = append(s.items, item)
}

func (s *Stack[T]) Pop() (T, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var zero T
	if len(s.items) == 0 {
		return zero, false
	}
	item := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return item, true
}

func (s *Stack[T]) Peek() (T, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var zero T
	if len(s.items) == 0 {
		return zero, false
	}
	return s.items[len(s.items)-1], true
}

func (s *Stack[T]) Size() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.items)
}

func (s *Stack[T]) IsEmpty() bool {
	return s.Size() == 0
}

type Queue[T any] struct {
	items []T
	mu    sync.Mutex
}

func NewQueue[T any]() *Queue[T] {
	return &Queue[T]{}
}

func (q *Queue[T]) Enqueue(item T) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.items = append(q.items, item)
}

func (q *Queue[T]) Dequeue() (T, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	var zero T
	if len(q.items) == 0 {
		return zero, false
	}
	item := q.items[0]
	q.items = q.items[1:]
	return item, true
}

func (q *Queue[T]) Size() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.items)
}

type Set[T comparable] struct {
	items map[T]struct{}
	mu    sync.RWMutex
}

func NewSet[T comparable]() *Set[T] {
	return &Set[T]{items: make(map[T]struct{})}
}

func (s *Set[T]) Add(item T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[item] = struct{}{}
}

func (s *Set[T]) Remove(item T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.items, item)
}

func (s *Set[T]) Contains(item T) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.items[item]
	return ok
}

func (s *Set[T]) ToSlice() []T {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]T, 0, len(s.items))
	for k := range s.items {
		result = append(result, k)
	}
	return result
}

func (s *Set[T]) Size() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.items)
}

type Pair[A, B any] struct {
	First  A
	Second B
}

func NewPair[A, B any](a A, b B) Pair[A, B] {
	return Pair[A, B]{First: a, Second: b}
}

type Result[T any] struct {
	Value T
	Err   error
}

func Ok[T any](value T) Result[T] {
	return Result[T]{Value: value}
}

func Err[T any](err error) Result[T] {
	return Result[T]{Err: err}
}

func (r Result[T]) IsOk() bool {
	return r.Err == nil
}

func (r Result[T]) Unwrap() T {
	if r.Err != nil {
		panic(r.Err)
	}
	return r.Value
}

type Optional[T any] struct {
	value   T
	present bool
}

func Some[T any](value T) Optional[T] {
	return Optional[T]{value: value, present: true}
}

func None[T any]() Optional[T] {
	return Optional[T]{}
}

func (o Optional[T]) IsPresent() bool { return o.present }

func (o Optional[T]) Get() (T, bool) { return o.value, o.present }

func (o Optional[T]) OrElse(defaultVal T) T {
	if o.present {
		return o.value
	}
	return defaultVal
}

// ============================================================
// IN-MEMORY STORE (Thread-Safe)
// ============================================================

type MemStore struct {
	data map[string][]byte
	mu   sync.RWMutex
}

func NewMemStore() *MemStore {
	return &MemStore{data: make(map[string][]byte)}
}

func (m *MemStore) Get(key string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	val, ok := m.data[key]
	if !ok {
		return nil, fmt.Errorf("key %q not found", key)
	}
	return val, nil
}

func (m *MemStore) Set(key string, value []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value
	return nil
}

func (m *MemStore) Delete(key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, key)
	return nil
}

func (m *MemStore) Keys() ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	keys := make([]string, 0, len(m.data))
	for k := range m.data {
		keys = append(keys, k)
	}
	return keys, nil
}

// ============================================================
// WORKER POOL
// ============================================================

type Job struct {
	ID      string
	Payload interface{}
	Handler func(ctx context.Context, payload interface{}) (interface{}, error)
}

type JobResult struct {
	JobID  string
	Output interface{}
	Error  error
}

type WorkerPool struct {
	numWorkers int
	jobs       chan Job
	results    chan JobResult
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
}

func NewWorkerPool(numWorkers, bufferSize int) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	return &WorkerPool{
		numWorkers: numWorkers,
		jobs:       make(chan Job, bufferSize),
		results:    make(chan JobResult, bufferSize),
		ctx:        ctx,
		cancel:     cancel,
	}
}

func (wp *WorkerPool) Start() {
	for i := 0; i < wp.numWorkers; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}
}

func (wp *WorkerPool) worker(id int) {
	defer wp.wg.Done()
	for {
		select {
		case <-wp.ctx.Done():
			return
		case job, ok := <-wp.jobs:
			if !ok {
				return
			}
			output, err := job.Handler(wp.ctx, job.Payload)
			wp.results <- JobResult{JobID: job.ID, Output: output, Error: err}
		}
	}
}

func (wp *WorkerPool) Submit(job Job) {
	wp.jobs <- job
}

func (wp *WorkerPool) Results() <-chan JobResult {
	return wp.results
}

func (wp *WorkerPool) Stop() {
	wp.cancel()
	close(wp.jobs)
	wp.wg.Wait()
	close(wp.results)
}

// ============================================================
// EVENT BUS
// ============================================================

type EventBus struct {
	subscribers map[string][]chan Event
	mu          sync.RWMutex
}

func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[string][]chan Event),
	}
}

func (eb *EventBus) Subscribe(eventType string) <-chan Event {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	ch := make(chan Event, 10)
	eb.subscribers[eventType] = append(eb.subscribers[eventType], ch)
	return ch
}

func (eb *EventBus) Publish(event Event) {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	for _, ch := range eb.subscribers[event.Type] {
		select {
		case ch <- event:
		default:
		}
	}
}

func (eb *EventBus) Unsubscribe(eventType string, ch <-chan Event) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	subs := eb.subscribers[eventType]
	for i, sub := range subs {
		if sub == ch {
			eb.subscribers[eventType] = append(subs[:i], subs[i+1:]...)
			close(sub)
			break
		}
	}
}

// ============================================================
// RATE LIMITER
// ============================================================

type RateLimiter struct {
	tokens   float64
	maxToken float64
	rate     float64
	lastTime time.Time
	mu       sync.Mutex
}

func NewRateLimiter(rate float64, burst float64) *RateLimiter {
	return &RateLimiter{
		tokens:   burst,
		maxToken: burst,
		rate:     rate,
		lastTime: time.Now(),
	}
}

func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	elapsed := now.Sub(rl.lastTime).Seconds()
	rl.lastTime = now
	rl.tokens = math.Min(rl.maxToken, rl.tokens+elapsed*rl.rate)
	if rl.tokens >= 1 {
		rl.tokens--
		return true
	}
	return false
}

func (rl *RateLimiter) Wait(ctx context.Context) error {
	for {
		if rl.Allow() {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Millisecond * 10):
		}
	}
}

// ============================================================
// CIRCUIT BREAKER
// ============================================================

type CircuitState int

const (
	CircuitClosed CircuitState = iota
	CircuitOpen
	CircuitHalfOpen
)

type CircuitBreaker struct {
	state        CircuitState
	failures     int
	maxFailures  int
	resetTimeout time.Duration
	lastFailure  time.Time
	mu           sync.Mutex
}

func NewCircuitBreaker(maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:        CircuitClosed,
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
	}
}

func (cb *CircuitBreaker) Execute(fn func() error) error {
	cb.mu.Lock()
	state := cb.state
	cb.mu.Unlock()

	if state == CircuitOpen {
		if time.Since(cb.lastFailure) > cb.resetTimeout {
			cb.mu.Lock()
			cb.state = CircuitHalfOpen
			cb.mu.Unlock()
		} else {
			return errors.New("circuit breaker is open")
		}
	}

	err := fn()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.failures++
		cb.lastFailure = time.Now()
		if cb.failures >= cb.maxFailures {
			cb.state = CircuitOpen
		}
		return err
	}

	cb.failures = 0
	cb.state = CircuitClosed
	return nil
}

// ============================================================
// RETRY MECHANISM
// ============================================================

type RetryConfig struct {
	MaxAttempts int
	Delay       time.Duration
	MaxDelay    time.Duration
	Multiplier  float64
	Jitter      bool
}

func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts: 3,
		Delay:       100 * time.Millisecond,
		MaxDelay:    10 * time.Second,
		Multiplier:  2.0,
		Jitter:      true,
	}
}

func Retry(ctx context.Context, cfg RetryConfig, fn func() error) error {
	var lastErr error
	delay := cfg.Delay

	for attempt := 0; attempt < cfg.MaxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}

		lastErr = fn()
		if lastErr == nil {
			return nil
		}

		if attempt == cfg.MaxAttempts-1 {
			break
		}

		sleep := delay
		if cfg.Jitter {
			n, _ := rand.Int(rand.Reader, big.NewInt(int64(delay)))
			sleep += time.Duration(n.Int64())
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(sleep):
		}

		delay = time.Duration(float64(delay) * cfg.Multiplier)
		if delay > cfg.MaxDelay {
			delay = cfg.MaxDelay
		}
	}
	return fmt.Errorf("max attempts reached: %w", lastErr)
}

// ============================================================
// LOGGER
// ============================================================

type Logger struct {
	level  LogLevel
	output io.Writer
	prefix string
	mu     sync.Mutex
}

func NewLogger(level LogLevel, output io.Writer, prefix string) *Logger {
	return &Logger{level: level, output: output, prefix: prefix}
}

func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	if level < l.level {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	levelStr := []string{"DEBUG", "INFO", "WARN", "ERROR", "FATAL"}[level]
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(l.output, "[%s] [%s] %s: %s\n",
		time.Now().Format(time.RFC3339), levelStr, l.prefix, msg)
}

func (l *Logger) Debug(format string, args ...interface{}) { l.log(LogDebug, format, args...) }
func (l *Logger) Info(format string, args ...interface{})  { l.log(LogInfo, format, args...) }
func (l *Logger) Warn(format string, args ...interface{})  { l.log(LogWarn, format, args...) }
func (l *Logger) Error(format string, args ...interface{}) { l.log(LogError, format, args...) }
func (l *Logger) Fatal(format string, args ...interface{}) {
	l.log(LogFatal, format, args...)
	os.Exit(1)
}

// ============================================================
// HTTP CLIENT WITH RETRY
// ============================================================

type HTTPClient struct {
	client  *http.Client
	baseURL string
	headers map[string]string
	retry   RetryConfig
	logger  *Logger
}

func NewHTTPClient(baseURL string, timeout time.Duration) *HTTPClient {
	return &HTTPClient{
		client:  &http.Client{Timeout: timeout},
		baseURL: baseURL,
		headers: make(map[string]string),
		retry:   DefaultRetryConfig(),
		logger:  NewLogger(LogInfo, os.Stdout, "HTTPClient"),
	}
}

func (c *HTTPClient) SetHeader(key, value string) {
	c.headers[key] = value
}

func (c *HTTPClient) SetBearerToken(token string) {
	c.headers["Authorization"] = "Bearer " + token
}

func (c *HTTPClient) buildRequest(method, path string, body io.Reader) (*http.Request, error) {
	u, err := url.JoinPath(c.baseURL, path)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(method, u, body)
	if err != nil {
		return nil, err
	}
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}
	return req, nil
}

func (c *HTTPClient) Do(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	var resp *http.Response
	err := Retry(ctx, c.retry, func() error {
		req, err := c.buildRequest(method, path, body)
		if err != nil {
			return err
		}
		req = req.WithContext(ctx)
		resp, err = c.client.Do(req)
		if err != nil {
			return err
		}
		if resp.StatusCode >= 500 {
			return fmt.Errorf("server error: %d", resp.StatusCode)
		}
		return nil
	})
	return resp, err
}

func (c *HTTPClient) Get(ctx context.Context, path string) (*http.Response, error) {
	return c.Do(ctx, http.MethodGet, path, nil)
}

func (c *HTTPClient) Post(ctx context.Context, path string, body io.Reader) (*http.Response, error) {
	return c.Do(ctx, http.MethodPost, path, body)
}

func (c *HTTPClient) PostJSON(ctx context.Context, path string, payload interface{}) (*http.Response, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	c.SetHeader("Content-Type", "application/json")
	return c.Post(ctx, path, bytes.NewReader(data))
}

// ============================================================
// ENCRYPTION UTILITIES
// ============================================================

func GenerateKey(size int) ([]byte, error) {
	key := make([]byte, size)
	_, err := io.ReadFull(rand.Reader, key)
	return key, err
}

func EncryptAES(key, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

func DecryptAES(key, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

func HashSHA256(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

func HashMD5(data []byte) string {
	h := md5.Sum(data)
	return hex.EncodeToString(h[:])
}

func HMACSHA256(key, data []byte) string {
	mac := hmac.New(sha256.New, key)
	mac.Write(data)
	return hex.EncodeToString(mac.Sum(nil))
}

func Base64Encode(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

func Base64Decode(s string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(s)
}

func GenerateRSAKeyPair(bits int) (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, bits)
}

func ExportRSAPrivateKeyAsPEM(key *rsa.PrivateKey) string {
	privBytes := x509.MarshalPKCS1PrivateKey(key)
	privPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: privBytes})
	return string(privPEM)
}

// ============================================================
// FILE UTILITIES
// ============================================================

func ReadFile(path string) ([]byte, error) {
	return os.ReadFile(filepath.Clean(path))
}

func WriteFile(path string, data []byte) error {
	return os.WriteFile(filepath.Clean(path), data, 0600)
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func ReadLines(path string) ([]string, error) {
	f, err := os.Open(filepath.Clean(path))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func WriteLines(path string, lines []string) error {
	f, err := os.Create(filepath.Clean(path))
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	for _, line := range lines {
		fmt.Fprintln(w, line)
	}
	return w.Flush()
}

func ReadCSV(path string) ([][]string, error) {
	f, err := os.Open(filepath.Clean(path))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return csv.NewReader(f).ReadAll()
}

func ReadJSON(path string, v interface{}) error {
	data, err := ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

func WriteJSON(path string, v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return WriteFile(path, data)
}

func CopyFile(src, dst string) error {
	source, err := os.Open(filepath.Clean(src))
	if err != nil {
		return err
	}
	defer source.Close()
	destination, err := os.Create(filepath.Clean(dst))
	if err != nil {
		return err
	}
	defer destination.Close()
	_, err = io.Copy(destination, source)
	return err
}

func WalkDir(root string, fn func(path string, info os.FileInfo) error) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		return fn(path, info)
	})
}

// ============================================================
// STRING UTILITIES
// ============================================================

func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func PadLeft(s string, length int, pad rune) string {
	for len([]rune(s)) < length {
		s = string(pad) + s
	}
	return s
}

func PadRight(s string, length int, pad rune) string {
	for len([]rune(s)) < length {
		s = s + string(pad)
	}
	return s
}

func IsAlphaNumeric(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func ToSnakeCase(s string) string {
	var result []rune
	for i, r := range s {
		if i > 0 && unicode.IsUpper(r) {
			result = append(result, '_')
		}
		result = append(result, unicode.ToLower(r))
	}
	return string(result)
}

func ToCamelCase(s string) string {
	parts := strings.Split(s, "_")
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	return strings.Join(parts, "")
}

func ReverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func CountWords(s string) int {
	return len(strings.Fields(s))
}

func WrapText(s string, width int) []string {
	words := strings.Fields(s)
	var lines []string
	var current strings.Builder
	for _, word := range words {
		if current.Len()+len(word)+1 > width && current.Len() > 0 {
			lines = append(lines, current.String())
			current.Reset()
		}
		if current.Len() > 0 {
			current.WriteByte(' ')
		}
		current.WriteString(word)
	}
	if current.Len() > 0 {
		lines = append(lines, current.String())
	}
	return lines
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

func isValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}

func isValidURL(rawURL string) bool {
	u, err := url.ParseRequestURI(rawURL)
	return err == nil && u.Scheme != "" && u.Host != ""
}

// ============================================================
// MATH & NUMBER UTILITIES
// ============================================================

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func Abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func Clamp(val, minVal, maxVal int) int {
	return Max(minVal, Min(maxVal, val))
}

func Sum(nums []int) int {
	total := 0
	for _, n := range nums {
		total += n
	}
	return total
}

func Average(nums []float64) float64 {
	if len(nums) == 0 {
		return 0
	}
	var sum float64
	for _, n := range nums {
		sum += n
	}
	return sum / float64(len(nums))
}

func Median(nums []float64) float64 {
	if len(nums) == 0 {
		return 0
	}
	sorted := make([]float64, len(nums))
	copy(sorted, nums)
	sort.Float64s(sorted)
	mid := len(sorted) / 2
	if len(sorted)%2 == 0 {
		return (sorted[mid-1] + sorted[mid]) / 2
	}
	return sorted[mid]
}

func StdDev(nums []float64) float64 {
	avg := Average(nums)
	var sumSq float64
	for _, n := range nums {
		diff := n - avg
		sumSq += diff * diff
	}
	return math.Sqrt(sumSq / float64(len(nums)))
}

func IsPrime(n int) bool {
	if n < 2 {
		return false
	}
	for i := 2; i <= int(math.Sqrt(float64(n))); i++ {
		if n%i == 0 {
			return false
		}
	}
	return true
}

func Fibonacci(n int) []int {
	if n <= 0 {
		return nil
	}
	fibs := make([]int, n)
	fibs[0] = 0
	if n > 1 {
		fibs[1] = 1
	}
	for i := 2; i < n; i++ {
		fibs[i] = fibs[i-1] + fibs[i-2]
	}
	return fibs
}

func GCD(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

func LCM(a, b int) int {
	return Abs(a*b) / GCD(a, b)
}

func PowerMod(base, exp, mod int) int {
	result := 1
	base %= mod
	for exp > 0 {
		if exp%2 == 1 {
			result = result * base % mod
		}
		exp /= 2
		base = base * base % mod
	}
	return result
}

// ============================================================
// SORTING ALGORITHMS
// ============================================================

func BubbleSort(arr []int) []int {
	n := len(arr)
	result := make([]int, n)
	copy(result, arr)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if result[j] > result[j+1] {
				result[j], result[j+1] = result[j+1], result[j]
			}
		}
	}
	return result
}

func MergeSort(arr []int) []int {
	if len(arr) <= 1 {
		return arr
	}
	mid := len(arr) / 2
	left := MergeSort(arr[:mid])
	right := MergeSort(arr[mid:])
	return merge(left, right)
}

func merge(left, right []int) []int {
	result := make([]int, 0, len(left)+len(right))
	i, j := 0, 0
	for i < len(left) && j < len(right) {
		if left[i] <= right[j] {
			result = append(result, left[i])
			i++
		} else {
			result = append(result, right[j])
			j++
		}
	}
	result = append(result, left[i:]...)
	result = append(result, right[j:]...)
	return result
}

func QuickSort(arr []int) []int {
	if len(arr) <= 1 {
		return arr
	}
	pivot := arr[len(arr)/2]
	var left, right, equal []int
	for _, v := range arr {
		switch {
		case v < pivot:
			left = append(left, v)
		case v > pivot:
			right = append(right, v)
		default:
			equal = append(equal, v)
		}
	}
	result := QuickSort(left)
	result = append(result, equal...)
	result = append(result, QuickSort(right)...)
	return result
}

func BinarySearch(arr []int, target int) int {
	lo, hi := 0, len(arr)-1
	for lo <= hi {
		mid := (lo + hi) / 2
		if arr[mid] == target {
			return mid
		} else if arr[mid] < target {
			lo = mid + 1
		} else {
			hi = mid - 1
		}
	}
	return -1
}

// ============================================================
// LINKED LIST
// ============================================================

type ListNode[T any] struct {
	Value T
	Next  *ListNode[T]
}

type LinkedList[T any] struct {
	Head *ListNode[T]
	size int
}

func (l *LinkedList[T]) Prepend(value T) {
	l.Head = &ListNode[T]{Value: value, Next: l.Head}
	l.size++
}

func (l *LinkedList[T]) Append(value T) {
	node := &ListNode[T]{Value: value}
	if l.Head == nil {
		l.Head = node
	} else {
		current := l.Head
		for current.Next != nil {
			current = current.Next
		}
		current.Next = node
	}
	l.size++
}

func (l *LinkedList[T]) ToSlice() []T {
	var result []T
	current := l.Head
	for current != nil {
		result = append(result, current.Value)
		current = current.Next
	}
	return result
}

func (l *LinkedList[T]) Size() int {
	return l.size
}

// ============================================================
// BINARY TREE
// ============================================================

type TreeNode struct {
	Val   int
	Left  *TreeNode
	Right *TreeNode
}

func InsertBST(root *TreeNode, val int) *TreeNode {
	if root == nil {
		return &TreeNode{Val: val}
	}
	if val < root.Val {
		root.Left = InsertBST(root.Left, val)
	} else {
		root.Right = InsertBST(root.Right, val)
	}
	return root
}

func InOrder(root *TreeNode) []int {
	if root == nil {
		return nil
	}
	result := InOrder(root.Left)
	result = append(result, root.Val)
	result = append(result, InOrder(root.Right)...)
	return result
}

func TreeHeight(root *TreeNode) int {
	if root == nil {
		return 0
	}
	left := TreeHeight(root.Left)
	right := TreeHeight(root.Right)
	if left > right {
		return left + 1
	}
	return right + 1
}

func SearchBST(root *TreeNode, val int) *TreeNode {
	if root == nil || root.Val == val {
		return root
	}
	if val < root.Val {
		return SearchBST(root.Left, val)
	}
	return SearchBST(root.Right, val)
}

// ============================================================
// GRAPH
// ============================================================

type Graph struct {
	vertices int
	edges    map[int][]int
}

func NewGraph(v int) *Graph {
	return &Graph{vertices: v, edges: make(map[int][]int)}
}

func (g *Graph) AddEdge(u, v int) {
	g.edges[u] = append(g.edges[u], v)
	g.edges[v] = append(g.edges[v], u)
}

func (g *Graph) BFS(start int) []int {
	visited := make(map[int]bool)
	queue := []int{start}
	visited[start] = true
	var result []int
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		result = append(result, node)
		for _, neighbor := range g.edges[node] {
			if !visited[neighbor] {
				visited[neighbor] = true
				queue = append(queue, neighbor)
			}
		}
	}
	return result
}

func (g *Graph) DFS(start int) []int {
	visited := make(map[int]bool)
	var result []int
	var dfs func(node int)
	dfs = func(node int) {
		visited[node] = true
		result = append(result, node)
		for _, neighbor := range g.edges[node] {
			if !visited[neighbor] {
				dfs(neighbor)
			}
		}
	}
	dfs(start)
	return result
}

func (g *Graph) HasCycle() bool {
	visited := make(map[int]bool)
	var hasCycle func(node, parent int) bool
	hasCycle = func(node, parent int) bool {
		visited[node] = true
		for _, neighbor := range g.edges[node] {
			if !visited[neighbor] {
				if hasCycle(neighbor, node) {
					return true
				}
			} else if neighbor != parent {
				return true
			}
		}
		return false
	}
	for v := 0; v < g.vertices; v++ {
		if !visited[v] {
			if hasCycle(v, -1) {
				return true
			}
		}
	}
	return false
}

// ============================================================
// MIDDLEWARE CHAIN
// ============================================================

type HandlerFunc func(http.ResponseWriter, *http.Request)

func Chain(h http.Handler, middlewares ...Middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

func LoggingMiddleware(logger *Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			logger.Info("%s %s", r.Method, r.URL.Path)
			next.ServeHTTP(w, r)
			logger.Info("completed in %s", time.Since(start))
		})
	}
}

func RecoveryMiddleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					log.Printf("panic: %v", err)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

func CORSMiddleware(allowedOrigins []string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			for _, allowed := range allowedOrigins {
				if allowed == "*" || allowed == origin {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					break
				}
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func RateLimitMiddleware(limiter *RateLimiter) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !limiter.Allow() {
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// ============================================================
// JSON RESPONSE HELPERS
// ============================================================

func RespondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func RespondSuccess(w http.ResponseWriter, data interface{}) {
	RespondJSON(w, http.StatusOK, APIResponse{Success: true, Data: data})
}

func RespondError(w http.ResponseWriter, status int, message string) {
	RespondJSON(w, status, APIResponse{Success: false, Error: message})
}

func RespondCreated(w http.ResponseWriter, data interface{}) {
	RespondJSON(w, http.StatusCreated, APIResponse{Success: true, Data: data})
}

// ============================================================
// CHANNEL PATTERNS
// ============================================================

func FanOut[T any](input <-chan T, numOutputs int) []<-chan T {
	outputs := make([]chan T, numOutputs)
	for i := range outputs {
		outputs[i] = make(chan T)
	}
	go func() {
		defer func() {
			for _, ch := range outputs {
				close(ch)
			}
		}()
		for val := range input {
			for _, ch := range outputs {
				ch <- val
			}
		}
	}()
	result := make([]<-chan T, numOutputs)
	for i, ch := range outputs {
		result[i] = ch
	}
	return result
}

func FanIn[T any](inputs ...<-chan T) <-chan T {
	output := make(chan T)
	var wg sync.WaitGroup
	for _, input := range inputs {
		wg.Add(1)
		go func(ch <-chan T) {
			defer wg.Done()
			for val := range ch {
				output <- val
			}
		}(input)
	}
	go func() {
		wg.Wait()
		close(output)
	}()
	return output
}

func Pipeline[T, U any](input <-chan T, fn func(T) U) <-chan U {
	output := make(chan U)
	go func() {
		defer close(output)
		for val := range input {
			output <- fn(val)
		}
	}()
	return output
}

func Batch[T any](input <-chan T, size int) <-chan []T {
	output := make(chan []T)
	go func() {
		defer close(output)
		batch := make([]T, 0, size)
		for val := range input {
			batch = append(batch, val)
			if len(batch) >= size {
				output <- batch
				batch = make([]T, 0, size)
			}
		}
		if len(batch) > 0 {
			output <- batch
		}
	}()
	return output
}

// ============================================================
// ATOMIC COUNTER
// ============================================================

type AtomicCounter struct {
	value int64
}

func (c *AtomicCounter) Increment() int64  { return atomic.AddInt64(&c.value, 1) }
func (c *AtomicCounter) Decrement() int64  { return atomic.AddInt64(&c.value, -1) }
func (c *AtomicCounter) Get() int64        { return atomic.LoadInt64(&c.value) }
func (c *AtomicCounter) Reset()            { atomic.StoreInt64(&c.value, 0) }
func (c *AtomicCounter) Add(n int64) int64 { return atomic.AddInt64(&c.value, n) }

// ============================================================
// XML ENCODING EXAMPLE
// ============================================================

type XMLProduct struct {
	XMLName xml.Name `xml:"product"`
	ID      string   `xml:"id,attr"`
	Name    string   `xml:"name"`
	Price   float64  `xml:"price"`
	InStock bool     `xml:"in_stock"`
}

func MarshalXML(v interface{}) (string, error) {
	data, err := xml.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", err
	}
	return xml.Header + string(data), nil
}

// ============================================================
// NETWORK UTILITIES
// ============================================================

func GetLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String(), nil
			}
		}
	}
	return "", errors.New("no local IP found")
}

func IsPortOpen(host string, port int, timeout time.Duration) bool {
	address := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// ============================================================
// REFLECTION UTILITIES
// ============================================================

func GetStructFields(v interface{}) []string {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil
	}
	fields := make([]string, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		fields[i] = t.Field(i).Name
	}
	return fields
}

func StructToMap(v interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	val := reflect.ValueOf(v)
	typ := reflect.TypeOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
		typ = typ.Elem()
	}
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		result[field.Name] = val.Field(i).Interface()
	}
	return result
}

// ============================================================
// HASH RING (CONSISTENT HASHING)
// ============================================================

type HashRing struct {
	ring     map[uint32]string
	sorted   []uint32
	replicas int
	mu       sync.RWMutex
}

func NewHashRing(replicas int) *HashRing {
	return &HashRing{
		ring:     make(map[uint32]string),
		replicas: replicas,
	}
}

func (h *HashRing) hashKey(key string) uint32 {
	hv := fnv32(key)
	return hv
}

func fnv32(key string) uint32 {
	const (
		offset32 uint32 = 2166136261
		prime32  uint32 = 16777619
	)
	hash := offset32
	for i := 0; i < len(key); i++ {
		hash ^= uint32(key[i])
		hash *= prime32
	}
	return hash
}

func (h *HashRing) Add(node string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for i := 0; i < h.replicas; i++ {
		key := h.hashKey(fmt.Sprintf("%s:%d", node, i))
		h.ring[key] = node
		h.sorted = append(h.sorted, key)
	}
	sort.Slice(h.sorted, func(i, j int) bool { return h.sorted[i] < h.sorted[j] })
}

func (h *HashRing) Get(key string) string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if len(h.ring) == 0 {
		return ""
	}
	hash := h.hashKey(key)
	idx := sort.Search(len(h.sorted), func(i int) bool { return h.sorted[i] >= hash })
	if idx == len(h.sorted) {
		idx = 0
	}
	return h.ring[h.sorted[idx]]
}

// ============================================================
// BLOOM FILTER
// ============================================================

type BloomFilter struct {
	bits    []bool
	size    uint
	hashFns []hash.Hash
}

func NewBloomFilter(size uint) *BloomFilter {
	return &BloomFilter{
		bits: make([]bool, size),
		size: size,
	}
}

func (bf *BloomFilter) hash(data []byte) []uint {
	h1 := fnv32(string(data))
	h2 := uint32(bits.RotateLeft32(h1, 5))
	return []uint{
		uint(h1) % uint(bf.size),
		uint(h2) % uint(bf.size),
		uint(h1^h2) % uint(bf.size),
	}
}

func (bf *BloomFilter) Add(item string) {
	for _, idx := range bf.hash([]byte(item)) {
		bf.bits[idx] = true
	}
}

func (bf *BloomFilter) MightContain(item string) bool {
	for _, idx := range bf.hash([]byte(item)) {
		if !bf.bits[idx] {
			return false
		}
	}
	return true
}

// ============================================================
// RUNTIME & SYSTEM INFO
// ============================================================

type SystemInfo struct {
	OS           string   `json:"os"`
	Arch         string   `json:"arch"`
	NumCPU       int      `json:"num_cpu"`
	GoVersion    string   `json:"go_version"`
	NumGoroutine int      `json:"num_goroutine"`
	MemStats     MemStats `json:"mem_stats"`
}

type MemStats struct {
	Alloc      uint64 `json:"alloc_mb"`
	TotalAlloc uint64 `json:"total_alloc_mb"`
	Sys        uint64 `json:"sys_mb"`
	NumGC      uint32 `json:"num_gc"`
}

func GetSystemInfo() SystemInfo {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	return SystemInfo{
		OS:           runtime.GOOS,
		Arch:         runtime.GOARCH,
		NumCPU:       runtime.NumCPU(),
		GoVersion:    runtime.Version(),
		NumGoroutine: runtime.NumGoroutine(),
		MemStats: MemStats{
			Alloc:      ms.Alloc / 1024 / 1024,
			TotalAlloc: ms.TotalAlloc / 1024 / 1024,
			Sys:        ms.Sys / 1024 / 1024,
			NumGC:      ms.NumGC,
		},
	}
}

// ============================================================
// UNSAFE POINTER TRICKS (for demonstration)
// ============================================================

func StringToBytes(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(&s))
}

func BytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// ============================================================
// BUILDER PATTERN
// ============================================================

type UserBuilder struct {
	user User
}

func NewUserBuilder() *UserBuilder {
	return &UserBuilder{}
}

func (b *UserBuilder) WithID(id string) *UserBuilder {
	b.user.ID = id
	return b
}

func (b *UserBuilder) WithUsername(username string) *UserBuilder {
	b.user.Username = username
	return b
}

func (b *UserBuilder) WithEmail(email string) *UserBuilder {
	b.user.Email = email
	return b
}

func (b *UserBuilder) WithRole(role string) *UserBuilder {
	b.user.Role = role
	return b
}

func (b *UserBuilder) WithName(first, last string) *UserBuilder {
	b.user.FirstName = first
	b.user.LastName = last
	return b
}

func (b *UserBuilder) Active() *UserBuilder {
	b.user.IsActive = true
	return b
}

func (b *UserBuilder) Build() (*User, error) {
	b.user.CreatedAt = time.Now()
	b.user.UpdatedAt = time.Now()
	if err := b.user.Validate(); err != nil {
		return nil, err
	}
	return &b.user, nil
}

// ============================================================
// FUNCTIONAL HELPERS
// ============================================================

func Map[T, U any](slice []T, fn func(T) U) []U {
	result := make([]U, len(slice))
	for i, v := range slice {
		result[i] = fn(v)
	}
	return result
}

func Filter[T any](slice []T, fn func(T) bool) []T {
	var result []T
	for _, v := range slice {
		if fn(v) {
			result = append(result, v)
		}
	}
	return result
}

func Reduce[T, U any](slice []T, initial U, fn func(U, T) U) U {
	acc := initial
	for _, v := range slice {
		acc = fn(acc, v)
	}
	return acc
}

func ForEach[T any](slice []T, fn func(T)) {
	for _, v := range slice {
		fn(v)
	}
}

func Any[T any](slice []T, fn func(T) bool) bool {
	for _, v := range slice {
		if fn(v) {
			return true
		}
	}
	return false
}

func All[T any](slice []T, fn func(T) bool) bool {
	for _, v := range slice {
		if !fn(v) {
			return false
		}
	}
	return true
}

func GroupBy[T any, K comparable](slice []T, fn func(T) K) map[K][]T {
	result := make(map[K][]T)
	for _, v := range slice {
		key := fn(v)
		result[key] = append(result[key], v)
	}
	return result
}

func Chunk[T any](slice []T, size int) [][]T {
	var chunks [][]T
	for i := 0; i < len(slice); i += size {
		end := i + size
		if end > len(slice) {
			end = len(slice)
		}
		chunks = append(chunks, slice[i:end])
	}
	return chunks
}

func Flatten[T any](slices [][]T) []T {
	var result []T
	for _, s := range slices {
		result = append(result, s...)
	}
	return result
}

func Unique[T comparable](slice []T) []T {
	seen := make(map[T]struct{})
	var result []T
	for _, v := range slice {
		if _, ok := seen[v]; !ok {
			seen[v] = struct{}{}
			result = append(result, v)
		}
	}
	return result
}

func Zip[T, U any](a []T, b []U) []Pair[T, U] {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	result := make([]Pair[T, U], n)
	for i := 0; i < n; i++ {
		result[i] = NewPair(a[i], b[i])
	}
	return result
}

// ============================================================
// MAIN
// ============================================================

func main() {
	fmt.Println("=== MegaGoApp v" + Version + " ===")

	// System info
	info := GetSystemInfo()
	data, _ := json.MarshalIndent(info, "", "  ")
	fmt.Println("System Info:", string(data))

	// Worker pool demo
	pool := NewWorkerPool(5, 20)
	pool.Start()

	for i := 0; i < 10; i++ {
		id := strconv.Itoa(i)
		pool.Submit(Job{
			ID:      id,
			Payload: i,
			Handler: func(ctx context.Context, payload interface{}) (interface{}, error) {
				n := payload.(int)
				return n * n, nil
			},
		})
	}

	go func() {
		for result := range pool.Results() {
			if result.Error != nil {
				fmt.Printf("Job %s failed: %v\n", result.JobID, result.Error)
			} else {
				fmt.Printf("Job %s = %v\n", result.JobID, result.Output)
			}
		}
	}()

	time.Sleep(100 * time.Millisecond)
	pool.Stop()

	// Generics demo
	nums := []int{5, 3, 8, 1, 9, 2, 7, 4, 6}
	fmt.Println("BubbleSort:", BubbleSort(nums))
	fmt.Println("MergeSort:", MergeSort(nums))
	fmt.Println("QuickSort:", QuickSort(nums))

	sorted := MergeSort(nums)
	fmt.Println("BinarySearch(7):", BinarySearch(sorted, 7))

	// Fibonacci
	fmt.Println("Fib(10):", Fibonacci(10))

	// Crypto
	key, _ := GenerateKey(32)
	plaintext := []byte("Hello, Go World!")
	encrypted, _ := EncryptAES(key, plaintext)
	decrypted, _ := DecryptAES(key, encrypted)
	fmt.Printf("Encrypted+Decrypted: %s\n", string(decrypted))

	// Hash ring
	ring := NewHashRing(3)
	ring.Add("node1")
	ring.Add("node2")
	ring.Add("node3")
	fmt.Println("HashRing node for 'user123':", ring.Get("user123"))

	// Bloom filter
	bf := NewBloomFilter(1000)
	bf.Add("golang")
	bf.Add("rust")
	fmt.Println("BloomFilter 'golang':", bf.MightContain("golang"))
	fmt.Println("BloomFilter 'python':", bf.MightContain("python"))

	// Functional helpers
	doubled := Map([]int{1, 2, 3, 4, 5}, func(n int) int { return n * 2 })
	evens := Filter([]int{1, 2, 3, 4, 5, 6}, func(n int) bool { return n%2 == 0 })
	total := Reduce([]int{1, 2, 3, 4, 5}, 0, func(acc, n int) int { return acc + n })
	fmt.Println("Doubled:", doubled)
	fmt.Println("Evens:", evens)
	fmt.Println("Sum:", total)

	// Stack & Queue
	stack := NewStack[string]()
	stack.Push("a")
	stack.Push("b")
	stack.Push("c")
	top, _ := stack.Pop()
	fmt.Println("Stack pop:", top)

	queue := NewQueue[int]()
	queue.Enqueue(1)
	queue.Enqueue(2)
	front, _ := queue.Dequeue()
	fmt.Println("Queue dequeue:", front)

	// User builder
	user, err := NewUserBuilder().
		WithID("u001").
		WithUsername("johndoe").
		WithEmail("john@example.com").
		WithName("John", "Doe").
		WithRole("admin").
		Active().
		Build()
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("User:", user.FullName(), "| Admin:", user.IsAdmin())
	}

	// Event bus
	bus := NewEventBus()
	ch := bus.Subscribe("user.created")
	bus.Publish(Event{
		ID:        "e001",
		Type:      "user.created",
		Source:    "auth-service",
		Timestamp: time.Now(),
		Data:      map[string]interface{}{"user_id": "u001"},
	})
	select {
	case evt := <-ch:
		fmt.Println("Received event:", evt.Type, "from", evt.Source)
	default:
	}

	// Rate limiter
	rl := NewRateLimiter(10, 5)
	allowed := 0
	for i := 0; i < 10; i++ {
		if rl.Allow() {
			allowed++
		}
	}
	fmt.Println("Rate limiter allowed:", allowed, "/ 10")

	// String utils
	fmt.Println("Snake case:", ToSnakeCase("HelloWorldGoLang"))
	fmt.Println("Camel case:", ToCamelCase("hello_world_go_lang"))
	fmt.Println("Reversed:", ReverseString("golang"))
	fmt.Println("Valid email:", isValidEmail("user@example.com"))
	fmt.Println("Valid URL:", isValidURL("https://example.com"))

	// BST
	var root *TreeNode
	for _, v := range []int{5, 3, 7, 1, 4, 6, 8} {
		root = InsertBST(root, v)
	}
	fmt.Println("BST InOrder:", InOrder(root))
	fmt.Println("BST Height:", TreeHeight(root))

	// Graph
	g := NewGraph(6)
	g.AddEdge(0, 1)
	g.AddEdge(1, 2)
	g.AddEdge(2, 3)
	g.AddEdge(3, 4)
	g.AddEdge(4, 5)
	fmt.Println("Graph BFS:", g.BFS(0))
	fmt.Println("Graph DFS:", g.DFS(0))

	// Counter
	var counter AtomicCounter
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			counter.Increment()
		}()
	}
	wg.Wait()
	fmt.Println("Atomic counter:", counter.Get())

	// Math
	floats := []float64{2, 4, 4, 4, 5, 5, 7, 9}
	fmt.Printf("Avg: %.2f, Median: %.2f, StdDev: %.2f\n",
		Average(floats), Median(floats), StdDev(floats))
	fmt.Println("IsPrime(97):", IsPrime(97))
	fmt.Println("GCD(12,8):", GCD(12, 8), "LCM:", LCM(12, 8))

	fmt.Println("=== Done ===")
}
