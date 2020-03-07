package routingA

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"unicode"
)

func toHeader(r rune) string {
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
			return string(r)
		} else if _, ok := reserve[r]; ok {
			return string(r)
		} else {
			return "n"
		}
	case unicode.IsDigit(r):
		return "l"
	case r == '\n':
		return `\n`
	case r == 0:
		return "#"
	default:
		if _, ok := reserve[r]; ok {
			return string(r)
		}
		return "k"
	}
}

func matchRegex(stackR []rune, regex string) bool {
	if len(regex) > len(stackR) {
		return false
	}
	stackR = stackR[len(stackR)-len(regex):]
	re := ""
	for i := range stackR {
		if _, ok := table[0][string(stackR[i])]; ok {
			re += string(stackR[i])
		} else {
			re += toHeader(stackR[i])
		}
	}
	//FIXME: 有可能漏判，即match率偏高
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

//LR(1)文法解析器，已完成语法分析
//TODO: 语义分析
func Parse(program string) {
	var stackR, stackS = make([]rune, 1), make([]int, 1)
	stackR[0] = 0
	stackS[0] = 0
	program = preprocess(program)
	program += string(stackR[0])
	str := []rune(program)
out:
	for i := 0; i < len(str); {
		msg := table[stackS[len(stackS)-1]][toHeader(str[i])]
		log.Println("[status]", "StackRune:", string(stackR), "StackState:", stackS, "Str[i]:", string(str[i]), "Header:", toHeader(str[i]), "Msg:", msg)
		switch {
		case strings.HasPrefix(msg, "[special1]"):
		loop:
			for j := i + 1; j < len(str); j++ {
				switch str[j] {
				case ':':
					msg = "r11"
					break loop
				case ',', ')':
					msg = "s44"
					break loop
				}
			}
		}
		switch {
		case strings.HasPrefix(msg, "s"):
			state, _ := strconv.Atoi(msg[1:])
			stackS = append(stackS, state)
			stackR = append(stackR, str[i])
			log.Println("[info]", "Shift:", string(str[i]), "Push state:", state)
			i++
		case strings.HasPrefix(msg, "r"):
			state, _ := strconv.Atoi(msg[1:])
			pro := productions[state]
			log.Println("[info]", string(pro.left), "<-", pro.right)
			stackS = stackS[:len(stackS)-len(pro.right)]
			gt := table[stackS[len(stackS)-1]][string(pro.left)]
			newState, err := strconv.Atoi(gt)
			if err != nil {
				log.Fatal("[error] " + fmt.Sprintf("table[%v][%v]", stackS[len(stackS)-1], string(pro.left)) + " <" + err.Error() + ">")
			}
			stackS = append(stackS, newState)
			if !matchRegex(stackR, pro.right) {
				log.Println("[warning]", "TrimSuffix:", string(stackR), "not match regex:", pro.right)
			}
			stackR = stackR[:len(stackR)-len(pro.right)]
			stackR = append(stackR, pro.left)
		case msg == "acc":
			break out
		case msg == "":
			log.Fatal(fmt.Sprintf("[error] position: %v unexpected character %c <...%v>", i, str[i], string(str[i:len(str)-1])))
		default:
			log.Fatal("[error] " + msg)
		}
	}
}
