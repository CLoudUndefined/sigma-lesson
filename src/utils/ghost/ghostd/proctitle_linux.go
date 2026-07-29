//go:build linux

package main

import (
	"syscall"
	"unsafe"
)

const prSetName = 15 // PR_SET_NAME из linux/prctl.h

// setProcTitle подделывает имя процесса, видимое в `ps`/`top`/`/proc/[pid]/comm`.
// Ограничение ядра - максимум 15 байт + null terminator, поэтому имена вида
// "kworker/R:7" (11 байт) укладываются с запасом.
func setProcTitle(name string) {
	b := append([]byte(name), 0)
	if len(b) > 16 {
		b = b[:16]
		b[15] = 0
	}
	syscall.Syscall(syscall.SYS_PRCTL, uintptr(prSetName), uintptr(unsafe.Pointer(&b[0])), 0)
}
