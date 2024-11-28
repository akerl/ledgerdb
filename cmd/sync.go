package cmd

import (
	"fmt"

	"github.com/akerl/timber/v2/log"
	"github.com/spf13/cobra"

	"github.com/akerl/ledgersql/config"
	"github.com/akerl/ledgersql/utils"
)

var logger = log.NewLogger("ledgersql.sync")

func syncRunner(_ *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("no config file provided")
	}

	logger.DebugMsgf("loading config file: %s", args[0])
	c, err := config.NewConfig(args[0])
	if err != nil {
		return err
	}

	accounts, transactions, err := utils.GetLedger(c)
	if err != nil {
		return err
	}
	logger.InfoMsgf("parsed %d transactions from %d accounts", len(transactions), len(accounts))

	return utils.WriteSQL(c, transactions)
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "sync ledger to database",
	RunE:  syncRunner,
}

func init() {
	rootCmd.AddCommand(syncCmd)
}
