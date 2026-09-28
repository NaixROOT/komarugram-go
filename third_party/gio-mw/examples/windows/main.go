// SPDX-License-Identifier: Unlicense OR MIT

package main

import (
	"context"
	"gio-mw/examples/windows/shell"
	"os"
	"os/signal"

	"gioui.org/app"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	go func() {
		newApplication := shell.NewApplication(ctx)

		shell.NewMainShell(newApplication)
		shell.NewMainShell(newApplication)

		newApplication.Wait()

		os.Exit(0)
	}()
	app.Main()
}
