package dnsPoison

type DnsPoison2 struct {
}

//
//func SupportV2() error {
//	_, err := exec.Command("sh", "-c", "modprobe xt_NFQUEUE").Output()
//	if err != nil {
//		return errors.New("Not support NFQUEUE")
//	}
//	return nil
//}
//
//func NewV2() (*DnsPoison, error) {
//	if err := SupportV2(); err != nil {
//		return nil, err
//	}
//	return &DnsPoison{}, nil
//}
//
//func callback(payload *nfqueue.Payload) int {
//	fmt.Printf("  id: %d\n", payload.Id)
//	fmt.Println(hex.Dump(payload.Data))
//	// Decode a packet
//	packet := gopacket.NewPacket(payload.Data, layers.LayerTypeIPv4, gopacket.NoCopy)
//	if etherLayer := packet.Layer(layers.LayerTypeEthernet); etherLayer != nil {
//		fmt.Println("This is a TCP packet!")
//		trans := packet.TransportLayer()
//		if trans == nil {
//			payload.SetVerdict(nfqueue.NF_ACCEPT)
//			return 0
//		}
//		transflow := trans.TransportFlow()
//		sPort, dPort := transflow.Endpoints()
//		if sPort.String() != "53" {
//			payload.SetVerdict(nfqueue.NF_ACCEPT)
//			return 0
//		}
//		sAddr, dAddr := packet.NetworkLayer().NetworkFlow().Endpoints()
//		// TODO: 暂不支持IPv6
//		sIP := net.ParseIP(dAddr.String()).To4()
//		if len(sIP) != net.IPv4len {
//			payload.SetVerdict(nfqueue.NF_ACCEPT)
//			return 0
//		}
//
//		var m dnsmessage.Message
//		err := m.Unpack(trans.LayerPayload())
//		if err != nil {
//			payload.SetVerdict(nfqueue.NF_ACCEPT)
//			return 0
//		}
//		// dns请求一般只有一个question
//		q := m.Questions[0]
//		if (q.Type != dnsmessage.TypeA && q.Type != dnsmessage.TypeAAAA) ||
//			q.Class != dnsmessage.ClassINET {
//			payload.SetVerdict(nfqueue.NF_ACCEPT)
//			return 0
//		}
//		switch a := m.Answers[0].Body.(type) {
//		case *dnsmessage.AResource:
//			if bytes.Equal(a.A[:], []byte{127, 0, 0, 1}) {
//				payload.SetVerdict(nfqueue.NF_DROP)
//				log.Println("dnsPoisonV2投毒:", sAddr.String()+":"+sPort.String(), "->", dAddr.String()+":"+dPort.String(), m.Questions)
//				go poison(&m, &sAddr, &sPort, &dAddr, &dPort)
//				return 0
//			}
//		}
//	}
//	payload.SetVerdict(nfqueue.NF_ACCEPT)
//	return 0
//}
//
//func (d *DnsPoison2) Run() {
//	q := new(nfqueue.Queue)
//
//	q.SetCallback(callback)
//
//	q.Init()
//
//	q.Unbind(syscall.AF_INET)
//	q.Bind(syscall.AF_INET)
//
//	q.CreateQueue(0)
//
//	c := make(chan os.Signal, 1)
//	signal.Notify(c, os.Interrupt)
//	go func() {
//		for sig := range c {
//			// sig is a ^C, handle it
//			_ = sig
//			q.StopLoop()
//		}
//	}()
//
//	// XXX Drop privileges here
//
//	q.Loop()
//	q.DestroyQueue()
//	q.Close()
//}
