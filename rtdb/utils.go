package rtdb

import "sync/atomic"

type AtomicBool uint32

func (b *AtomicBool) Set(v bool) {
	if v {
		atomic.StoreUint32((*uint32)(b), 1)
	} else {
		atomic.StoreUint32((*uint32)(b), 0)
	}
}

func (b *AtomicBool) IsSet() bool {
	return atomic.LoadUint32((*uint32)(b)) > 0
}

type AtomicInt16 struct {
	value atomic.Value
}

func (i *AtomicInt16) Set(v int16) {
	i.value.Store(v)
}

func (i AtomicInt16) Get() (int16, bool) {
	if v := i.value.Load(); v != nil {
		return v.(int16), true
	}
	return -1, false
}

type AtomicError struct {
	value atomic.Value
}

func (e *AtomicError) Error() error {
	if err := e.value.Load(); err != nil {
		return err.(error)
	}

	return nil
}

func (e *AtomicError) Set(v error) {
	e.value.Store(v)
}
