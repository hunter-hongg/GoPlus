package future

import (
    "context"
    "time"
)

// ==================== 接口定义 ====================

// Future 单返回值Future接口
type Future[T any] interface {
    Get() T
    GetWithTimeout(timeout time.Duration) (T, bool)
    Wait(timeout ...time.Duration) bool
    IsDone() bool
    Cancel()
    Error() error
}

// Future2 双返回值Future接口
type Future2[T1, T2 any] interface {
    Get() (T1, T2)
    GetWithTimeout(timeout time.Duration) (T1, T2, bool)
    Wait(timeout ...time.Duration) bool
    IsDone() bool
    Cancel()
    Error() error
}

// Future3 三返回值Future接口
type Future3[T1, T2, T3 any] interface {
    Get() (T1, T2, T3)
    GetWithTimeout(timeout time.Duration) (T1, T2, T3, bool)
    Wait(timeout ...time.Duration) bool
    IsDone() bool
    Cancel()
    Error() error
}

// ==================== 实现结构体 ====================

// futureImpl 单返回值实现
type futureImpl[T any] struct {
    ctx        context.Context
    cancelFunc context.CancelFunc
    result     T
    done       chan struct{}
    err        error
}

// futureImpl2 双返回值实现
type futureImpl2[T1, T2 any] struct {
    ctx        context.Context
    cancelFunc context.CancelFunc
    result1    T1
    result2    T2
    done       chan struct{}
    err        error
}

// futureImpl3 三返回值实现
type futureImpl3[T1, T2, T3 any] struct {
    ctx        context.Context
    cancelFunc context.CancelFunc
    result1    T1
    result2    T2
    result3    T3
    done       chan struct{}
    err        error
}

// ==================== 构造函数 ====================

// New 创建单返回值Future
func New[T any](fn func() T) Future[T] {
    return NewWithContext[T](context.Background(), fn)
}

// NewWithContext 创建带Context的单返回值Future
func NewWithContext[T any](ctx context.Context, fn func() T) Future[T] {
    childCtx, cancel := context.WithCancel(ctx)
    f := &futureImpl[T]{
        ctx:        childCtx,
        cancelFunc: cancel,
        done:       make(chan struct{}),
    }
    
    go f.execute(fn)
    return f
}

// New2 创建双返回值Future
func New2[T1, T2 any](fn func() (T1, T2)) Future2[T1, T2] {
    return New2WithContext[T1, T2](context.Background(), fn)
}

// New2WithContext 创建带Context的双返回值Future
func New2WithContext[T1, T2 any](ctx context.Context, fn func() (T1, T2)) Future2[T1, T2] {
    childCtx, cancel := context.WithCancel(ctx)
    f := &futureImpl2[T1, T2]{
        ctx:        childCtx,
        cancelFunc: cancel,
        done:       make(chan struct{}),
    }
    
    go f.execute(fn)
    return f
}

// New3 创建三返回值Future
func New3[T1, T2, T3 any](fn func() (T1, T2, T3)) Future3[T1, T2, T3] {
    return New3WithContext[T1, T2, T3](context.Background(), fn)
}

// New3WithContext 创建带Context的三返回值Future
func New3WithContext[T1, T2, T3 any](ctx context.Context, fn func() (T1, T2, T3)) Future3[T1, T2, T3] {
    childCtx, cancel := context.WithCancel(ctx)
    f := &futureImpl3[T1, T2, T3]{
        ctx:        childCtx,
        cancelFunc: cancel,
        done:       make(chan struct{}),
    }
    
    go f.execute(fn)
    return f
}

// NewE 创建返回(T, error)的Future
func NewE[T any](fn func() (T, error)) Future[T] {
    return NewWithContextE[T](context.Background(), fn)
}

// NewWithContextE 创建带Context的(T, error) Future
func NewWithContextE[T any](ctx context.Context, fn func() (T, error)) Future[T] {
    childCtx, cancel := context.WithCancel(ctx)
    f := &futureImpl[T]{
        ctx:        childCtx,
        cancelFunc: cancel,
        done:       make(chan struct{}),
    }
    
    go f.executeWithError(fn)
    return f
}

// New2E 创建返回(T1, T2, error)的Future
func New2E[T1, T2 any](fn func() (T1, T2, error)) Future2[T1, T2] {
    return New2WithContextE[T1, T2](context.Background(), fn)
}

// New2WithContextE 创建带Context的(T1, T2, error) Future
func New2WithContextE[T1, T2 any](ctx context.Context, fn func() (T1, T2, error)) Future2[T1, T2] {
    childCtx, cancel := context.WithCancel(ctx)
    f := &futureImpl2[T1, T2]{
        ctx:        childCtx,
        cancelFunc: cancel,
        done:       make(chan struct{}),
    }
    
    go f.executeWithError(fn)
    return f
}

// ==================== 执行方法 ====================

func (f *futureImpl[T]) execute(fn func() T) {
    defer close(f.done)
    select {
    case <-f.ctx.Done():
        f.err = f.ctx.Err()
    default:
        f.result = fn()
    }
}

func (f *futureImpl[T]) executeWithError(fn func() (T, error)) {
    defer close(f.done)
    select {
    case <-f.ctx.Done():
        f.err = f.ctx.Err()
    default:
        var err error
        f.result, err = fn()
        f.err = err
    }
}

func (f *futureImpl2[T1, T2]) execute(fn func() (T1, T2)) {
    defer close(f.done)
    select {
    case <-f.ctx.Done():
        f.err = f.ctx.Err()
    default:
        f.result1, f.result2 = fn()
    }
}

func (f *futureImpl2[T1, T2]) executeWithError(fn func() (T1, T2, error)) {
    defer close(f.done)
    select {
    case <-f.ctx.Done():
        f.err = f.ctx.Err()
    default:
        var err error
        f.result1, f.result2, err = fn()
        f.err = err
    }
}

func (f *futureImpl3[T1, T2, T3]) execute(fn func() (T1, T2, T3)) {
    defer close(f.done)
    select {
    case <-f.ctx.Done():
        f.err = f.ctx.Err()
    default:
        f.result1, f.result2, f.result3 = fn()
    }
}

// ==================== 核心方法实现 ====================

// ---- 单返回值方法 ----
func (f *futureImpl[T]) Get() T {
    <-f.done
    return f.result
}

func (f *futureImpl[T]) GetWithTimeout(timeout time.Duration) (T, bool) {
    select {
    case <-f.done:
        return f.result, true
    case <-time.After(timeout):
        var zero T
        return zero, false
    case <-f.ctx.Done():
        var zero T
        return zero, false
    }
}

// ---- 双返回值方法 ----
func (f *futureImpl2[T1, T2]) Get() (T1, T2) {
    <-f.done
    return f.result1, f.result2
}

func (f *futureImpl2[T1, T2]) GetWithTimeout(timeout time.Duration) (T1, T2, bool) {
    select {
    case <-f.done:
        return f.result1, f.result2, true
    case <-time.After(timeout):
        var zero1 T1
        var zero2 T2
        return zero1, zero2, false
    case <-f.ctx.Done():
        var zero1 T1
        var zero2 T2
        return zero1, zero2, false
    }
}

// ---- 三返回值方法 ----
func (f *futureImpl3[T1, T2, T3]) Get() (T1, T2, T3) {
    <-f.done
    return f.result1, f.result2, f.result3
}

func (f *futureImpl3[T1, T2, T3]) GetWithTimeout(timeout time.Duration) (T1, T2, T3, bool) {
    select {
    case <-f.done:
        return f.result1, f.result2, f.result3, true
    case <-time.After(timeout):
        var zero1 T1
        var zero2 T2
        var zero3 T3
        return zero1, zero2, zero3, false
    case <-f.ctx.Done():
        var zero1 T1
        var zero2 T2
        var zero3 T3
        return zero1, zero2, zero3, false
    }
}

// ==================== 通用方法 ====================

// Wait 等待完成（可带超时）
func (f *futureImpl[T]) Wait(timeout ...time.Duration) bool {
    if len(timeout) > 0 {
        select {
        case <-f.done:
            return true
        case <-time.After(timeout[0]):
            return false
        case <-f.ctx.Done():
            return false
        }
    }
    
    <-f.done
    return true
}

func (f *futureImpl2[T1, T2]) Wait(timeout ...time.Duration) bool {
    if len(timeout) > 0 {
        select {
        case <-f.done:
            return true
        case <-time.After(timeout[0]):
            return false
        case <-f.ctx.Done():
            return false
        }
    }
    
    <-f.done
    return true
}

func (f *futureImpl3[T1, T2, T3]) Wait(timeout ...time.Duration) bool {
    if len(timeout) > 0 {
        select {
        case <-f.done:
            return true
        case <-time.After(timeout[0]):
            return false
        case <-f.ctx.Done():
            return false
        }
    }
    
    <-f.done
    return true
}

// IsDone 检查是否完成
func (f *futureImpl[T]) IsDone() bool {
    select {
    case <-f.done:
        return true
    default:
        return false
    }
}

func (f *futureImpl2[T1, T2]) IsDone() bool {
    select {
    case <-f.done:
        return true
    default:
        return false
    }
}

func (f *futureImpl3[T1, T2, T3]) IsDone() bool {
    select {
    case <-f.done:
        return true
    default:
        return false
    }
}

// Cancel 取消任务
func (f *futureImpl[T]) Cancel() {
    f.cancelFunc()
}

func (f *futureImpl2[T1, T2]) Cancel() {
    f.cancelFunc()
}

func (f *futureImpl3[T1, T2, T3]) Cancel() {
    f.cancelFunc()
}

// Error 获取错误信息
func (f *futureImpl[T]) Error() error {
    <-f.done
    return f.err
}

func (f *futureImpl2[T1, T2]) Error() error {
    <-f.done
    return f.err
}

func (f *futureImpl3[T1, T2, T3]) Error() error {
    <-f.done
    return f.err
}

// ==================== 工具函数 ====================

// Async 单返回值的快捷函数
func Async[T any](fn func() T) Future[T] {
    return New(fn)
}

// AsyncE (T, error)的快捷函数
func AsyncE[T any](fn func() (T, error)) Future[T] {
    return NewE(fn)
}

// Async2 双返回值的快捷函数
func Async2[T1, T2 any](fn func() (T1, T2)) Future2[T1, T2] {
    return New2(fn)
}

// Async2E (T1, T2, error)的快捷函数
func Async2E[T1, T2 any](fn func() (T1, T2, error)) Future2[T1, T2] {
    return New2E(fn)
}

// Async3 三返回值的快捷函数
func Async3[T1, T2, T3 any](fn func() (T1, T2, T3)) Future3[T1, T2, T3] {
    return New3(fn)
}

// Then 链式调用：Future完成后执行下一个任务
func Then[T1, T2 any](f Future[T1], fn func(T1) T2) Future[T2] {
    return New(func() T2 {
        result := f.Get()
        return fn(result)
    })
}

// Then2 双返回值的链式调用
func Then2[T1, T2, R any](f Future2[T1, T2], fn func(T1, T2) R) Future[R] {
    return New(func() R {
        r1, r2 := f.Get()
        return fn(r1, r2)
    })
}

// All 等待所有Future完成（单返回值）
func All[T any](futures ...Future[T]) Future[[]T] {
    return New(func() []T {
        results := make([]T, len(futures))
        for i, f := range futures {
            results[i] = f.Get()
        }
        return results
    })
}

// Any 等待任意一个Future完成（单返回值）
func Any[T any](futures ...Future[T]) Future[T] {
    return New(func() T {
        done := make(chan T, 1)
        
        for _, f := range futures {
            go func(future Future[T]) {
                done <- future.Get()
            }(f)
        }
        
        return <-done
    })
}

// Map 对Future结果进行转换
func Map[T, R any](f Future[T], fn func(T) R) Future[R] {
    return New(func() R {
        return fn(f.Get())
    })
}
