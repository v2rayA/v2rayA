package infra

// 198.18.0.0/15
// FIXME: exceed
type reservedIP uint32

func (r reservedIP) IP() (ip [4]byte) {
	k := uint32(r)
	for i := 3; i >= 0; i-- {
		ip[i] = byte(k & 0xff)
		k >>= 8
	}
	ip[0] += 198
	ip[1] += 18
	return
}

func (r reservedIP) Next() reservedIP {
	if r&0xff == 254 {
		return r + 3
	}
	return r + 1
}
