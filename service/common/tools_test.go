package common

import (
	"testing"
)

func TestUrlEncoded(t *testing.T) {
	str := `试试1+就试试!`
	t.Log(UrlEncoded(str))
}
