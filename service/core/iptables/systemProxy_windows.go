//go:build windows
// +build windows

package iptables

import (
	"fmt"
	"github.com/v2rayA/v2rayA/pkg/util/log"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
	"os/exec"
	"syscall"
	"time"
	"unsafe"
)

const (
	errnoERROR_IO_PENDING            = 997
	INTERNET_OPTION_SETTINGS_CHANGED = 39
)

var (
	errERROR_IO_PENDING    error = syscall.Errno(errnoERROR_IO_PENDING)
	modwininet                   = syscall.NewLazyDLL("wininet.dll")
	procInternetSetOptionW       = modwininet.NewProc("InternetSetOptionW")
)

// errnoErr returns common boxed Errno values, to prevent
// allocations at runtime.
func errnoErr(e syscall.Errno) error {
	switch e {
	case 0:
		return nil
	case errnoERROR_IO_PENDING:
		return errERROR_IO_PENDING
	}
	// TODO: add more here, after collecting data on the common
	// error values see on Windows. (perhaps when running
	// all.bat?)
	return e
}

var (
	modadvapi32 = windows.NewLazySystemDLL("advapi32.dll")
	modkernel32 = windows.NewLazySystemDLL("kernel32.dll")

	procGetCurrentThread = modkernel32.NewProc("GetCurrentThread")
	procOpenThreadToken  = modadvapi32.NewProc("OpenThreadToken")
	procImpersonateSelf  = modadvapi32.NewProc("ImpersonateSelf")
	procRevertToSelf     = modadvapi32.NewProc("RevertToSelf")
)

func GetCurrentThread() (pseudoHandle windows.Handle, err error) {
	r0, _, e1 := syscall.Syscall(procGetCurrentThread.Addr(), 0, 0, 0, 0)
	pseudoHandle = windows.Handle(r0)
	if pseudoHandle == 0 {
		if e1 != 0 {
			err = errnoErr(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func OpenThreadToken(h windows.Handle, access uint32, self bool, token *windows.Token) (err error) {
	var _p0 uint32
	if self {
		_p0 = 1
	} else {
		_p0 = 0
	}
	r1, _, e1 := syscall.Syscall6(procOpenThreadToken.Addr(), 4, uintptr(h), uintptr(access), uintptr(_p0), uintptr(unsafe.Pointer(token)), 0, 0)
	if r1 == 0 {
		if e1 != 0 {
			err = errnoErr(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func ImpersonateSelf() (err error) {
	r0, _, e1 := syscall.Syscall(procImpersonateSelf.Addr(), 1, uintptr(2), 0, 0)
	if r0 == 0 {
		if e1 != 0 {
			err = errnoErr(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func RevertToSelf() (err error) {
	r0, _, e1 := syscall.Syscall(procRevertToSelf.Addr(), 0, 0, 0, 0)
	if r0 == 0 {
		if e1 != 0 {
			err = errnoErr(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func OpenCurrentThreadToken() (windows.Token, error) {
	if e := ImpersonateSelf(); e != nil {
		return 0, e
	}
	defer RevertToSelf()
	t, e := GetCurrentThread()
	if e != nil {
		return 0, e
	}
	var tok windows.Token
	e = OpenThreadToken(t, windows.TOKEN_QUERY, true, &tok)
	if e != nil {
		return 0, e
	}
	return tok, nil
}

func HasAdminRights() (bool, error) {
	// https://github.com/golang/go/issues/28804
	var sid *windows.SID
	err := windows.AllocateAndInitializeSid(&windows.SECURITY_NT_AUTHORITY, 2, windows.SECURITY_BUILTIN_DOMAIN_RID, windows.DOMAIN_ALIAS_RID_ADMINS, 0, 0, 0, 0, 0, 0, &sid)
	if err != nil {
		return false, err
	}
	defer windows.FreeSid(sid)

	token, err := OpenCurrentThreadToken()
	if err != nil {
		return false, err
	}

	member, err := token.IsMember(sid)
	if err != nil {
		return false, err
	}
	return member, nil
}

func InternetOptionSettingsChanged() (syscall.Handle, error) {
	// https://github.com/Diving-Fish/maimaidx-prober/blob/1673cff975e6bbb14eeeda6260af131ac036f974/proxy/lib/wininet_windows.go#L18
	// https://docs.microsoft.com/en-us/windows/win32/wininet/option-flags
	p1 := uint16(0)
	p2 := uint64(INTERNET_OPTION_SETTINGS_CHANGED)
	p3 := uint16(0)
	p4 := uint64(0)
	r1, _, e1 := syscall.Syscall6(
		procInternetSetOptionW.Addr(),
		4,
		uintptr(p1),
		uintptr(p2),
		uintptr(p3),
		uintptr(p4),
		0,
		0,
	)
	if r1 == 0 {
		if e1 != 0 {
			return 0, e1
		} else {
			return 0, syscall.EINVAL
		}
	}
	return syscall.Handle(r1), nil
}

const profileListPath = "SOFTWARE\\Microsoft\\Windows NT\\CurrentVersion\\ProfileList"

func getProfileListSubKeyNames() ([]string, error) {
	// https://github.com/ninoseki/ninoseki.github.io/blob/0e911fe146c0de7b2fa4449ff9c2ac1ebe755b67/docs/_posts/2020-07-25-windows-registry-with-go.md
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, profileListPath, registry.ENUMERATE_SUB_KEYS)
	if err != nil {
		return nil, fmt.Errorf("unable to open registry key %q: %v", profileListPath, err)
	}
	defer k.Close()
	return k.ReadSubKeyNames(0)
}

type systemProxy struct{}

var SystemProxy systemProxy

func (p *systemProxy) AddIPWhitelist(cidr string) {}

func (p *systemProxy) RemoveIPWhitelist(cidr string) {}

type todo struct {
	Key    registry.Key
	Prefix string
}

func (p *systemProxy) GetSetupCommands() Setter {
	hasAdminRights, err := HasAdminRights()
	if err != nil {
		return NewErrorSetter(err)
	}
	setter := Setter{
		PreFunc: func() error {
			var todolist []todo
			if hasAdminRights {
				// https://docs.microsoft.com/en-us/windows/win32/services/services-and-the-registry
				// A service should not access HKEY_CURRENT_USER or HKEY_CLASSES_ROOT, especially when impersonating a user.
				sids, err := getProfileListSubKeyNames()
				if err != nil {
					log.Debug("GetSetupCommands: getProfileListSubKeyNames: %v", err)
					return err
				}
				for _, sid := range sids {
					todolist = append(todolist, todo{
						Key:    registry.USERS,
						Prefix: sid + `\`,
					})
				}
			} else {
				todolist = append(todolist, todo{
					Key:    registry.CURRENT_USER,
					Prefix: "",
				})
			}
			for _, todo := range todolist {
				key, err := registry.OpenKey(todo.Key, todo.Prefix+`SOFTWARE\Microsoft\Windows\CurrentVersion\Internet Settings`, registry.ALL_ACCESS)
				if err != nil {
					log.Debug("GetSetupCommands: OpenKey %v, %v: %v", todo.Key, todo.Prefix+`SOFTWARE\Microsoft\Windows\CurrentVersion\Internet Settings`, err)
					continue
				}
				defer key.Close()
				_ = key.DeleteValue("AutoConfigURL")
				if err = key.SetDWordValue("ProxyEnable", 1); err != nil {
					log.Debug("GetSetupCommands: key: %v: SetDWordValue ProxyEnable: %v", todo.Key, err)
					return err
				}
				if err = key.SetStringValue("ProxyServer", "127.0.0.1:32345"); err != nil {
					log.Debug("GetSetupCommands: key: %v: SetStringValue ProxyServer: %v", todo.Key, err)
					return err
				}
			}
			_, _ = InternetOptionSettingsChanged()
			if hasAdminRights {
				// https://helpcenter.gsx.com/hc/en-us/articles/216487418-How-to-Import-Internet-Explorer-Proxy-Configuration-for-PowerShell-Use
				// You can browse the Internet and open OWA successfully using Internet Explorer (IE) but you cannot connect to Office 365 using PowerShell.
				// To fix this, we set Windows Proxy settings using NETSH for all applications that rely on default system configuration.
				time.Sleep(time.Second)
				return exec.Command("netsh", "winhttp", "import", "proxy", "source=ie").Run()
			} else {
				return nil
			}
		},
	}
	return setter
}

func (p *systemProxy) GetCleanCommands() Setter {
	hasAdminRights, err := HasAdminRights()
	if err != nil {
		return NewErrorSetter(err)
	}
	setter := Setter{
		PreFunc: func() error {
			var todolist []todo
			if hasAdminRights {
				// https://docs.microsoft.com/en-us/windows/win32/services/services-and-the-registry
				// A service should not access HKEY_CURRENT_USER or HKEY_CLASSES_ROOT, especially when impersonating a user.
				sids, err := getProfileListSubKeyNames()
				if err != nil {
					return err
				}
				for _, sid := range sids {
					todolist = append(todolist, todo{
						Key:    registry.USERS,
						Prefix: sid + `\`,
					})
				}
			} else {
				todolist = append(todolist, todo{
					Key:    registry.CURRENT_USER,
					Prefix: "",
				})
			}

			var errs []error
			for _, todo := range todolist {
				key, _, err := registry.CreateKey(todo.Key, todo.Prefix+`SOFTWARE\Microsoft\Windows\CurrentVersion\Internet Settings`, registry.ALL_ACCESS)
				if err != nil {
					errs = append(errs, err)
					continue
				}
				defer key.Close()
				if err = key.SetDWordValue("ProxyEnable", 0); err != nil {
					errs = append(errs, err)
					continue
				}
			}
			_, _ = InternetOptionSettingsChanged()
			if hasAdminRights {
				// https://helpcenter.gsx.com/hc/en-us/articles/216487418-How-to-Import-Internet-Explorer-Proxy-Configuration-for-PowerShell-Use
				// You can browse the Internet and open OWA successfully using Internet Explorer (IE) but you cannot connect to Office 365 using PowerShell.
				// To fix this, we set Windows Proxy settings using NETSH for all applications that rely on default system configuration.
				e := exec.Command("netsh", "winhttp", "import", "proxy", "source=ie").Run()
				if e != nil {
					errs = append(errs, e)
				}
			}
			if len(errs) > 0 {
				return errs[0]
			}
			return nil
		},
	}
	return setter
}
