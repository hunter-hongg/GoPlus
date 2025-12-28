package option

// Option 表示一个可选值，类似于Rust的Option
type Option[T any] struct {
    value   T
    present bool
}

// 基本操作
func (o Option[T]) IsSome() bool { return o.present }
func (o Option[T]) IsNone() bool { return !o.present }

func (o Option[T]) Unwrap() T {
    if !o.present {
        panic("called `Unwrap()` on a `None` value")
    }
    return o.value
}

func (o Option[T]) UnwrapOr(defaultValue T) T {
    if o.present {
        return o.value
    }
    return defaultValue
}

func (o Option[T]) UnwrapOrElse(f func() T) T {
    if o.present {
        return o.value
    }
    return f()
}

func (o Option[T]) Expect(msg string) T {
    if !o.present {
        panic(msg)
    }
    return o.value
}

// 独立函数版本的各种操作

// Map 将 Option[T] 转换为 Option[U]
func Map[T, U any](opt Option[T], f func(T) U) Option[U] {
    if !opt.present {
        return Option[U]{present: false}
    }
    return Option[U]{
        value:   f(opt.value),
        present: true,
    }
}

// MapOr 将 Option[T] 转换为 U，提供默认值
func MapOr[T, U any](opt Option[T], defaultValue U, f func(T) U) U {
    if !opt.present {
        return defaultValue
    }
    return f(opt.value)
}

// MapOrElse 将 Option[T] 转换为 U，提供默认函数
func MapOrElse[T, U any](opt Option[T], defaultFn func() U, f func(T) U) U {
    if !opt.present {
        return defaultFn()
    }
    return f(opt.value)
}

// AndThen 链式调用返回 Option 的函数
func AndThen[T, U any](opt Option[T], f func(T) Option[U]) Option[U] {
    if !opt.present {
        return Option[U]{present: false}
    }
    return f(opt.value)
}

// Or 返回第一个 Some 的 Option，否则返回 other
func Or[T any](opt Option[T], other Option[T]) Option[T] {
    if opt.present {
        return opt
    }
    return other
}

// OrElse 返回第一个 Some 的 Option，否则调用函数
func OrElse[T any](opt Option[T], f func() Option[T]) Option[T] {
    if opt.present {
        return opt
    }
    return f()
}

// And 如果两个都是 Some，返回第二个，否则返回 None
func And[T any](opt Option[T], other Option[T]) Option[T] {
    if opt.present {
        return other
    }
    return Option[T]{present: false}
}

// Filter 如果值满足谓词则返回 Some，否则返回 None
func Filter[T any](opt Option[T], predicate func(T) bool) Option[T] {
    if opt.present && predicate(opt.value) {
        return opt
    }
    return Option[T]{present: false}
}

// ToResult 将 Option 转换为 Result
func ToResult[T, E any](opt Option[T], err E) Result[T, E] {
    if opt.present {
        return Result[T, E]{
            value: opt.value,
            ok:    true,
        }
    }
    return Result[T, E]{
        err: err,
        ok:  false,
    }
}

// 创建函数
func Some[T any](value T) Option[T] {
    return Option[T]{
        value:   value,
        present: true,
    }
}

func None[T any]() Option[T] {
    return Option[T]{present: false}
}

// ============================================================================
// Result 类型实现（同样采用独立函数方法）
// ============================================================================

// Result 表示操作结果，可以是成功(Ok)或失败(Err)
type Result[T any, E any] struct {
    value T
    err   E
    ok    bool
}

// 基本操作
func (r Result[T, E]) IsOk() bool  { return r.ok }
func (r Result[T, E]) IsErr() bool { return !r.ok }

func (r Result[T, E]) Unwrap() T {
    if !r.ok {
        panic("called `Unwrap()` on an `Err` value")
    }
    return r.value
}

func (r Result[T, E]) UnwrapOr(defaultValue T) T {
    if r.ok {
        return r.value
    }
    return defaultValue
}

func (r Result[T, E]) UnwrapOrElse(f func(E) T) T {
    if r.ok {
        return r.value
    }
    return f(r.err)
}

func (r Result[T, E]) Expect(msg string) T {
    if !r.ok {
        panic(msg)
    }
    return r.value
}

func (r Result[T, E]) UnwrapErr() E {
    if r.ok {
        panic("called `UnwrapErr()` on an `Ok` value")
    }
    return r.err
}

// 独立函数版本的各种操作

// MapResult 将 Result[T, E] 转换为 Result[U, E]
func MapResult[T, U, E any](res Result[T, E], f func(T) U) Result[U, E] {
    if !res.ok {
        return Result[U, E]{
            err: res.err,
            ok:  false,
        }
    }
    return Result[U, E]{
        value: f(res.value),
        ok:    true,
    }
}

// MapErr 将 Result[T, E] 转换为 Result[T, F]
func MapErr[T, E, F any](res Result[T, E], f func(E) F) Result[T, F] {
    if !res.ok {
        return Result[T, F]{
            err: f(res.err),
            ok:  false,
        }
    }
    return Result[T, F]{
        value: res.value,
        ok:    true,
    }
}

// AndThenResult 链式调用返回 Result 的函数
func AndThenResult[T, U, E any](res Result[T, E], f func(T) Result[U, E]) Result[U, E] {
    if !res.ok {
        return Result[U, E]{
            err: res.err,
            ok:  false,
        }
    }
    return f(res.value)
}

// OrElseResult 处理错误情况
func OrElseResult[T, E, F any](res Result[T, E], f func(E) Result[T, F]) Result[T, F] {
    if res.ok {
        return Result[T, F]{
            value: res.value,
            ok:    true,
        }
    }
    return f(res.err)
}

// ToOption 将 Result 转换为 Option
func (r Result[T, E]) ToOption() Option[T] {
    if r.ok {
        return Some(r.value)
    }
    return None[T]()
}

// 创建函数
func Ok[T any, E any](value T) Result[T, E] {
    return Result[T, E]{
        value: value,
        ok:    true,
    }
}

func Err[T any, E any](err E) Result[T, E] {
    return Result[T, E]{
        err: err,
        ok:  false,
    }
}

// ============================================================================
// 工具函数
// ============================================================================

// MatchOption 模式匹配Option
func MatchOption[T any, U any](opt Option[T], someFunc func(T) U, noneFunc func() U) U {
    if opt.present {
        return someFunc(opt.value)
    }
    return noneFunc()
}

// MatchResult 模式匹配Result
func MatchResult[T any, E any, U any](res Result[T, E], okFunc func(T) U, errFunc func(E) U) U {
    if res.ok {
        return okFunc(res.value)
    }
    return errFunc(res.err)
}

// Zip 将两个 Option 合并为一个 Option 对
func Zip[T, U any](a Option[T], b Option[U]) Option[struct {
    First  T
    Second U
}] {
    if a.present && b.present {
        return Some(struct {
            First  T
            Second U
        }{
            First:  a.value,
            Second: b.value,
        })
    }
    return None[struct {
        First  T
        Second U
    }]()
}

// Unzip 将 Option 对分解为两个 Option
func Unzip[T, U any](opt Option[struct {
    First  T
    Second U
}]) (Option[T], Option[U]) {
    if opt.present {
        return Some(opt.value.First), Some(opt.value.Second)
    }
    return None[T](), None[U]()
}

// FromResult 将 Result 转换为 Option（忽略错误）
func FromResult[T, E any](res Result[T, E]) Option[T] {
    if res.ok {
        return Some(res.value)
    }
    return None[T]()
}

// FromPtr 从指针创建 Option
func FromPtr[T any](ptr *T) Option[T] {
    if ptr != nil {
        return Some(*ptr)
    }
    return None[T]()
}

// ToPtr 将 Option 转换为指针
func ToPtr[T any](opt Option[T]) *T {
    if opt.present {
        return &opt.value
    }
    return nil
}

// Flatten 将 Option[Option[T]] 转换为 Option[T]
func Flatten[T any](opt Option[Option[T]]) Option[T] {
    if opt.present {
        return opt.value
    }
    return None[T]()
}
