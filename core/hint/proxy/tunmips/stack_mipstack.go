// SPDX-License-Identifier: MPL-2.0
package tunmips

import (
	"context"
	"errors"
	"net/netip"
	"os"
	"sync"

	"github.com/metacubex/mipstack"
	"golang.zx2c4.com/wireguard/tun"
)

// mipsStack drives a mipstack.Stack with two pumps: device reads are written
// into the stack, stack output is written to the device.
type mipsStack struct {
	mtu   uint32
	stack *mipstack.Stack
	tcp   *mipstack.TCPForwarder
	udp   *mipstack.UDPForwarder

	dev       tun.Device
	closeOnce sync.Once
	done      chan struct{}
	wg        sync.WaitGroup
	// pumpErr records the first fatal pump error for Close callers.
	pumpErr error
	errMu   sync.Mutex
}

func newMipsStack(mtu uint32) *mipsStack {
	return &mipsStack{mtu: mtu, done: make(chan struct{})}
}

// Start builds the stack in transparent mode: it owns no address, admits
// every unicast packet the device delivers, and answers from whatever
// address the application dialed. The interface address belongs to the
// operating system side of the link and must not be listed here, otherwise
// replies to it would be routed to the stack's internal loopback instead of
// the device.
func (s *mipsStack) Start(dev tun.Device, h Handlers) error {
	st, err := mipstack.New(mipstack.Config{
		Promiscuous: true,
		MTU:         s.mtu,
	})
	if err != nil {
		return err
	}
	tcp, err := mipstack.NewTCPForwarder(st, mipstack.TCPForwarderOptions{}, func(r *mipstack.TCPForwarderRequest) {
		flow := Flow{Source: r.Flow().Source, Destination: r.Flow().Destination}
		conn, err := r.Accept(context.Background())
		if err != nil {
			return
		}
		if h.TCP == nil {
			conn.Close()
			return
		}
		h.TCP(flow, conn)
	})
	if err != nil {
		st.Close()
		return err
	}
	udp, err := mipstack.NewUDPForwarder(st, mipstack.UDPForwarderOptions{}, func(r *mipstack.UDPForwarderRequest) {
		if h.UDP == nil {
			r.Drop()
			return
		}
		flow := Flow{Source: r.Flow().Source, Destination: r.Flow().Destination}
		responder, err := r.Detach()
		if err != nil {
			return
		}
		h.UDP(flow, responder.Payload(), func(payload []byte, from netip.AddrPort) error {
			_, err := responder.ReplyFrom(payload, from)
			return err
		})
	})
	if err != nil {
		tcp.Close()
		st.Close()
		return err
	}
	if err := st.Start(); err != nil {
		udp.Close()
		tcp.Close()
		st.Close()
		return err
	}
	s.stack, s.tcp, s.udp, s.dev = st, tcp, udp, dev

	s.wg.Add(2)
	go s.pumpDeviceToStack()
	go s.pumpStackToDevice()
	return nil
}

func (s *mipsStack) pumpDeviceToStack() {
	defer s.wg.Done()
	batch := s.dev.BatchSize()
	bufs := make([][]byte, batch)
	for i := range bufs {
		bufs[i] = make([]byte, packetOffset+maxPacketSize)
	}
	sizes := make([]int, batch)
	packets := make([][]byte, 0, batch)
	for {
		n, err := s.dev.Read(bufs, sizes, packetOffset)
		if err != nil {
			s.fail(err)
			return
		}
		packets = packets[:0]
		for i := 0; i < n; i++ {
			if sizes[i] == 0 {
				continue
			}
			packets = append(packets, bufs[i][:packetOffset+sizes[i]])
		}
		if len(packets) == 0 {
			continue
		}
		if _, err := s.stack.Write(packets, packetOffset); err != nil {
			s.fail(err)
			return
		}
	}
}

func (s *mipsStack) pumpStackToDevice() {
	defer s.wg.Done()
	batch := s.dev.BatchSize()
	bufs := make([][]byte, batch)
	for i := range bufs {
		bufs[i] = make([]byte, packetOffset+maxPacketSize)
	}
	sizes := make([]int, batch)
	packets := make([][]byte, 0, batch)
	for {
		n, err := s.stack.Read(bufs, sizes, packetOffset)
		if err != nil {
			s.fail(err)
			return
		}
		packets = packets[:0]
		for i := 0; i < n; i++ {
			if sizes[i] == 0 {
				continue
			}
			packets = append(packets, bufs[i][:packetOffset+sizes[i]])
		}
		if len(packets) == 0 {
			continue
		}
		if _, err := s.dev.Write(packets, packetOffset); err != nil {
			s.fail(err)
			return
		}
	}
}

// fail records the first pump error and tears the stack down so the other
// pump unblocks. A closed device or stack is the normal shutdown path and is
// not recorded.
func (s *mipsStack) fail(err error) {
	if errors.Is(err, os.ErrClosed) || errors.Is(err, mipstack.ErrClosed) {
		s.Close()
		return
	}
	s.errMu.Lock()
	if s.pumpErr == nil {
		s.pumpErr = err
	}
	s.errMu.Unlock()
	s.Close()
}

// Err returns the first fatal pump error, or nil after a clean shutdown.
func (s *mipsStack) Err() error {
	s.errMu.Lock()
	defer s.errMu.Unlock()
	return s.pumpErr
}

// Close stops both forwarders and the stack. The device is owned by the
// caller; closing it is what unblocks the device pump.
func (s *mipsStack) Close() error {
	s.closeOnce.Do(func() {
		close(s.done)
		if s.udp != nil {
			s.udp.Close()
		}
		if s.tcp != nil {
			s.tcp.Close()
		}
		if s.stack != nil {
			s.stack.Close()
		}
	})
	return nil
}
