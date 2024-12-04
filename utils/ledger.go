package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/akerl/timber/v2/log"

	"github.com/akerl/ledgergraph/config"
)

var logger = log.NewLogger("ledgergraph.utils")

var registerCmd = []string{
	"register",
	"--cleared",
	"--sort=date",
	"--no-pager",
	"--format=%(date) %(account) %(quantity(amount)) %(quantity(total)) %(payee)\n",
}
var accountsCmd = []string{
	"accounts",
	"--no-pager",
}

// Transaction defines a ledger event
type Transaction struct {
	Time    time.Time
	Account string
	Payee   string
	Amount  float64
	Total   float64
}

var cache = []Transaction{}

// SyncLedger runs a loop to continously update the ledger file
func SyncLedger(c config.Config) {
	for {
		var err error
		cache, err = loadLedger(c)
		if err != nil {
			logger.InfoMsgf("failed automatic sync: %s", err)
		}
		time.Sleep(30 * time.Minute)
	}
}

// ReadLedgerFunc returns a function that reads the current state of the ledger file
func ReadLedgerFunc(c config.Config) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, _ *http.Request) {
		var err error
		if len(cache) == 0 {
			cache, err = loadLedger(c)
			if err != nil {
				http.Error(w, "failed to load ledger", http.StatusInternalServerError)
				logger.InfoMsgf("failed to load ledger: %s", err)
				return
			}
		}
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(cache)
		if err != nil {
			http.Error(w, "failed to marshal", http.StatusInternalServerError)
			logger.InfoMsgf("failed to marshal: %s", err)
			return
		}
	}
}

func loadLedger(c config.Config) ([]Transaction, error) {
	accounts, err := getAccounts(c)
	if err != nil {
		return []Transaction{}, err
	}
	logger.DebugMsgf("found %d accounts", len(accounts))

	t := []Transaction{}
	for _, account := range accounts {
		newT, err := getTransactions(c, account)
		if err != nil {
			return []Transaction{}, err
		}
		t = append(t, newT...)
	}
	return t, nil
}

func getAccounts(c config.Config) ([]string, error) {
	return runCommand(c, accountsCmd)
}

func getTransactions(c config.Config, account string) ([]Transaction, error) {
	logger.DebugMsgf("loading transactions from account: %s", account)
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
	logger.DebugMsgf("found %d transactions on account", len(t))
	return t, nil
}

func runCommand(c config.Config, cmdString []string) ([]string, error) {
	file := fmt.Sprintf("--file=%s", c.DataFile)
	args := append([]string{file}, cmdString...)
	cmd := exec.Command("ledger", args...)
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
