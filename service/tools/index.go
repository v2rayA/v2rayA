package tools

import "github.com/matoous/go-nanoid"

var secret string

func init() {
	//屡次启动的Secret都不一样
	id, err := gonanoid.Nanoid()
	if err != nil {
		panic(err)
	}
	secret = id
}

func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func Min(a, b int) int {
	if a > b {
		return b
	}
	return a
}