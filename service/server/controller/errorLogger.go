package controller

import (
	"fmt"

	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/pkg/util/log"
)

func logError(err interface{}) error {
	e, ok := err.(error)
	if !ok {
		e = fmt.Errorf("%v", err)
	}
	log.Error("%v", e)
	return e
}

func badRequest(field string, err interface{}) error {
	return common.Coded("BAD_REQUEST", logError(err), map[string]interface{}{"field": field})
}
