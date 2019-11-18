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
	if t != nil { //有备份过iptablets，清理一下
		if err = t.RestoreRules(); err != nil {
			return
		}
	}
	return
}
