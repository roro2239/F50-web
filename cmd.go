package main

import (
	"bufio"
	"errors"
	"io"
	"os"
	"os/exec"
	"sync"
)

var ttydCmd = NewCommandExecutor(`#!/bin/sh
SOCK_FILE="${F50_WEB_RUN_DIR}/ttyd-termux.sock"
TTYD_BIN="${F50_WEB_BIN_DIR}/ttyd"
if [ ! -x "$TTYD_BIN" ]; then
    TTYD_BIN="$(command -v ttyd 2>/dev/null || true)"
fi
if [ -z "$TTYD_BIN" ] || [ ! -x "$TTYD_BIN" ]; then
    echo "ttyd binary not found"
    exit 1
fi
SHELL_BIN="${F50_WEB_SHELL}"
if ! pgrep -f "$TTYD_BIN -q -W -i $SOCK_FILE $SHELL_BIN" > /dev/null; then
    "$TTYD_BIN" -q -W -i "$SOCK_FILE" "$SHELL_BIN" &
fi
`)

func init() {

}

type CommandExecutor struct {
	sh          string
	cmd         *exec.Cmd
	logBuffer   []string
	maxLogLines int
	status      string
	mutex       sync.Mutex
	wg          sync.WaitGroup // 用于等待日志捕获完成
}

func NewCommandExecutor(sh string) *CommandExecutor {
	return &CommandExecutor{
		sh:          sh,
		maxLogLines: 100,
		status:      "stopped",
	}
}

func (e *CommandExecutor) Start() error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if e.status == "running" {
		return errors.New("command is already running")
	}

	e.cmd = exec.Command(shellPath(), "-c", e.sh)
	e.cmd.Env = append(os.Environ(), runtimeEnv()...)

	stdout, err := e.cmd.StdoutPipe()
	if err != nil {
		return err
	}

	e.cmd.Stderr = e.cmd.Stdout

	if err := e.cmd.Start(); err != nil {
		return err
	}

	e.status = "running"

	e.wg.Add(1) // 增加等待组计数
	go func() {
		defer e.wg.Done() // 在 goroutine 结束时减少等待组计数
		reader := bufio.NewReader(stdout)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					printf("读取错误: %v\n", err)
				}
				break
			}

			e.mutex.Lock()
			e.logBuffer = append(e.logBuffer, line)
			if len(e.logBuffer) > e.maxLogLines {
				e.logBuffer = e.logBuffer[len(e.logBuffer)-e.maxLogLines:]
			}
			e.mutex.Unlock()
		}

		// 等待命令执行完成
		e.cmd.Wait()

		e.mutex.Lock()
		e.status = "stopped"
		e.mutex.Unlock()
	}()

	return nil
}

func (e *CommandExecutor) Stop() error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if e.status != "running" {
		return errors.New("command is not running")
	}

	if err := e.cmd.Process.Kill(); err != nil {
		return err
	}

	e.status = "stopped"
	return nil
}

func (e *CommandExecutor) GetStatus() string {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	return e.status
}

func (e *CommandExecutor) GetLogs() []string {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	return append([]string(nil), e.logBuffer...)
}

func (e *CommandExecutor) Wait() {
	e.mutex.Lock()
	if e.status != "running" {
		e.mutex.Unlock()
		return
	}
	e.mutex.Unlock()
	e.wg.Wait()
}
