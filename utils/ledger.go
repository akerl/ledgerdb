package utils

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/akerl/ledgersql/config"
)

var registerCmd = []string{
	"ledger",
	"register",
	"--cleared",
	"--sort=date",
	"--no-pager",
	"--format=%(date) %(account) %(quantity(amount)) %(quantity(total)) %(payee)\n",
}
var accountsCmd = []string{
	"ledger",
	"accounts",
	"--no-pager",
}

// Transaction defines a ledger event
type Transaction struct {
	Time    time.Time
	Account string
	Amount  float64
	Total   float64
	Payee   string
}

// GetLedger returns all ledgers and transactions
func GetLedger(c config.Config) ([]string, []Transaction, error) {
	accounts, err := getAccounts(c)
	if err != nil {
		return []string{}, []Transaction{}, err
	}

	t := []Transaction{}
	for _, account := range accounts {
		newT, err := getTransactions(c, account)
		if err != nil {
			return []string{}, []Transaction{}, err
		}
		t = append(t, newT...)
	}
	return accounts, t, nil
}

func getAccounts(c config.Config) ([]string, error) {
	return runCommand(c, accountsCmd)
}

func getTransactions(c config.Config, account string) ([]Transaction, error) {
	joinedCmd := append(registerCmd, account)
	lines, err := runCommand(c, joinedCmd)
	if err != nil {
		return []Transaction{}, err
	}

	t := make([]Transaction, len(lines))

	for index, line := range lines {
		fields := strings.SplitN(line, " ", 5)
		date, err := time.Parse("2006/01/02", fields[0])
		if err != nil {
			fmt.Println(len(lines))
			return []Transaction{}, err
		}
		amount, err := strconv.ParseFloat(fields[2], 64)
		if err != nil {
			return []Transaction{}, err
		}

		total, err := strconv.ParseFloat(fields[3], 64)
		if err != nil {
			return []Transaction{}, err
		}

		t[index] = Transaction{
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

	var outb strings.Builder
	cmd.Stdout = &outb

	err := cmd.Run()
	if err != nil {
		return []string{}, err
	}

	trimmed := strings.TrimSpace(outb.String())
	if len(trimmed) == 0 {
		return []string{}, nil
	}

	lines := strings.Split(trimmed, "\n")
	return lines, nil
}
