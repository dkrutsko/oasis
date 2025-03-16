package runtime

import (
	"os"
	"sync"
)

////////////////////////////////////////////////////////////////////////////////

type Runtime struct {
	offsetsData []byte

	strCache map[string]string
	intCache map[string]int64
	lock     sync.RWMutex
}

////////////////////////////////////////////////////////////////////////////////

func New() *Runtime {

	return &Runtime{
		offsetsData: nil,

		strCache: make(map[string]string),
		intCache: make(map[string]int64),
	}
}

////////////////////////////////////////////////////////////////////////////////

func (r *Runtime) Create() error {

	var err error
	// Try and read the contents of the offsets file
	r.offsetsData, err = os.ReadFile("./offsets.json")
	if err != nil {
		return err
	}

	return nil
}
