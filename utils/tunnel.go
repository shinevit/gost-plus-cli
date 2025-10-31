package utils

import (
	"fmt"
	"strings"

	"github.com/go-gost/gost.plus/tunnel"
	"github.com/go-gost/gost.plus/utils/fp"
	opt "github.com/go-gost/gost.plus/utils/fp/option"
	"github.com/go-gost/x/service"
)

func GetDisplayState(instance tunnel.Tunnel) string {
	// State => string
	upperCaseFn := func(st service.State) string {
		return strings.ToUpper(string(st))
	}
	return fp.Map(GetState(instance), upperCaseFn)
}

func GetState(tun tunnel.Tunnel) service.State {
	defaultState := service.StateFailed
	if tun.IsClosed() {
		defaultState = service.StateClosed
	}

	getStateFlow := opt.Flow1(
		opt.MapResult(func(st *service.Status) service.State {
			return st.State()
		}),
		opt.GetOrElse(defaultState),
	)

	err, state := getStateFlow.Run(tun.Status())
	if err != nil {
		fmt.Println(err)
	}
	return state
}
