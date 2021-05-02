package controller

import "github.com/v2rayA/v2rayA/common/errors"

func logError(base error, info ...interface{}) (err *errors.Error) {
	if len(info) == 0 {
		err = errors.New(base)
	} else {
		err = errors.New(info...).Base(base)
	}
	err.AtWarning().WriteToLog()
	return
}
