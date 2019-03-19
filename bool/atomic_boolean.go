package atomicbool

import "sync/atomic"

type atomicBool int32

// Bool is an atomic Boolean,
// Its methods are all atomic, thus safe to be called by
// multiple goroutines simultaneously.
type Bool interface {
	Set()
	Unset()
	IsSet() bool
}

func NewBool(ok bool) Bool {
	atomicBoolean := new(atomicBool)
	if ok {
		atomicBoolean.Set()
	}
	return atomicBoolean
}

func (self *atomicBool) Set() {
	atomic.StoreInt32((*int32)(self), 1)
}

func (self *atomicBool) Unset() {
	atomic.StoreInt32((*int32)(self), 0)
}

// IsSet returns whether the Boolean is true
func (self *atomicBool) IsSet() bool {
	return (atomic.LoadInt32((*int32)(self)) == 1)
}

func (self *atomicBool) Value() bool {
	return atomic.LoadInt32((*int32)(self))
}
