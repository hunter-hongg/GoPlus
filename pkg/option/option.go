package option

// Option 表示一个可选值，类似于Rust的Option
type Option[T any] interface {
    IsSome() bool
    IsNone() bool
    Unwrap() T
    UnwrapOr(defaultValue T) T
    UnwrapOrElse(f func() T) T
    Expect(msg string) T
    Map[U any](f func(T) U) Option[U]
    AndThen[U any](f func(T) Option[U]) Option[U]
    Or(other Option[T]) Option[T]
    And(other Option[T]) Option[T]
    ToResult[E any](err E) Result[T, E]
}

// Some 表示有值
type some[T any] struct {
    value T
}

func (s *some[T]) IsSome() bool { return true }
func (s *some[T]) IsNone() bool { return false }
func (s *some[T]) Unwrap() T    { return s.value }
func (s *some[T]) UnwrapOr(defaultValue T) T { return s.value }
func (s *some[T]) UnwrapOrElse(f func() T) T { return s.value }
func (s *some[T]) Expect(msg string) T { return s.value }

func (s *some[T]) Map[U any](f func(T) U) Option[U] {
    return Some[U](f(s.value))
}

func (s *some[T]) AndThen[U any](f func(T) Option[U]) Option[U] {
    return f(s.value)
}

func (s *some[T]) Or(other Option[T]) Option[T] {
    return s
}

func (s *some[T]) And(other Option[T]) Option[T] {
    return other
}

func (s *some[T]) ToResult[E any](err E) Result[T, E] {
    return Ok[T, E](s.value)
}

// None 表示无值
type none[T any] struct{}

func (n *none[T]) IsSome() bool { return false }
func (n *none[T]) IsNone() bool { return true }
func (n *none[T]) Unwrap() T    { panic("called `Unwrap()` on a `None` value") }
func (n *none[T]) UnwrapOr(defaultValue T) T { return defaultValue }
func (n *none[T]) UnwrapOrElse(f func() T) T { return f() }
func (n *none[T]) Expect(msg string) T { panic(msg) }

func (n *none[T]) Map[U any](f func(T) U) Option[U] {
    return None[U]()
}

func (n *none[T]) AndThen[U any](f func(T) Option[U]) Option[U] {
    return None[U]()
}

func (n *none[T]) Or(other Option[T]) Option[T] {
    return other
}

func (n *none[T]) And(other Option[T]) Option[T] {
    return n
}

func (n *none[T]) ToResult[E any](err E) Result[T, E] {
    return Err[T, E](err)
}

// Some 创建有值的Option
func Some[T any](value T) Option[T] {
    return &some[T]{value: value}
}

// None 创建无值的Option
func None[T any]() Option[T] {
    return &none[T]{}
}

// ============================================================================
// Result 类型实现
// ============================================================================

// Result 表示操作结果，可以是成功(Ok)或失败(Err)，类似于Rust的Result
type Result[T any, E any] interface {
    IsOk() bool
    IsErr() bool
    Unwrap() T
    UnwrapOr(defaultValue T) T
    UnwrapOrElse(f func(E) T) T
    Expect(msg string) T
    UnwrapErr() E
    Map[U any](f func(T) U) Result[U, E]
    MapErr[F any](f func(E) F) Result[T, F]
    AndThen[U any](f func(T) Result[U, E]) Result[U, E]
    OrElse[F any](f func(E) Result[T, F]) Result[T, F]
    ToOption() Option[T]
}

// ok 表示成功结果
type ok[T any, E any] struct {
    value T
}

func (o *ok[T, E]) IsOk() bool  { return true }
func (o *ok[T, E]) IsErr() bool { return false }
func (o *ok[T, E]) Unwrap() T   { return o.value }
func (o *ok[T, E]) UnwrapOr(defaultValue T) T { return o.value }
func (o *ok[T, E]) UnwrapOrElse(f func(E) T) T { return o.value }
func (o *ok[T, E]) Expect(msg string) T { return o.value }
func (o *ok[T, E]) UnwrapErr() E { panic("called `UnwrapErr()` on an `Ok` value") }

func (o *ok[T, E]) Map[U any](f func(T) U) Result[U, E] {
    return Ok[U, E](f(o.value))
}

func (o *ok[T, E]) MapErr[F any](f func(E) F) Result[T, F] {
    return Ok[T, F](o.value)
}

func (o *ok[T, E]) AndThen[U any](f func(T) Result[U, E]) Result[U, E] {
    return f(o.value)
}

func (o *ok[T, E]) OrElse[F any](f func(E) Result[T, F]) Result[T, F] {
    return Ok[T, F](o.value)
}

func (o *ok[T, E]) ToOption() Option[T] {
    return Some[T](o.value)
}

// err 表示失败结果
type err[T any, E any] struct {
    error E
}

func (e *err[T, E]) IsOk() bool  { return false }
func (e *err[T, E]) IsErr() bool { return true }
func (e *err[T, E]) Unwrap() T   { panic("called `Unwrap()` on an `Err` value") }
func (e *err[T, E]) UnwrapOr(defaultValue T) T { return defaultValue }
func (e *err[T, E]) UnwrapOrElse(f func(E) T) T { return f(e.error) }
func (e *err[T, E]) Expect(msg string) T { panic(msg) }
func (e *err[T, E]) UnwrapErr() E { return e.error }

func (e *err[T, E]) Map[U any](f func(T) U) Result[U, E] {
    return Err[U, E](e.error)
}

func (e *err[T, E]) MapErr[F any](f func(E) F) Result[T, F] {
    return Err[T, F](f(e.error))
}

func (e *err[T, E]) AndThen[U any](f func(T) Result[U, E]) Result[U, E] {
    return Err[U, E](e.error)
}

func (e *err[T, E]) OrElse[F any](f func(E) Result[T, F]) Result[T, F] {
    return f(e.error)
}

func (e *err[T, E]) ToOption() Option[T] {
    return None[T]()
}

// Ok 创建成功结果
func Ok[T any, E any](value T) Result[T, E] {
    return &ok[T, E]{value: value}
}

// Err 创建失败结果
func Err[T any, E any](error E) Result[T, E] {
    return &err[T, E]{error: error}
}

// ============================================================================
// 工具函数
// ============================================================================

// MatchOption 模式匹配Option
func MatchOption[T any, U any](opt Option[T], someFunc func(T) U, noneFunc func() U) U {
    if opt.IsSome() {
        return someFunc(opt.Unwrap())
    }
    return noneFunc()
}

// MatchResult 模式匹配Result
func MatchResult[T any, E any, U any](res Result[T, E], okFunc func(T) U, errFunc func(E) U) U {
    if res.IsOk() {
        return okFunc(res.Unwrap())
    }
    return errFunc(res.UnwrapErr())
}
