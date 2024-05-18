/* SPDX-License-Identifier: MIT */
package cmd

import (
	"bufio"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/afiestas/git-decent/config"
	"github.com/afiestas/git-decent/internal"
	"github.com/muesli/termenv"
	"github.com/spf13/cobra"
)

//go:embed config-template.ini
var configTemplate string

var profile = termenv.ColorProfile()
var (
	primaryStyle   = termenv.Style{}.Foreground(termenv.ForegroundColor())
	secondaryStyle = termenv.Style{}.Foreground(profile.Color("14")).Bold()
	accentStyle    = termenv.Style{}.Foreground(profile.Color("11")).Bold()
	successStyle   = termenv.Style{}.Foreground(profile.Color("2")).Bold()
	warningStyle   = termenv.Style{}.Foreground(profile.Color("3"))
	errorStyle     = termenv.Style{}.Foreground(profile.Color("9")).Bold()
	infoStyle      = termenv.Style{}.Foreground(profile.Color("12")).Bold()
)

var rootCmd = &cobra.Command{
	Use:   "git-decent",
	Short: "Ammends your commits so it looks like you are behaving",
	Long: `Git-Decent is a small tool designed to help you, night owls,
maintain appearances while working during unconventional hours...`,

	Run: func(cmd *cobra.Command, args []string) {
		restoreConsole, err := termenv.EnableVirtualTerminalProcessing(termenv.DefaultOutput())
		if err != nil {
			panic(err)
		}
		defer restoreConsole()

		r, err := getRepo()
		if err != nil {
			return
		}

		ops, _ := r.GetSectionOptions("decent")

		if len(ops) == 0 {
			asnwer, err := yesNoQuestion("Git decent is not configured, do you want to do it now?")
			if err != nil {
				printError(err)
				return
			}
			if !asnwer {
				return
			}

			initConfiguration(r)
			ops, _ = r.GetSectionOptions("decent")
		}

		rc, _ := config.GetGitRawConfig(&ops)
		s, _ := config.NewScheduleFromRaw(&rc)
		fmt.Println(infoStyle.Styled("Schedule:"))
		printSchedule(s)
		fmt.Println()

		_, err = r.LogWithRevision("-1")
		if err != nil {
			fmt.Println(errorStyle.Styled("❌ Couldn't get log"))
			printError(err)
			return
		}

		fmt.Println(infoStyle.Styled("Current status"))
		upstream := r.BranchUpstream(r.CurrentBranch())
		fmt.Println("Upstream branch", secondaryStyle.Styled(upstream))
		aLog := fmt.Sprintf("%s...", upstream)
		log, err := r.LogWithRevision(aLog)
		if err != nil {
			printError(err)
			return
		}

		fmt.Println("Unpushed commits:", secondaryStyle.Styled(fmt.Sprintf("%d", len(log))))
		if len(log) == 0 {
			return
		}

		amendedCount := 0
		var lastDate *time.Time = nil
		for k, commit := range log {
			if commit.Prev != nil {
				commit.Prev = log[k-1]
				lastDate = &commit.Prev.Date
			}
			commitDate := commit.Date
			amended := internal.Amend(commitDate, lastDate, s)
			sameDay := amended.Day() == commit.Date.Day()
			sameTime := amended.Minute() == commitDate.Minute() && amended.Hour() == commitDate.Hour()

			fmt.Println("✨", commit.Message)
			day := commitDate.Format("Mon")
			if !sameDay {
				day = accentStyle.Styled(day)
			}
			timeStr := commitDate.Format("15:04")
			if !sameTime {
				timeStr = secondaryStyle.Styled(timeStr)
			}

			fmt.Printf(
				"    %s %s %s ",
				commitDate.Format(time.DateOnly),
				day,
				timeStr,
			)
			if amended == commit.Date {
				fmt.Printf("✅")
			} else {
				amendedCount += 1
				day := amended.Format("Mon")
				if !sameDay {
					day = accentStyle.Styled(day)
				}
				time := amended.Format("15:04")
				if !sameTime {
					time = secondaryStyle.Styled(time)
				}
				fmt.Printf("➡️ %s %s",
					day,
					time,
				)
			}
			fmt.Println()

			commit.Date = amended
			log[k] = commit
		}

		fmt.Println("Amended commits:", secondaryStyle.Styled(fmt.Sprintf("%d", amendedCount)))
		if amendedCount == 0 {
			return
		}

		answer, err := yesNoQuestion("Do you want to ament the dates?")
		if err != nil {
			printError(err)
			return
		}

		if !answer {
			return
		}

		err = r.AmendDates(log)
		if err != nil {
			fmt.Println("❌", errorStyle.Styled("Error amending the dates"))
			printError(err)
		}
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}

}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func getRepo() (*internal.GitRepo, error) {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println("❌", errorStyle.Styled("Couldn't get cwd"), err)
		return nil, err
	}
	r, err := internal.NewGitRepo(cwd)
	if err != nil {
		fmt.Println("❌", errorStyle.Styled("Couldn't open the repository"), err)
		return nil, err
	}
	if !r.IsGitRepo() {
		fmt.Println(errorStyle.Styled("❌ Not a git repository"))
		fmt.Println("The directory", secondaryStyle.Styled(cwd), "does not appear to be a git repo")
		return nil, err
	}

	return r, nil
}

func initConfiguration(repo *internal.GitRepo) error {
	rawC, err := openGitEditor()
	if err != nil {
		return err
	}

	for x := time.Monday; x < time.Saturday; x++ {
		if len(rawC.Days[x]) == 0 {
			continue
		}
		err = repo.SetConfig("decent."+strings.Title(x.String()), rawC.Days[x])
		if err != nil {
			return err
		}
	}
	if len(rawC.Days[time.Sunday]) > 0 {
		err = repo.SetConfig("decent."+strings.Title(time.Sunday.String()), rawC.Days[time.Sunday])
		if err != nil {
			return err
		}
	}
	return nil
}

func openGitEditor() (*config.RawScheduleConfig, error) {
	gitcfg := exec.Command("git", "var", "GIT_EDITOR")
	editorName, err := gitcfg.Output()
	if err != nil {
		return nil, fmt.Errorf("openGitEditor coudn't fetch the GIT_EDITOR var")
	}

	if len(editorName) == 0 {
		return nil, fmt.Errorf("openGitEditor empty editor configured")
	}

	var args []string

	editorCmd := strings.Fields(string(editorName))
	cmd := editorCmd[0]
	if len(editorCmd) > 0 {
		args = editorCmd[1:]
	}

	f, err := os.CreateTemp(os.TempDir(), "schedule-tempalte")
	if err != nil {
		return nil, fmt.Errorf("openGitEditor can't create tmp file %w", err)
	}

	defer func() {
		f.Close()
		os.RemoveAll(f.Name())
	}()

	f.WriteString(configTemplate)
	f.Seek(0, 0)

	args = append(args, f.Name())
	for {
		editor := exec.Command(cmd, args...)
		editor.Stdin = os.Stdin
		editor.Stdout = os.Stdout
		editor.Stderr = os.Stderr

		err = editor.Run()
		if err != nil {
			return nil, err
		}

		rawC, err := config.NewScheduleFromPlainText(f)
		if err == nil {
			_, err := config.NewScheduleFromRaw(rawC)
			if err == nil {
				return rawC, nil
			}
		}

		fmt.Println("the configuration coudln't be parsed", err)
		answer, err := yesNoQuestion("Do you want to edit it again?")
		if err != nil {
			return nil, err
		}
		if !answer {
			return nil, nil
		}
	}
}

func yesNoQuestion(question string) (bool, error) {
	fmt.Println(primaryStyle.Styled(question), primaryStyle.Bold().Styled("(Y/n)"))

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

func printSchedule(schedule config.Schedule) {
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

func printError(err error) {
	var commandError *internal.CommandError
	switch {
	case errors.As(err, &commandError):
		fmt.Println("   ", secondaryStyle.Bold().Styled("Command:"), primaryStyle.Styled(commandError.Command))
		if len(commandError.Stdout) > 0 {
			fmt.Println("   ", secondaryStyle.Bold().Styled("Stdout:"), primaryStyle.Styled(commandError.Stdout))
		}
		if len(commandError.Stderr) > 0 {
			fmt.Println("   ", secondaryStyle.Bold().Styled("Stderr:"), primaryStyle.Styled(commandError.Stderr))
		}
	default:
		fmt.Println("   ", primaryStyle.Styled(err.Error()))
	}

}
