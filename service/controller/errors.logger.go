package controller

import "v2ray.com/core/common/errors"

func logError(base error, info ...string) (err *errors.Error) {
	if len(info) == 0 {
		err = errors.New(base)
	} else {
		err = errors.New(info).Base(base)
	}
	err.AtWarning().WriteToLog()
	return
}
