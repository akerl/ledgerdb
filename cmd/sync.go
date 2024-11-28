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

const (
	registerCmd = []string{
		"ledger",
		"register",
		"--cleared",
		"--sort=date",
		"--no-pager",
		"--format='%(date) %(account) %(quantity(amount)) %(quantity(total)) %(payee)\n'",
	}
	accountsCmd = []string{
		"ledger",
		"accounts",
		"--no-pager",
	}
)

type transaction struct {
	Time    time.Time
	Account string
	Change  float64
	Total   float64
	Payee   string
}

func syncRunner(cmd *cobra.Command, args []string) error {
	flags := cmd.Flags()

	if len(args) < 1 {
		return fmt.Errorf("no config file provided")
	}

	c, err := config.NewConfig(args[0])
	if err != nil {
		return err
	}

	ledger, err := getLedger(c)
	if err != nil {
		return err
	}
}

func getLedger(c *ledgersql.Config) ([]transaction, error) {
	accounts, err := getAccounts(c)

	t := []transaction{}
	for _, account := range accounts {
		newT, err := getTransactions(c, account)
		if err != nil {
			return []transaction{}, err
		}
		t = append(t, newT)
	}
	return t, nil
}

func getAccounts(c *ledgersql.Config) ([]string, error) {
	return runCommand(accountsCmd)
}

func getTransactions(c *ledgersql.Config, account string) ([]transaction, error) {
	lines, err := runCommand(registerCmd)
	t := make([]transactions, len(lines))

	for index, line := range lines {
		fields := strings.SplitN(line, 5)
		date, err := time.Parse("2026/2/1", fields[0])
		if err != nil {
			return []transaction{}, err
		}
		t[index] = transaction{
			Time:    date,
			Account: fields[1],
			Amount:  strconv.ParseFloat(fields[2], 64),
			Total:   strconv.ParseFloat(fields[3], 64),
			Payee:   fields[4],
		}
	}
	return t, nil
}

func runCommand(c *ledgersql.Config, cmd []string) ([]string, err) {
	cmd := exec.Command(cmd[0], cmd[1:]...)
	e.Dir = c.DataDir

	var outb bytes.Buffer
	e.Stdout = &outb

	err := e.Run()
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
