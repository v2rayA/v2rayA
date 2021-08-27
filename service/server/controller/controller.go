package controller

import (
	"fmt"
	"sync"
)

var (
	updating      bool
	updatingMu    sync.Mutex
	processingErr = fmt.Errorf("the last request is being processed")
)
