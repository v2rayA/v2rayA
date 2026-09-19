package service

import "github.com/v2rayA/v2rayA/kernel/v2ray"

type ApplyCoreConfigError struct {
	UpdateErr        error
	RestoreStoreErr  error
	RestoreUpdateErr error
}

func (e *ApplyCoreConfigError) Error() string {
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
