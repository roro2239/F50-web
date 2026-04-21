package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

var (
	baseDir    string
	runDir     string
	binDir     string
	configFile = "F50-web.conf"
)

func initRuntime() {
	if baseDir != "" {
		return
	}

	exePath, err := os.Executable()
	if err == nil {
		exeName := filepath.Base(exePath)
		if (exeName == "linker64" || exeName == "linker") && len(os.Args) > 0 {
			argPath := ""
			if len(os.Args) > 1 {
				argPath = os.Args[1]
			}
			if argPath == "" {
				argPath = os.Args[0]
			}
			if argPath != "" {
				if !filepath.IsAbs(argPath) {
					argPath, _ = filepath.Abs(argPath)
				}
				baseDir = filepath.Dir(argPath)
			}
		}
	}
	if baseDir == "" {
		if err != nil {
			baseDir = "."
		} else {
			baseDir = filepath.Dir(exePath)
		}
	}
	runDir = filepath.Join(baseDir, "runtime")
	binDir = filepath.Join(baseDir, "bin")
	configFile = filepath.Join(baseDir, "F50-web.conf")

	mustMkdir(runDir)
	mustMkdir(binDir)
}

func mustMkdir(path string) {
	if err := os.MkdirAll(path, 0755); err != nil {
		print("创建目录失败:", path, err)
	}
}

func runtimePath(name string) string {
	initRuntime()
	return filepath.Join(runDir, name)
}

func shellPath() string {
	if runtime.GOOS == "android" {
		return "/system/bin/sh"
	}
	if path, err := exec.LookPath("sh"); err == nil {
		return path
	}
	return "sh"
}

func buildRuntimePATH() string {
	parts := []string{binDir}
	currentPath := os.Getenv("PATH")
	if currentPath != "" {
		parts = append(parts, currentPath)
	}
	return strings.Join(parts, string(os.PathListSeparator))
}

func runtimeEnv() []string {
	initRuntime()
	return []string{
		"PATH=" + buildRuntimePATH(),
		"F50_WEB_BASE_DIR=" + baseDir,
		"F50_WEB_RUN_DIR=" + runDir,
		"F50_WEB_BIN_DIR=" + binDir,
		"F50_WEB_SHELL=" + shellPath(),
		"HOME=" + runDir,
		"TMPDIR=" + runDir,
	}
}

func bundledBinary(name string) string {
	initRuntime()
	return filepath.Join(binDir, name)
}

func shellQuote(path string) string {
	return strconvQuote(path)
}

func resolveCommand(name string) string {
	if path, err := exec.LookPath(name); err == nil {
		return path
	}
	return bundledBinary(name)
}

func strconvQuote(s string) string {
	return fmt.Sprintf("%q", s)
}
