// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package browser

import (
	"errors"
	"os"
	"os/exec"
	"runtime"
	"time"
)

var cliBrowsers = map[string]struct{}{
	"www-browser": {},
	"lynx":        {},
	"w3m":         {},
	"links":       {},
	"elinks":      {},
}

func Browsers() []string {
	return []string{
		//GUI browsers
		"chrome",
		"google-chrome",
		"chromium",
		"firefox",

		//CLI browsers
		"www-browser",
		"lynx",
		"w3m",
		"links",
		"elinks",
	}
}

func Other() []string {
	return []string{
		"termux-open-url",
	}
}

// Commands returns a list of possible commands to use to open a url.
func Commands() [][]string {
	var cmds [][]string
	if exe := os.Getenv("BROWSER"); exe != "" {
		cmds = append(cmds, []string{exe})
	}
	switch runtime.GOOS {
	case "darwin":
		cmds = append(cmds, []string{"/usr/bin/open"})
	case "windows":
		cmds = append(cmds, []string{"cmd", "/c", "start"})
	default:
		if os.Getenv("DISPLAY") != "" {
			// xdg-open is only for use in a desktop environment.
			cmds = append(cmds, []string{"xdg-open"})
		}
	}
	for _, browsers := range [][]string{Browsers(), Other()} {
		for _, browser := range browsers {
			cmds = append(cmds, []string{browser})
		}
	}
	return cmds
}

func MustOpen(url string) error {
	if !Open(url) {
		return openBrowserFailErr()
	}
	return nil
}

func Open(url string) bool {
	for _, args := range Commands() {
		cmd := exec.Command(args[0], append(args[1:], url)...)
		if _, ok := cliBrowsers[args[0]]; ok {
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err == nil {
				return true
			} else if _, ok := err.(*exec.ExitError); ok {
				return true
			}
		} else if cmd.Start() == nil && appearsSuccessful(cmd, 3*time.Second) {
			return true
		}
	}
	return false
}

// appearsSuccessful reports whether the command appears to have run successfully.
// If the command runs longer than the timeout, it's deemed successful.
// If the command runs within the timeout, it's deemed successful if it exited cleanly.
func appearsSuccessful(cmd *exec.Cmd, timeout time.Duration) bool {
	errc := make(chan error, 1)
	go func() {
		errc <- cmd.Wait()
	}()

	select {
	case <-time.After(timeout):
		return true
	case err := <-errc:
		return err == nil
	}
}

func errorsNewWrapper(message string) error {
	return errors.New(message)
}

func openBrowserFailErr() error {
	return errorsNewWrapper("Failed to open web browser.")
}
