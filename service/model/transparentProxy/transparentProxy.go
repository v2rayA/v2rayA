package transparentProxy

func StartTransparentProxy() (t *IpTablesMangle,err error) {
	t = new(IpTablesMangle)
	if err = t.BackupRules(); err != nil {
		return
	}
	if err = t.WriteRules(); err != nil {
		_ = t.RestoreRules()
		return
	}
	return
}

func StopTransparentProxy(t *IpTablesMangle) (err error) {
	if err = t.RestoreRules(); err != nil {
		return
	}
	return
}
