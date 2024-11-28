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
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return err
	}

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

	return tx.Commit()
}

func createNewTable(tx *sql.Tx) error {
	if _, err := tx.Exec(fmt.Sprintf(`DROP TABLE IF EXISTS "%s"`, newTable)); err != nil {
		return err
	}
	_, err := tx.Exec(fmt.Sprintf(
		`CREATE TABLE "%s"
		date date NOT NULL,
		account text NOT NULL,
		payee text NOT NULL,
		amount money NOT NULL,
		total money NOT NULL`,
		newTable,
	))
	return err
}

func loadTransactions(tx *sql.Tx, t []Transaction) error {
	statement := fmt.Sprintf(
		`INSERT INTO "%s"
		(date, account, payee, amount, total)
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
	_, err := tx.Exec(
		fmt.Sprintf(`ALTER TABLE IF EXISTS "%s" RENAME "%s"`, realTable, oldTable),
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec(fmt.Sprintf(`ALTER TABLE "%s" RENAME "%s"`, newTable, realTable))
	if err != nil {
		return err
	}

	_, err = tx.Exec(fmt.Sprintf(`"DROP TABLE IF EXISTS "%s"`, oldTable))
	return err
}
