package utils

import (
	"errors"
	"fmt"
	"strings"

	"github.com/afiestas/git-decent/ui"
)

type DebugBlock struct {
	Title string
	Lines []string
}

func (db *DebugBlock) AddLine(name string, parts ...string) {
	if !ui.IsVerbose() {
		return
	}
	line := fmt.Sprintf("%s: %s", name, strings.Join(parts, " "))
	db.Lines = append(db.Lines, line)
}

func (db *DebugBlock) Print() {
	if !ui.IsVerbose() {
		return
	}

	ui.Debug(db.Title)
	for _, line := range db.Lines {
		ui.Debug("    " + strings.TrimSpace(line))
	}
}

func WrapE(msg string, errs ...error) error {
	newErr := errors.New(msg)
	errs = append([]error{newErr}, errs...)
	return errors.Join(errs...)
}

func Debug(str ...string) {
	ui.Debug(str...)
}
