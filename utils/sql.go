package utils

import (
	"database/sql"
	"fmt"
	"time"

	// using Postgres for SQL
	_ "github.com/lib/pq"

	"github.com/akerl/ledgersql/config"
)

const (
	realTable = "ledger"
	newTable  = "newledger"
	oldTable  = "oldledger"
)

// WriteSQL loads transactions into the database
func WriteSQL(c config.Config, t []Transaction) error {
	connStr := fmt.Sprintf(
		"dbname=ledgerdb user=admin password=%s host=%s sslmode=require",
		c.DatabasePassword,
		c.DatabaseHost,
	)
	logger.DebugMsgf("connecting to database: %s", c.DatabaseHost)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return err
	}

	logger.DebugMsg("starting transaction")
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := createNewTable(tx); err != nil {
		return err
	}

	if err := loadTransactions(tx, t); err != nil {
		return err
	}

	if err := swapTables(tx); err != nil {
		return err
	}

	logger.DebugMsg("committing transaction")
	return tx.Commit()
}

func createNewTable(tx *sql.Tx) error {
	logger.DebugMsg("dropping stale new table if it exists")
	if _, err := tx.Exec(fmt.Sprintf(`DROP TABLE IF EXISTS "%s"`, newTable)); err != nil {
		return err
	}
	logger.DebugMsg("creating new table")
	_, err := tx.Exec(fmt.Sprintf(
		`CREATE TABLE "%s" (
			time date NOT NULL,
			account text NOT NULL,
			payee text NOT NULL,
			amount numeric NOT NULL,
			total numeric NOT NULL
		)`,
		newTable,
	))
	return err
}

func loadTransactions(tx *sql.Tx, t []Transaction) error {
	logger.DebugMsgf("inserting %d transactions", len(t))
	statement := fmt.Sprintf(
		`INSERT INTO "%s"
		(time, account, payee, amount, total)
		VALUES ($1, $2, $3, $4, $5)`,
		newTable,
	)
	for _, item := range t {
		_, err := tx.Exec(
			statement,
			item.Time.Format(time.DateOnly),
			item.Account,
			item.Payee,
			item.Amount,
			item.Total,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func swapTables(tx *sql.Tx) error {
	logger.DebugMsg("renaming real table if it exists")
	_, err := tx.Exec(
		fmt.Sprintf(`ALTER TABLE IF EXISTS "%s" RENAME TO "%s"`, realTable, oldTable),
	)
	if err != nil {
		return err
	}

	logger.DebugMsg("moving new table to real table")
	_, err = tx.Exec(fmt.Sprintf(`ALTER TABLE "%s" RENAME TO "%s"`, newTable, realTable))
	if err != nil {
		return err
	}

	logger.DebugMsg("dropping old table")
	_, err = tx.Exec(fmt.Sprintf(`DROP TABLE IF EXISTS "%s"`, oldTable))
	return err
}
