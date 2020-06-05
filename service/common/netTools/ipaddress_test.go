package netTools

import (
	"testing"
)

func TestIsIntranet4(t *testing.T) {
	tests := [][2]interface{}{
		{[4]byte{192, 168, 50, 1}, true},
		{[4]byte{172, 20, 10, 1}, true},
		{[4]byte{172, 16, 10, 1}, true},
		{[4]byte{127, 16, 10, 1}, true},
		{[4]byte{222, 16, 10, 1}, false},
		{[4]byte{10, 0, 0, 145}, true},
		{[4]byte{1, 2, 3, 4}, false},
	}
	for _, tt := range tests {
		ipv4 := tt[0].([4]byte)
		if ans := IsIntranet4(&ipv4); ans == tt[1] {
			t.Log(tt[0], "result", tt[1])
		} else {
			t.Fatal(tt[0], "expect", tt[1], ", wrong answer is:", ans)
		}
	}
}
