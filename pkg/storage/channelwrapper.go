package storage

type ChannelWrapper[T any] interface {
	InitChannel() chan T
	Get() chan T
}
