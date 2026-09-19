package service

import (
	"fmt"

	"github.com/v2rayA/v2rayA/kernel/v2ray"
)

type ApplyCoreConfigError struct {
	UpdateErr        error
	RestoreStoreErr  error
	RestoreUpdateErr error
}

// Error names the step that failed after the update did: a caller that
// prints it tells the user whether the previous value and the core came back.
func (e *ApplyCoreConfigError) Error() string {
	switch {
	case e.RestoreStoreErr != nil:
		return fmt.Sprintf("%v; the previous value could not be restored (%v)", e.UpdateErr, e.RestoreStoreErr)
	case e.RestoreUpdateErr != nil:
		return fmt.Sprintf("%v; the core could not be restarted with the previous value (%v)", e.UpdateErr, e.RestoreUpdateErr)
	}
	return e.UpdateErr.Error()
}

func (e *ApplyCoreConfigError) Unwrap() error {
	return e.UpdateErr
}

func (e *ApplyCoreConfigError) Restored() bool {
	return e.RestoreStoreErr == nil && e.RestoreUpdateErr == nil
}

func ApplyCoreConfig(snapshot func() (restore func() error), store func() error) error {
	restore := snapshot()
	if err := store(); err != nil {
		return err
	}
	if !v2ray.ProcessManager.Running() {
		return nil
	}
	updateErr := v2ray.UpdateV2RayConfig()
	if updateErr == nil {
		return nil
	}
	failure := &ApplyCoreConfigError{UpdateErr: updateErr}
	if failure.RestoreStoreErr = restore(); failure.RestoreStoreErr == nil {
		failure.RestoreUpdateErr = v2ray.UpdateV2RayConfig()
	}
	return failure
}
