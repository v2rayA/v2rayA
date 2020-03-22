package netTools

import (
	Trie "V2RayA/dataStructure/trie"
	"strconv"
	"strings"
)

var intranet4 = []string{
	"0.0.0.0/32",
	"10.0.0.0/8",
	"127.0.0.0/8",
	"169.254.0.0/16",
	"172.16.0.0/12",
	"192.168.0.0/16",
	"224.0.0.0/4",
	"240.0.0.0/4",
	"255.255.255.255/32",
}

var trie *Trie.Trie

func init() {
	dict := make([]string, 0, len(intranet4))
	for _, intra := range intranet4 {
		grp := strings.SplitN(intra, "/", 2)
		l, _ := strconv.Atoi(grp[1])
		arr := strings.Split(grp[0], ".")
		buf := ""
		for _, sec := range arr {
			itg, _ := strconv.Atoi(sec)
			tmp := strconv.FormatInt(int64(itg), 2)
			buf += strings.Repeat("0", 8-len(tmp)) + tmp
			if len(buf) >= l {
				break
			}
		}
		dict = append(dict, buf[:l])
	}
	trie = Trie.New(dict)
}

func IsIntranet4(ipv4 [4]byte) bool {
	var buf string
	for _, b := range ipv4 {
		tmp := strconv.FormatInt(int64(b), 2)
		buf += strings.Repeat("0", 8-len(tmp)) + tmp
	}
	return trie.Match(buf) != ""
}
