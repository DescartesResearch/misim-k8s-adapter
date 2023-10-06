package storage

type Buffer[T any] interface {
	Size() int
	Empty() bool
	Items() []T
	Clear() []T
	Put(newItem T)
	PutAll(newItems []T)
}
