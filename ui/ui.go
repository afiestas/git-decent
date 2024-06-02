package ui

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/afiestas/git-decent/config"
	"github.com/muesli/termenv"
)

type PrettyPrinter interface {
	PrettyPrint()
}

var verbose bool

var profile = termenv.ColorProfile()
var (
	PrimaryStyle   = termenv.Style{}.Foreground(termenv.ForegroundColor())
	SecondaryStyle = termenv.Style{}.Foreground(profile.Color("14")).Bold()
	AccentStyle    = termenv.Style{}.Foreground(profile.Color("11")).Bold()
	successStyle   = termenv.Style{}.Foreground(profile.Color("2")).Bold()
	warningStyle   = termenv.Style{}.Foreground(profile.Color("3"))
	ErrorStyle     = termenv.Style{}.Foreground(profile.Color("9")).Bold()
	InfoStyle      = termenv.Style{}.Foreground(profile.Color("12")).Bold()
	SoftStyle      = termenv.Style{}.Foreground(profile.Color("8"))
)

func SetVerbose(enabled bool) {
	verbose = enabled
}

func IsVerbose() bool {
	return verbose
}

func Info(title string, info string) {
	fmt.Println(title, SecondaryStyle.Styled(info))
}

func Title(str string) {
	fmt.Println(InfoStyle.Styled(str))
}

func Error(str string) {
	fmt.Println("❌", ErrorStyle.Styled(str))
}

func Debug(str ...string) {
	if !verbose {
		return
	}

	fmt.Println(SoftStyle.Styled(strings.Join(str, " ")))
}

func Print(str ...string) {
	fmt.Println(PrimaryStyle.Styled(strings.Join(str, " ")))
}

func YesNoQuestion(question string) (bool, error) {
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

func PrintSchedule(schedule config.Schedule) {
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

func PrintAmend(before time.Time, after time.Time, msg string) {
	sameDay := after.Day() == before.Day()
	sameTime := after.Minute() == before.Minute() && after.Hour() == before.Hour()

	fmt.Println("✨", msg)
	day := before.Format("Mon")
	if !sameDay {
		day = AccentStyle.Styled(day)
	}
	timeStr := before.Format("15:04")
	if !sameTime {
		timeStr = SecondaryStyle.Styled(timeStr)
	}

	fmt.Printf(
		"    %s %s %s ",
		before.Format(time.DateOnly),
		day,
		timeStr,
	)
	if after == before {
		fmt.Printf("✅")
	} else {
		day := after.Format("Mon")
		if !sameDay {
			day = AccentStyle.Styled(day)
		}
		time := after.Format("15:04")
		if !sameTime {
			time = SecondaryStyle.Styled(time)
		}
		fmt.Printf("➡️ %s %s",
			day,
			time,
		)
	}
	fmt.Println()
}

func PrintError(err error) {
	if pp, ok := err.(PrettyPrinter); ok {
		pp.PrettyPrint()
		return
	}

	if unwrap, ok := err.(interface{ Unwrap() []error }); ok {
		printErrors(unwrap.Unwrap())
		return
	}
	printError(err)
}

func printErrors(errs []error) {
	if len(errs) == 0 {
		return
	}

	fmt.Println("❌", PrimaryStyle.Styled(errs[0].Error()))
	for _, err := range errs[1:] {
		if pp, ok := err.(PrettyPrinter); ok {
			pp.PrettyPrint()
			continue
		}
		fmt.Println("   ", PrimaryStyle.Styled(err.Error()))
	}
}
func printError(err error) {
	fmt.Println("❌", PrimaryStyle.Styled(err.Error()))
}

var restoreConsole func() error

func Setup() error {
	var err error
	restoreConsole, err = termenv.EnableVirtualTerminalProcessing(termenv.DefaultOutput())
	return err
}

func TearDown() error {
	return restoreConsole()
}
