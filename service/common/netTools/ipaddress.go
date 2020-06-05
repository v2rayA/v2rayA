package netTools

import (
	"bytes"
	"strconv"
	"strings"
	Trie "v2rayA/dataStructure/trie"
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
}

var jokernet4 = []string{
	"0.0.0.0/8",
	"127.0.0.0/8",
	"240.0.0.0/4",
}

var trieIntranet *Trie.Trie
var trieJokernet *Trie.Trie

func init() {
	trieIntranet = Init(intranet4)
	trieJokernet = Init(jokernet4)
}

func Init(CIDRs []string) *Trie.Trie {
	dict := make([]string, 0, len(CIDRs))
	for _, CIDR := range CIDRs {
		grp := strings.SplitN(CIDR, "/", 2)
		l, _ := strconv.Atoi(grp[1])
		arr := strings.Split(grp[0], ".")
		var builder strings.Builder
		for _, sec := range arr {
			itg, _ := strconv.Atoi(sec)
			tmp := strconv.FormatInt(int64(itg), 2)
			builder.WriteString(strings.Repeat("0", 8-len(tmp)))
			builder.WriteString(tmp)
			if builder.Len() >= l {
				break
			}
		}
		dict = append(dict, builder.String()[:l])
	}
	return Trie.New(dict)
}

func ipv4ToBin(ipv4 *[4]byte) string {
	var buff = new(bytes.Buffer)
	for _, b := range ipv4 {
		tmp := strconv.FormatInt(int64(b), 2)
		buff.WriteString(strings.Repeat("0", 8-len(tmp)) + tmp)
	}
	return buff.String()
}

func IsIntranet4(ipv4 *[4]byte) bool {
	return trieIntranet.Match(ipv4ToBin(ipv4)) != ""
}

func IsJokernet4(ipv4 *[4]byte) bool {
	return trieJokernet.Match(ipv4ToBin(ipv4)) != ""
}
