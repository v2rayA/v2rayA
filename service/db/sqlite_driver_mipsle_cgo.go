//go:build (mips || mipsle || mips64 || mips64le || mips64p32 || mips64p32le || loong64) && cgo

package db

import _ "github.com/mattn/go-sqlite3"

const sqliteDriverName = "sqlite3"
