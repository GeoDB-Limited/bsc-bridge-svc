package postgres

import (
	"database/sql"
	sq "github.com/Masterminds/squirrel"
	"github.com/bsc-bridge-svc/internal/config"
	"github.com/bsc-bridge-svc/internal/data"
	"github.com/pkg/errors"
)

type Transfers interface {
	New() Transfers
	SelectStatus(data.Status) ([]data.Transfer, error)
	Select() ([]data.Transfer, error)
	CreateTransfer(data.Transfer) error
	UpdateTransfer(data.Transfer) error
}

type transfers struct {
	db  *sql.DB
	sql sq.SelectBuilder
}

const (
	transfersTable = "transfers"
)

var transfersSelect = sq.Select(all).From(transfersTable).PlaceholderFormat(sq.Dollar)

func NewTransfers(cfg config.Config) Transfers {
	return &transfers{
		db:  cfg.DB(),
		sql: transfersSelect.RunWith(cfg.DB()),
	}
}

func (t *transfers) New() Transfers {
	return &transfers{
		db:  t.db,
		sql: transfersSelect.RunWith(t.db),
	}
}

func (t *transfers) newInsert() sq.InsertBuilder {
	return sq.Insert(transfersTable).RunWith(t.db).PlaceholderFormat(sq.Dollar)
}

func (t *transfers) CreateTransfer(transfer data.Transfer) error {
	_, err := t.newInsert().SetMap(transfer.ToMap()).Exec()
	if err != nil {
		return errors.Wrap(err, "failed to insert transfer")
	}
	return nil
}

func (t *transfers) newUpdate() sq.UpdateBuilder {
	return sq.Update(transfersTable).RunWith(t.db).PlaceholderFormat(sq.Dollar)
}

func (t *transfers) UpdateTransfer(transfer data.Transfer) error {
	_, err := t.newUpdate().SetMap(transfer.ToMap()).Where(sq.Eq{"id": transfer.ID}).Exec()
	if err != nil {
		return errors.Wrap(err, "failed to update transfer data")
	}
	return nil
}

func (t *transfers) Select() ([]data.Transfer, error) {
	rows, err := t.sql.Query()
	if err != nil {
		return nil, errors.Wrap(err, "failed to query rows")
	}
	result := make([]data.Transfer, 0)

	for rows.Next() {
		transfer := data.Transfer{}
		err = rows.Scan(
			&transfer.ID,
			&transfer.Address,
			&transfer.Amount,
			&transfer.Denom,
			&transfer.Status,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan rows")
		}
		result = append(result, transfer)
	}

	if len(result) == 0 {
		return nil, nil
	}

	return result, nil
}

func (t *transfers) SelectStatus(status data.Status) ([]data.Transfer, error) {
	t.sql = t.sql.Where(sq.Eq{"status": status})
	return t.Select()
}
