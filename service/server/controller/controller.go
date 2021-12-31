package controller

import (
	"fmt"
	"sync"
)

var (
	updating      bool
	updatingMu    sync.Mutex
	processingErr = fmt.Errorf("Processing last request")
)
