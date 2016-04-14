package testdata

import (
	"sync"
	"sync/atomic"
)

func OkFunc() {
	var x *sync.Mutex
	p := x
	var y sync.Mutex
	p = &y

	var z = sync.Mutex{}
	w := sync.Mutex{}

	w = sync.Mutex{}
	q := struct{ L sync.Mutex }{
		L: sync.Mutex{},
	}

	yy := []Tlock{
		Tlock{},
		Tlock{
			once: sync.Once{},
		},
	}

	nl := new(sync.Mutex)
	mx := make([]sync.Mutex, 10)
	xx := struct{ L *sync.Mutex }{
		L: new(sync.Mutex),
	}
}

type Tlock struct {
	once sync.Once
}

func BadFunc() {
	var x *sync.Mutex
	p := x
	var y sync.Mutex
	p = &y
	*p = *x // ERROR "assignment copies lock value to \*p: sync.Mutex"

	var t Tlock
	var tp *Tlock
	tp = &t
	*tp = t // ERROR "assignment copies lock value to \*tp: testdata.Tlock contains sync.Once contains sync.Mutex"
	t = *tp // ERROR "assignment copies lock value to t: testdata.Tlock contains sync.Once contains sync.Mutex"

	y := *x   // ERROR "assignment copies lock value to y: sync.Mutex"
	var z = t // ERROR "variable declaration copies lock value to z: testdata.Tlock contains sync.Once contains sync.Mutex"

	w := struct{ L sync.Mutex }{
		L: *x, // ERROR "literal copies lock value from \*x: sync.Mutex"
	}
	var q = map[int]Tlock{
		1: t,   // ERROR "literal copies lock value from t: testdata.Tlock contains sync.Once contains sync.Mutex"
		2: *tp, // ERROR "literal copies lock value from \*tp: testdata.Tlock contains sync.Once contains sync.Mutex"
	}
	yy := []Tlock{
		t,   // ERROR "literal copies lock value from t: testdata.Tlock contains sync.Once contains sync.Mutex"
		*tp, // ERROR "literal copies lock value from \*tp: testdata.Tlock contains sync.Once contains sync.Mutex"
	}

	// override 'new' keyword
	new := func(interface{}) {}
	new(t) // ERROR "function call copies lock value: testdata.Tlock contains sync.Once contains sync.Mutex"
}

// SyncTypesCheck checks copying of sync.* types except sync.Mutex
func SyncTypesCheck() {
	// sync.RWMutex copying
	var rwmuX sync.RWMutex
	var rwmuXX = sync.RWMutex{}
	rwmuY := rwmuX     // ERROR "assignment copies lock value to rwmuY: sync.RWMutex"
	rwmuY = rwmuX      // ERROR "assignment copies lock value to rwmuY: sync.RWMutex"
	var rwmuYY = rwmuX // ERROR "variable declaration copies lock value to rwmuYY: sync.RWMutex"
	rwmuP := &rwmuX
	rwmuZ := &sync.RWMutex{}

	// sync.Cond copying
	var condX sync.Cond
	var condXX = sync.Cond{}
	condY := condX     // ERROR "assignment copies lock value to condY: sync.Cond contains internal/nocopy.NoCopy"
	condY = condX      // ERROR "assignment copies lock value to condY: sync.Cond contains internal/nocopy.NoCopy"
	var condYY = condX // ERROR "variable declaration copies lock value to condYY: sync.Cond contains internal/nocopy.NoCopy"
	condP := &condX
	condZ := &sync.Cond{
		L: &sync.Mutex{},
	}
	condZ = sync.NewCond(&sync.Mutex{})

	// sync.WaitGroup copying
	var wgX sync.WaitGroup
	var wgXX = sync.WaitGroup{}
	wgY := wgX     // ERROR "assignment copies lock value to wgY: sync.WaitGroup contains internal/nocopy.NoCopy"
	wgY = wgX      // ERROR "assignment copies lock value to wgY: sync.WaitGroup contains internal/nocopy.NoCopy"
	var wgYY = wgX // ERROR "variable declaration copies lock value to wgYY: sync.WaitGroup contains internal/nocopy.NoCopy"
	wgP := &wgX
	wgZ := &sync.WaitGroup{}

	// sync.Pool copying
	var poolX sync.Pool
	var poolXX = sync.Pool{}
	poolY := poolX     // ERROR "assignment copies lock value to poolY: sync.Pool contains internal/nocopy.NoCopy"
	poolY = poolX      // ERROR "assignment copies lock value to poolY: sync.Pool contains internal/nocopy.NoCopy"
	var poolYY = poolX // ERROR "variable declaration copies lock value to poolYY: sync.Pool contains internal/nocopy.NoCopy"
	poolP := &poolX
	poolZ := &sync.Pool{}

	// sync.Once copying
	var onceX sync.Once
	var onceXX = sync.Once{}
	onceY := onceX     // ERROR "assignment copies lock value to onceY: sync.Once contains sync.Mutex"
	onceY = onceX      // ERROR "assignment copies lock value to onceY: sync.Once contains sync.Mutex"
	var onceYY = onceX // ERROR "variable declaration copies lock value to onceYY: sync.Once contains sync.Mutex"
	onceP := &onceX
	onceZ := &sync.Once{}
}

// AtomicTypesCheck checks copying of sync/atomic types
func AtomicTypesCheck() {
	// atomic.Value copying
	var vX atomic.Value
	var vXX = atomic.Value{}
	vY := vX     // ERROR "assignment copies lock value to vY: sync/atomic.Value contains internal/nocopy.NoCopy"
	vY = vX      // ERROR "assignment copies lock value to vY: sync/atomic.Value contains internal/nocopy.NoCopy"
	var vYY = vX // ERROR "variable declaration copies lock value to vYY: sync/atomic.Value contains internal/nocopy.NoCopy"
	vP := &vX
	vZ := &atomic.Value{}
}
