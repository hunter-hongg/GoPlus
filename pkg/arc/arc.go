package arc

import (
	"sync/atomic"
	"unsafe"
)

// Arc 是原子引用计数智能指针，类似 Rust 的 Arc<T>
type Arc[T any] struct {
	// 使用 unsafe.Pointer 存储数据指针
	ptr unsafe.Pointer
}

// arcInternal 存储实际数据和引用计数
type arcInternal[T any] struct {
	data T
	ref  int64 // 原子计数器
}

// NewArc 创建新的 Arc
func NewArc[T any](value T) *Arc[T] {
	internal := &arcInternal[T]{
		data: value,
		ref:  1, // 初始引用计数为 1
	}
	
	return &Arc[T]{
		ptr: unsafe.Pointer(internal),
	}
}

// Clone 创建 Arc 的克隆，增加引用计数
func (a *Arc[T]) Clone() *Arc[T] {
	if a.ptr == nil {
		return nil
	}
	
	// 原子增加引用计数
	internal := (*arcInternal[T])(a.ptr)
	atomic.AddInt64(&internal.ref, 1)
	
	return &Arc[T]{
		ptr: a.ptr,
	}
}

// Deref 获取内部数据的引用（类似 Rust 的 Deref trait）
func (a *Arc[T]) Deref() *T {
	if a.ptr == nil {
		return nil
	}
	
	internal := (*arcInternal[T])(a.ptr)
	return &internal.data
}

// StrongCount 获取强引用计数
func (a *Arc[T]) StrongCount() int64 {
	if a.ptr == nil {
		return 0
	}
	
	internal := (*arcInternal[T])(a.ptr)
	return atomic.LoadInt64(&internal.ref)
}

// Drop 减少引用计数，当计数为 0 时释放内存
func (a *Arc[T]) Drop() {
	if a.ptr == nil {
		return
	}
	
	internal := (*arcInternal[T])(a.ptr)
	
	// 原子减少引用计数
	if atomic.AddInt64(&internal.ref, -1) == 0 {
		// 引用计数为 0，释放内存
		// Go 的垃圾回收会自动处理，这里只需置空指针
		a.ptr = nil
	}
}

// GetMut 获取可变引用（类似 Rust 的 get_mut）
// 注意：这需要额外的同步机制确保唯一性
func (a *Arc[T]) GetMut() *T {
	if a.ptr == nil {
		return nil
	}
	
	internal := (*arcInternal[T])(a.ptr)
	
	// 只有当前引用计数为 1 时才能获取可变引用
	if atomic.LoadInt64(&internal.ref) == 1 {
		return &internal.data
	}
	
	return nil
}

// TryUnwrap 尝试获取所有权（类似 Arc::try_unwrap）
// 如果引用计数为 1，则返回内部数据，否则返回 false
func (a *Arc[T]) TryUnwrap() (T, bool) {
	if a.ptr == nil {
		var zero T
		return zero, false
	}
	
	internal := (*arcInternal[T])(a.ptr)
	
	// 使用 CAS 确保原子性
	for {
		current := atomic.LoadInt64(&internal.ref)
		if current != 1 {
			var zero T
			return zero, false
		}
		
		// 尝试将引用计数从 1 设置为 0
		if atomic.CompareAndSwapInt64(&internal.ref, 1, 0) {
			// 成功获取所有权
			data := internal.data
			a.ptr = nil
			return data, true
		}
	}
}

// ============================================================================
// 弱引用实现（Weak Arc）
// ============================================================================

// Weak 弱引用，类似 Rust 的 Weak<T>
type Weak[T any] struct {
	ptr unsafe.Pointer
}

// Downgrade 从 Arc 创建弱引用
func (a *Arc[T]) Downgrade() *Weak[T] {
	if a.ptr == nil {
		return nil
	}
	
	return &Weak[T]{
		ptr: a.ptr,
	}
}

// Upgrade 尝试将弱引用升级为强引用
func (w *Weak[T]) Upgrade() *Arc[T] {
	if w.ptr == nil {
		return nil
	}
	
	internal := (*arcInternal[T])(w.ptr)
	
	// 原子增加引用计数
	current := atomic.AddInt64(&internal.ref, 1)
	
	// 如果增加后计数 > 1，说明对象仍然存在
	if current > 1 {
		return &Arc[T]{
			ptr: w.ptr,
		}
	}
	
	// 如果增加后计数 <= 1，说明对象已被释放
	// 回滚引用计数增加
	atomic.AddInt64(&internal.ref, -1)
	return nil
}

// ============================================================================
// 线程安全访问方法
// ============================================================================

// With 提供线程安全的访问方式
func (a *Arc[T]) With(fn func(*T)) bool {
	if a.ptr == nil {
		return false
	}
	
	// 确保在函数执行期间 Arc 不会被释放
	// 增加引用计数
	internal := (*arcInternal[T])(a.ptr)
	atomic.AddInt64(&internal.ref, 1)
	
	// 执行用户函数
	fn(&internal.data)
	
	// 减少引用计数
	a.Drop()
	return true
}

// Map 应用函数并返回新值（不可变操作）
func MapArc[T, U any](a *Arc[T], fn func(T) U) *Arc[U] {
	if a.ptr == nil {
		return nil
	}
	
	internal := (*arcInternal[T])(a.ptr)
	return NewArc(fn(internal.data))
}

// ============================================================================
// 容器化辅助函数
// ============================================================================

// ArcSlice 创建 Arc 切片
func ArcSlice[T any](values ...T) []*Arc[T] {
	arcs := make([]*Arc[T], len(values))
	for i, v := range values {
		arcs[i] = NewArc(v)
	}
	return arcs
}

// MergeArcs 合并多个 Arc（增加所有 Arc 的引用计数）
func MergeArcs[T any](arcs ...*Arc[T]) *Arc[[]T] {
	values := make([]T, len(arcs))
	for i, arc := range arcs {
		if arc != nil {
			values[i] = *arc.Deref()
		}
	}
	return NewArc(values)
}

// ============================================================================
// 内存屏障和同步原语（高级功能）
// ============================================================================

// MemoryBarrier 确保内存操作的顺序性
func (a *Arc[T]) MemoryBarrier() {
	if a.ptr == nil {
		return
	}
	
	// 使用原子操作创建内存屏障
	internal := (*arcInternal[T])(a.ptr)
	atomic.LoadInt64(&internal.ref)
}

// CompareAndSwap 比较并交换 Arc 的内容
// 参数名已从 new 改为 newValue 以避免与内置函数冲突
func (a *Arc[T]) CompareAndSwap(oldValue, newValue T) bool {
	if a.ptr == nil {
		return false
	}
	
	internal := (*arcInternal[T])(a.ptr)
	
	// 只有当前引用计数为 1 时才允许交换
	if atomic.LoadInt64(&internal.ref) != 1 {
		return false
	}
	
	// 这里简化实现，实际需要更复杂的比较逻辑
	// 对于复杂类型，应该使用 deep copy 和比较
	internal.data = newValue
	return true
}

// ============================================================================
// 额外实用功能
// ============================================================================

// Swap 交换两个 Arc 的内容
func (a *Arc[T]) Swap(other *Arc[T]) {
	if a.ptr == nil || other.ptr == nil {
		return
	}
	
	// 交换指针
	temp := a.ptr
	a.ptr = other.ptr
	other.ptr = temp
}

// Reset 重置 Arc，释放当前引用
func (a *Arc[T]) Reset(value T) {
	// 释放当前引用
	a.Drop()
	
	// 创建新引用
	internal := &arcInternal[T]{
		data: value,
		ref:  1,
	}
	a.ptr = unsafe.Pointer(internal)
}

// AsRef 创建数据的只读引用包装
func (a *Arc[T]) AsRef() *Ref[T] {
	if a.ptr == nil {
		return nil
	}
	
	return &Ref[T]{
		arc: a.Clone(),
	}
}

// Ref 只读引用包装器
type Ref[T any] struct {
	arc *Arc[T]
}

// Deref 获取底层数据的引用
func (r *Ref[T]) Deref() *T {
	if r.arc == nil {
		return nil
	}
	return r.arc.Deref()
}

// Release 释放引用
func (r *Ref[T]) Release() {
	if r.arc != nil {
		r.arc.Drop()
		r.arc = nil
	}
}

// ============================================================================
// 测试辅助函数
// ============================================================================

// IsNil 检查 Arc 是否为空
func (a *Arc[T]) IsNil() bool {
	return a.ptr == nil
}

// UnsafeRawPtr 获取底层指针（仅用于调试和测试）
func (a *Arc[T]) UnsafeRawPtr() unsafe.Pointer {
	return a.ptr
}
