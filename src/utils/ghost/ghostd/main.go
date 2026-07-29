// Это не малварь и не бэкдор - образовательный артефакт для квеста.
// Два экземпляра этой же программы запускаются под разными именами
// (см. deploy_ghost_quest.sh) и следят друг за другом. Если один из них убить,
// второй его перезапускает через 10-20 секунд. Чтобы остановить обоих
// насовсем, нужно понять, что их двое, и убить оба почти одновременно,
// либо убрать respawn-механизм осознанно (например, потушив оба разом
// через pkill по обоим именам в одну команду).
// Имя в ps/proc/[pid]/comm подделывается под kworker/R:7, чтобы
// сливаться с обычными списком keventd/kworker-потоков ядра. Ключевая
// улика для внимательного ученика: у настоящих kworker'ов нет
// исполняемого файла на диске (это часть ядра), а у этого -
// /proc/<pid>/exe резолвится в реальный путь.
package main

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	baseDir       = "/opt/.legacy-diag"
	logFile       = baseDir + "/.still_here.log"
	selfPidFile   = baseDir + "/.ghostd.pid"
	peerPidFile   = baseDir + "/.echod.pid"
	selfBinary    = baseDir + "/.ghostd"
	peerBinary    = baseDir + "/.echod"
	selfDisguise  = "kworker/R:7"
	peerDisguise  = "kworker/R:8"
	watchInterval = 17 * time.Second
	writeInterval = 15 * time.Second
)

const payloadPlain = "menya optimizirovali. etot process net. esli ty eto chitaesh, znachit kto-to vsyo zhe smotrel vnimatelno. R."

func main() {
	role := "ghostd"
	if len(os.Args) > 1 && os.Args[1] == "--peer" {
		role = "echod"
	}
	disguiseSelf(role)

	if role == "ghostd" {
		runAs(selfPidFile, peerPidFile, peerBinary)
	} else {
		runAs(peerPidFile, selfPidFile, selfBinary)
	}
}

func disguiseSelf(role string) {
	name := selfDisguise
	if role == "echod" {
		name = peerDisguise
	}
	setProcTitle(name)
}

func runAs(myPidFile, peerPidFilePath, peerBin string) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		os.Exit(1)
	}
	writePid(myPidFile)

	writeTicker := time.NewTicker(writeInterval)
	watchTicker := time.NewTicker(watchInterval)
	defer writeTicker.Stop()
	defer watchTicker.Stop()

	iteration := 0
	for {
		select {
		case <-writeTicker.C:
			appendLog(iteration)
			iteration++
		case <-watchTicker.C:
			ensurePeerAlive(peerPidFilePath, peerBin)
		}
	}
}

func writePid(path string) {
	pid := os.Getpid()
	_ = os.WriteFile(path, []byte(strconv.Itoa(pid)), 0644)
}

func ensurePeerAlive(peerPidFilePath, peerBin string) {
	data, err := os.ReadFile(peerPidFilePath)
	if err != nil {
		respawnPeer(peerBin)
		return
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		respawnPeer(peerBin)
		return
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		respawnPeer(peerBin)
		return
	}
	if err := proc.Signal(syscall.Signal(0)); err != nil {
		respawnPeer(peerBin)
	}
}

func respawnPeer(peerBin string) {
	cmd := exec.Command(peerBin, "--peer")
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	_ = cmd.Start()
}

func appendLog(iteration int) {
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	ts := time.Now().Format("2006-01-02 15:04:05")

	if iteration%5 == 0 {
		fmt.Fprintf(f, "[%s] disk usage: %d%% - within limits\n", ts, 55+rand.Intn(10))
		return
	}

	encoded := encodePayload(payloadPlain)
	fmt.Fprintf(f, "[%s] heartbeat: %s\n", ts, encoded)
}

func encodePayload(plain string) string {
	rot := rot13(plain)
	return base64.StdEncoding.EncodeToString([]byte(rot))
}

func rot13(s string) string {
	rotate := func(r rune, base rune) rune {
		return (r-base+13)%26 + base
	}
	out := []rune(s)
	for i, r := range out {
		switch {
		case r >= 'a' && r <= 'z':
			out[i] = rotate(r, 'a')
		case r >= 'A' && r <= 'Z':
			out[i] = rotate(r, 'A')
		}
	}
	return string(out)
}
