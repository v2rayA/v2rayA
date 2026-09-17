package controller

import (
	"fmt"
	"sync"

	"github.com/v2rayA/v2rayA/common"
)

var (
	updating      bool
	updatingMu    sync.Mutex
	processingErr = common.Coded("REQUEST_IN_PROGRESS", fmt.Errorf("the last request is being processed"), nil)
)
