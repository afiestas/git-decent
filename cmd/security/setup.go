package security

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/afiestas/git-decent/ui"
)

var cleanup func()

func Setup(ui ui.UserInterface) error {
	err := recursivePreventionEnv()
	if err != nil {
		return err
	}

	err = recursivePreventLockfile(ui)
	if err != nil {
		return err
	}

	return nil
}

func TearDown() error {
	cleanup()
	return nil
}

func recursivePreventionEnv() error {
	if os.Getenv("GIT_AMEND_OPERATION") == "1" {
		return errors.New("prevented execution because of env GIT_AMEND_OPERATION = 1")
	}
	return nil
}

func recursivePreventLockfile(ui ui.UserInterface) error {
	fName := filepath.Join(os.TempDir(), "git-decent-hook-lock-file")
	f, err := os.OpenFile(fName, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0644)

	if err != nil {
		return fmt.Errorf("couldn't create lock file %w", err)
	}

	err = f.Close()
	if err != nil {

		return fmt.Errorf("couldn't close lock file %w", err)
	}

	cleanup = func() {
		if f != nil {
			fmt.Println("Cleaning up")
			os.Remove(f.Name())
		}
	}

	if ui.IsVerbose() {
		fmt.Println("Lock file", f.Name())
	}

	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Recovered from panic: %v\n", r)
			cleanup()
			os.Exit(1)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM, syscall.SIGABRT)
	go func() {
		sig := <-sigCh
		if ui.IsVerbose() {
			fmt.Printf("Received signal: %s, cleaning up...\n", sig)
		}

		cleanup()
		os.Exit(1)
	}()

	return nil
}
