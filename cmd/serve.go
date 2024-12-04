package cmd

import (
	"fmt"
	"io/fs"
	"net/http"

	"github.com/akerl/timber/v2/log"
	"github.com/spf13/cobra"

	"github.com/akerl/ledgergraph/config"
	"github.com/akerl/ledgergraph/content"
	"github.com/akerl/ledgergraph/utils"
)

var logger = log.NewLogger("ledgergraph.sync")

func serveRunner(_ *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("no config file provided")
	}

	logger.DebugMsgf("loading config file: %s", args[0])
	c, err := config.NewConfig(args[0])
	if err != nil {
		return err
	}

	go utils.SyncLedger(c)

	subFS, err := fs.Sub(content.Static, "static")
	if err != nil {
		return err
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /data", utils.ReadLedgerFunc(c))
	mux.Handle("GET /", http.FileServer(http.FS(subFS)))
	return http.ListenAndServe(c.BindString(), mux)
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "serve ledger web UI",
	RunE:  serveRunner,
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
