// Package view provides a layer of abstraction for all UI operations,
// decoupling the core logic from the presentation layer (e.g., pterm).
package view

import (
	"errors"
	"fmt"
	"os"

	"github.com/pterm/pterm"
)

// default view for the user that supports interactive prompting and has nice UI features such as spinners
type ptermView struct {
	multiPrinter *pterm.MultiPrinter
	spinners     map[string]*pterm.SpinnerPrinter
	isParallel   bool
}

func newDefaultView() RunnerView {
	return &ptermView{
		spinners:   make(map[string]*pterm.SpinnerPrinter),
		isParallel: false,
	}
}

func (v *ptermView) StartFile(file string) {
	pterm.DefaultSection.Println("Running 👟 " + file)
}

func (v *ptermView) EndFile(file string, err error) {
}

func (v *ptermView) StartStage(stageName string, chunkCount int, verbose bool) {
	if verbose {
		pterm.DefaultSection.WithLevel(2).Printf("stage %s with %d chunks\n", stageName, chunkCount)
	}
}

func (v *ptermView) Warning(message string) {
	pterm.Warning.Println(message)
}

func (v *ptermView) Error(message string) {
	pterm.Error.Println(message)
}

func (v *ptermView) InteractivePromptForCommand(prompt string, commandName string, isInteractive *bool) (string, error) {
	result, err := pterm.DefaultInteractiveContinue.WithDefaultText(commandName).Show()
	if err != nil {
		return "", err
	}
	if result == "all" {
		*isInteractive = false
	}
	return result, nil
}

func (v *ptermView) StartCommand(id, text string) error {
	var sp *pterm.SpinnerPrinter
	var err error
	if v.isParallel {

		if v.multiPrinter == nil {
			return errors.New("Start a parallel parallel requires entering parallel mode first")
		}
		sp, err = pterm.DefaultSpinner.WithWriter(v.multiPrinter.NewWriter()).Start(text)
	} else {
		sp, err = pterm.DefaultSpinner.Start(text)
	}
	if err != nil {
		return err
	}
	v.spinners[id] = sp
	return nil
}

func (v *ptermView) HasLogger(id string) bool {
	sp, _ := v.getSpinner(id)
	return sp != nil
}

func (v *ptermView) getSpinner(id string) (*pterm.SpinnerPrinter, error) {
	sp, ok := v.spinners[id]
	if !ok {
		return nil, errors.New("No spinner for id " + id)
	}
	return sp, nil
}

func (v *ptermView) StopCommand(id string, success bool, message string) error {
	sp, err := v.getSpinner(id)
	if err != nil {
		return err
	}
	if success {
		if message != "" {
			sp.Success(message)
		} else {
			sp.Success()
		}
	} else {
		sp.Fail(message)
	}
	delete(v.spinners, id)
	return nil
}

func (v *ptermView) DryRunCommand(id, text string) error {
	sp, err := v.getSpinner(id)
	if err != nil {
		return err
	}
	sp.InfoPrinter = &pterm.PrefixPrinter{
		MessageStyle: &pterm.Style{pterm.FgLightBlue},
		Prefix: pterm.Prefix{
			Style: &pterm.Style{pterm.FgBlack, pterm.BgLightBlue},
			Text:  " DRY-RUN ",
		},
	}
	sp.Info(text)
	delete(v.spinners, id)
	return nil
}

func (v *ptermView) SkipCommand(id, text string) error {
	sp, err := v.getSpinner(id)
	if err != nil {
		return err
	}
	sp.InfoPrinter = &pterm.PrefixPrinter{
		MessageStyle: &pterm.Style{pterm.FgLightBlue},
		Prefix: pterm.Prefix{
			Style: &pterm.Style{pterm.FgBlack, pterm.BgLightBlue},
			Text:  " SKIPPED ",
		},
	}
	sp.Info(text)
	delete(v.spinners, id)
	return nil
}

func (v *ptermView) Info(message string) {
	pterm.Info.Println(message)
}

func (v *ptermView) KillCommand(id, text string) error {
	sp, err := v.getSpinner(id)
	if err != nil {
		return err
	}
	sp.Fail(fmt.Sprintf("Killed %s", text))
	delete(v.spinners, id)
	return nil
}

func (v *ptermView) DeclareParallelMode() {
	v.isParallel = true
	v.multiPrinter = pterm.DefaultMultiPrinter.WithWriter(os.Stdout)
}

func (v *ptermView) StartParallelMode() error {
	if v.multiPrinter == nil {
		return errors.New("Declare parallel mode prior to start it")
	}
	_, err := v.multiPrinter.Start()
	if err != nil {
		return err
	}
	return nil
}

func (v *ptermView) QuitParallelMode() error {
	v.isParallel = false
	if v.multiPrinter == nil {
		return errors.New("Declare parallel mode prior to ")
	}
	if !v.multiPrinter.IsActive {
		return errors.New("Start the parallel mode before stopping it")
	}
	_, err := v.multiPrinter.Stop()
	if err != nil {
		return err
	}
	v.multiPrinter = nil
	return nil
}
