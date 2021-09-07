package iptables

import (
	"fmt"
	"github.com/v2rayA/go-uci"
	"github.com/v2rayA/v2rayA/common"
	"os"
	"strconv"
)

type Mark struct {
	Mask uint32
}

type MarkGetter func() (*Mark, error)

func mwan3Mark() (*Mark, error) {
	if common.IsOpenWrt() {
		const defaultMask = "0x3F00"
		b, err := os.ReadFile("/etc/config/mwan3")
		if err != nil {
			return nil, err
		}
		cfg, err := uci.Parse("mwan3", string(b))
		if err != nil {
			return nil, err
		}
		var mask string
		if globals := cfg.Get("globals"); globals == nil {
			mask = defaultMask
		} else {
			mMask := globals.Get("mmx_mask")
			if len(mMask.Values) == 0 {
				mask = defaultMask
			} else {
				mask = mMask.Values[0]
			}
		}
		if mask, err := strconv.ParseUint(mask, 16, 32); err != nil {
			return nil, fmt.Errorf("invalid mwan3 mask")
		} else {
			return &Mark{Mask: uint32(mask)}, nil
		}
	}
	return nil, fmt.Errorf("current OS is not OpenWrt")
}

func KnownMarks() []Mark {
	getters := []MarkGetter{
		mwan3Mark,
	}
	var marks []Mark
	for _, getter := range getters {
		if mark, _ := getter(); mark != nil {
			marks = append(marks, *mark)
		}
	}
	return marks
}
