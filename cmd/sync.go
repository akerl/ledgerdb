package cmd

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/akerl/ledgersql/config"
)

var registerCmd = []string{
	"ledger",
	"register",
	"--cleared",
	"--sort=date",
	"--no-pager",
	"--format='%(date) %(account) %(quantity(amount)) %(quantity(total)) %(payee)\n'",
}
var accountsCmd = []string{
	"ledger",
	"accounts",
	"--no-pager",
}

type transaction struct {
	Time    time.Time
	Account string
	Amount  float64
	Total   float64
	Payee   string
}

func syncRunner(_ *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("no config file provided")
	}

	c, err := config.NewConfig(args[0])
	if err != nil {
		return err
	}

	_, err = getLedger(c)
	if err != nil {
		return err
	}
	return nil
}

func getLedger(c config.Config) ([]transaction, error) {
	accounts, err := getAccounts(c)
	if err != nil {
		return []transaction{}, err
	}

	t := []transaction{}
	for _, account := range accounts {
		newT, err := getTransactions(c, account)
		if err != nil {
			return []transaction{}, err
		}
		t = append(t, newT...)
	}
	return t, nil
}

func getAccounts(c config.Config) ([]string, error) {
	return runCommand(c, accountsCmd)
}

func getTransactions(c config.Config, account string) ([]transaction, error) {
	joinedCmd := append(registerCmd, account)
	lines, err := runCommand(c, joinedCmd)
	if err != nil {
		return []transaction{}, err
	}

	t := make([]transaction, len(lines))

	for index, line := range lines {
		fields := strings.SplitN(line, " ", 5)
		date, err := time.Parse("2026/2/1", fields[0])
		if err != nil {
			return []transaction{}, err
		}
		amount, err := strconv.ParseFloat(fields[2], 64)
		if err != nil {
			return []transaction{}, err
		}

		total, err := strconv.ParseFloat(fields[3], 64)
		if err != nil {
			return []transaction{}, err
		}

		t[index] = transaction{
			Time:    date,
			Account: fields[1],
			Amount:  amount,
			Total:   total,
			Payee:   fields[4],
		}
	}
	return t, nil
}

func runCommand(c config.Config, cmdString []string) ([]string, error) {
	cmd := exec.Command(cmdString[0], cmdString[1:]...)
	cmd.Dir = c.DataDir

	var outb bytes.Buffer
	cmd.Stdout = &outb

	err := cmd.Run()
	if err != nil {
		return []string{}, err
	}

	lines := strings.Split(outb.String(), "\n")
	return lines, nil
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "sync ledger to database",
	RunE:  syncRunner,
}

func init() {
	rootCmd.AddCommand(syncCmd)
}
