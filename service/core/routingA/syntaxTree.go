package routingA

import (
	"v2rayA/common"
	"v2rayA/global"
	"v2ray.com/core/common/errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

func toSym(r rune) rune {
	o := map[rune]interface{}{
		',':  nil,
		'\'': nil,
		'"':  nil,
		'(':  nil,
		')':  nil,
	}
	reserve := map[rune]interface{}{
		',':  nil,
		'\'': nil,
		'"':  nil,
		'(':  nil,
		')':  nil,
		':':  nil,
		'&':  nil,
		'-':  nil,
		'>':  nil,
		'=':  nil,
	}
	switch {
	case unicode.IsPunct(r):
		if _, ok := o[r]; ok {
			return r
		} else if _, ok := reserve[r]; ok {
			return r
		} else {
			return 'n'
		}
	case unicode.IsDigit(r):
		return 'l'
	case r == '\n':
		return 'r'
	case r == 0:
		return '#'
	default:
		if _, ok := reserve[r]; ok {
			return r
		}
		return 'k'
	}
}

func matchRegex(syms []rune, regex string) bool {
	if len(regex) > len(syms) {
		return false
	}
	syms = syms[len(syms)-len(regex):]
	re := ""
	//FIXME: 由于n在后期处理时有时包含:&->=，有时不包含，导致这里必须特判，所以本函数有可能漏判，即match率偏高
	equal := func(str, regex string) bool {
		for i := range str {
			if regex[i] == 'n' {
				switch str[i] {
				case ':', '&', '-', '>', '=':
				default:
					if str[i] != regex[i] {
						return false
					}
				}
			} else if regex[i] == 'k' {
				switch str[i] {
				case 'k', 'l', 'n':
				default:
					if str[i] != regex[i] {
						return false
					}
				}
			} else if str[i] != regex[i] {
				return false
			}
		}
		return true
	}
	return equal(re, regex)
}

func preprocess(program string) string {
	lines := strings.Split(program, "\n")
	for i := range lines {
		lines[i] = strings.TrimSpace(lines[i])
		if len(lines[i]) > 0 && lines[i][0] == '#' {
			lines[i] = ""
		}
	}
	return strings.Join(lines, "\n")
}

func passSpace(r rune, sTop int) bool {
	if unicode.IsSpace(r) && r != '\n' {
		//contextRadius := 5
		//context := string(str[tools.Max(i-contextRadius, 0):i]) + "<" + string(str[i]) + ">" + string(str[i+1:tools.Min(i+contextRadius+1, len(str))])
		switch sTop {
		case 0, 12, 15, 22, 27, 55, 37, 18, 25, 20, 50, 44, 52:
			//log.Println("skip", sTop, "<"+string(str[i])+">", "context:", context)
			//pass space
			return true
		default:
			//log.Println(sTop, "<"+string(str[i])+">", "context:", context)
		}
	}
	return false
}
func postHandleMsg(msg string, str []rune, i int) string {
	switch {
	case strings.HasPrefix(msg, "[special1]"):
	loop:
		for j := i + 1; j < len(str); j++ {
			switch (str)[j] {
			case ':':
				msg = "r11"
				break loop
			case ',', ')':
				msg = "s44"
				break loop
			}
		}
	}
	return msg
}

//LR(1)文法
func generateSyntaxTree(program string) (S symbol, err error) {
	var logsBuf strings.Builder
	defer func() {
		if err != nil && global.IsDebug() {
			err = errors.New(logsBuf.String() + err.Error())
		}
	}()
	var stackR, stackS = make(symbols, 1), make([]int, 1)
	stackR[0] = symbol{sym: 0}
	stackS[0] = 0
	program = preprocess(program)
	program += string(stackR[0].sym)
	str := []rune(program)
out:
	for i := 0; i < len(str); {
		sTop := stackS[len(stackS)-1]
		//略过空格
		if passSpace(str[i], sTop) {
			i++
			continue
		}
		sym := toSym(str[i])
		val := string(str[i])
		msg := table[sTop][string(sym)]
		//特殊处理
		msg = postHandleMsg(msg, str, i)
		logsBuf.WriteString(fmt.Sprintln("[status]", "StackRune:", stackR.String(), "StackState:", stackS, "Str[i]:", val, "Sym:", string(sym), "Msg:", msg))
		switch {
		case strings.HasPrefix(msg, "s"):
			state, _ := strconv.Atoi(msg[1:])
			stackS = append(stackS, state)
			stackR = append(stackR, symbol{sym: sym, val: val})
			logsBuf.WriteString(fmt.Sprintln("[info]", "Shift:", val, "Push state:", state))
			i++
		case strings.HasPrefix(msg, "r"):
			state, _ := strconv.Atoi(msg[1:])
			production := productions[state]
			logsBuf.WriteString(fmt.Sprintln("[info]", string(production.left), "<-", production.right))
			stackS = stackS[:len(stackS)-len(production.right)]
			gt := table[stackS[len(stackS)-1]][string(production.left)]
			var newState int
			newState, err = strconv.Atoi(gt)
			if err != nil {
				err = errors.New("[error] " + fmt.Sprintf("table[%v][%v]", stackS[len(stackS)-1], string(production.left)) + " <" + err.Error() + ">")
				return
			}
			stackS = append(stackS, newState)
			if !matchRegex(stackR.Runes(), production.right) {
				logsBuf.WriteString(fmt.Sprintln("[warning]", "TrimSuffix:", stackR.String(), "not match regex:", production.right))
			}
			//形成语法树
			reducedSyms := stackR[len(stackR)-len(production.right):]
			stackR = stackR[:len(stackR)-len(production.right)]
			children := make([]symbol, len(reducedSyms))
			copy(children, reducedSyms)
			val := ""
			for _, s := range reducedSyms {
				val += s.val
			}
			stackR = append(stackR, symbol{sym: production.left, children: children, val: val})
		case msg == "acc":
			S = stackR[len(stackR)-1]
			break out
		case msg == "":
			end := strings.Index(string(str[i:]), "\n")
			if end == -1 {
				end = 0x3f3f3f3f
			}
			err = errors.New(fmt.Sprintf("[error] position: %v unexpected character %c <...%v>", i, str[i], string(str[i:common.Min(len(str)-1, i+end)])))
			return
		default:
			err = errors.New("[error] " + msg)
			return
		}
	}
	return
}
