package ui

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/afiestas/git-decent/config"
	"github.com/afiestas/git-decent/internal"
	"github.com/muesli/termenv"
)

type UserInterface struct {
	verbose bool
}

var Ui UserInterface

var profile = termenv.ColorProfile()
var (
	PrimaryStyle   = termenv.Style{}.Foreground(termenv.ForegroundColor())
	SecondaryStyle = termenv.Style{}.Foreground(profile.Color("14")).Bold()
	AccentStyle    = termenv.Style{}.Foreground(profile.Color("11")).Bold()
	successStyle   = termenv.Style{}.Foreground(profile.Color("2")).Bold()
	warningStyle   = termenv.Style{}.Foreground(profile.Color("3"))
	ErrorStyle     = termenv.Style{}.Foreground(profile.Color("9")).Bold()
	InfoStyle      = termenv.Style{}.Foreground(profile.Color("12")).Bold()
)

func (l *UserInterface) IsVerbose() bool {
	return l.verbose
}

func (l *UserInterface) SetVerbose(verbose bool) {
	l.verbose = verbose
}

func (l *UserInterface) Info(title string, info string) {
	fmt.Println(title, SecondaryStyle.Styled(info))
}

func (l *UserInterface) Title(str string) {
	fmt.Println(InfoStyle.Styled(str))
}

func (l *UserInterface) Error(str string) {
	fmt.Println("❌", ErrorStyle.Styled(str))
}

func (l *UserInterface) Debug(str string) {
	if !l.verbose {
		return
	}

	fmt.Println(str)
}

func (l *UserInterface) YesNoQuestion(question string) (bool, error) {
	fmt.Println(PrimaryStyle.Styled(question), PrimaryStyle.Bold().Styled("(Y/n)"))

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}

	input = strings.TrimSpace(input)

	switch strings.ToLower(input) {
	case "n":
		return false, nil
	default:
		return true, err
	}
}

func (l *UserInterface) PrintSchedule(schedule config.Schedule) {
	for x := time.Monday; x <= time.Saturday; x++ {
		s := schedule.Days[x].DecentFrames.String()
		if len(s) == 0 {
			s = "↪️ " + schedule.Days[x].ClosestDecentDay.String()
		}
		fmt.Printf("📅 %-10s %s\n", x.String()+":", s)
	}
	s := schedule.Days[0].DecentFrames.String()
	if len(s) == 0 {
		s = "↪️ " + schedule.Days[0].ClosestDecentDay.String()
	}
	fmt.Printf("📅 %-10s %s\n", time.Sunday.String()+":", s)
}

func (l *UserInterface) PrintError(err error) {
	var commandError *internal.CommandError
	switch {
	case errors.As(err, &commandError):
		fmt.Println("   ", SecondaryStyle.Bold().Styled("Command:"), PrimaryStyle.Styled(commandError.Command))
		if len(commandError.Stdout) > 0 {
			fmt.Println("   ", SecondaryStyle.Bold().Styled("Stdout:"), PrimaryStyle.Styled(commandError.Stdout))
		}
		if len(commandError.Stderr) > 0 {
			fmt.Println("   ", SecondaryStyle.Bold().Styled("Stderr:"), PrimaryStyle.Styled(commandError.Stderr))
		}
	default:
		fmt.Println("   ", PrimaryStyle.Styled(err.Error()))
	}

}
