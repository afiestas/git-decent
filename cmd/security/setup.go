package security

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	u "github.com/afiestas/git-decent/utils"
)

var cleanup func()

func Setup() error {
	err := recursivePreventionEnv()
	if err != nil {
		return err
	}

	err = recursivePreventLockfile()
	if err != nil {
		return err
	}

	return nil
}

func TearDown() error {
	if cleanup != nil {
		cleanup()
	}

	return nil
}

func recursivePreventionEnv() error {
	if os.Getenv("GIT_AMEND_OPERATION") == "1" {
		return errors.New("prevented execution because of env GIT_AMEND_OPERATION = 1")
	}
	return nil
}

func recursivePreventLockfile() error {
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
			u.Debug("Cleaning up")
			os.Remove(f.Name())
		}
	}

	u.Debug("Lock file:", f.Name())

	defer func() {
		if r := recover(); r != nil {
			u.Debug(fmt.Sprintf("Recovered from panic: %v\n", r))
			cleanup()
			os.Exit(1)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM, syscall.SIGABRT)
	go func() {
		sig := <-sigCh
		u.Debug(fmt.Sprintf("Received signal: %s, cleaning up...\n", sig))

		cleanup()
		os.Exit(1)
	}()

	return nil
}
