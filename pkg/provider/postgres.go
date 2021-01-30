package provider

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
	"x-qdo/jiraclick/pkg/config"
	"x-qdo/jiraclick/pkg/model"
)

type postgresDB struct {
	conn         *pg.DB
	transactions map[context.Context]*pg.Tx
}

type queryFunc func(query *orm.Query)

func NewPostgres(cfg *config.Config) (*postgresDB, error) {
	postgresDB := new(postgresDB)

	opt, err := pg.ParseURL(cfg.Postgres.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Postgres: %s", err.Error())
	}

	if cfg.Postgres.Insecure {
		opt.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	} else {
		opt.TLSConfig = nil
	}

	db := pg.Connect(opt)

	if err := db.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping Postgres: %s", err.Error())
	}
	postgresDB.conn = db
	postgresDB.transactions = make(map[context.Context]*pg.Tx)

	return postgresDB, nil
}

func (db *postgresDB) Begin(ctx context.Context) error {
	tx, err := db.conn.BeginContext(ctx)
	if err != nil {
		return err
	}

	db.transactions[ctx] = tx

	return nil
}

func (db *postgresDB) Commit(ctx context.Context) error {
	if tx, found := db.transactions[ctx]; found {
		err := tx.Commit()
		if err != nil {
			return err
		}
		delete(db.transactions, ctx)
		return nil
	}
	return fmt.Errorf("transaction for context is not found")
}

func (db *postgresDB) Rollback(ctx context.Context) error {
	if tx, found := db.transactions[ctx]; found {
		err := tx.Rollback()
		if err != nil {
			return err
		}
		delete(db.transactions, ctx)
	}
	return nil
}

func (db *postgresDB) getConnection(ctx context.Context) orm.DB {
	if tx, found := db.transactions[ctx]; found {
		return tx
	}

	return db.conn
}

func (db *postgresDB) Close() error {
	return db.conn.Close()
}

func (db *postgresDB) queryApplyRelations(query *orm.Query, relations []string) {
	for _, relationName := range relations {
		query.Relation(relationName)
	}
}

func (db *postgresDB) modelGet(ctx context.Context, model interface{}, relations []string) error {
	query := db.getConnection(ctx).Model(model).WherePK()
	db.queryApplyRelations(query, relations)

	if err := query.Select(); err != nil {
		return err
	}

	return nil
}

func (db *postgresDB) modelQueryGet(ctx context.Context, model interface{}, queryFunc queryFunc) error {
	query := db.getConnection(ctx).Model(model).WherePK()

	queryFunc(query)

	if err := query.Select(); err != nil {
		return err
	}

	return nil
}

func (db *postgresDB) modelDelete(ctx context.Context, model interface{}) error {
	if _, err := db.getConnection(ctx).Model(model).WherePK().Delete(); err != nil {
		return err
	}

	return nil
}

func (db *postgresDB) modelExists(ctx context.Context, model interface{}) (bool, error) {
	if exists, err := db.getConnection(ctx).Model(model).WherePK().Exists(); err != nil {
		return false, err
	} else {
		return exists, nil
	}
}

func (db *postgresDB) modelInsert(ctx context.Context, model interface{}) error {
	if _, err := db.getConnection(ctx).Model(model).Insert(); err != nil {
		return err
	}

	return nil
}

func (db *postgresDB) modelUpdate(ctx context.Context, model interface{}) error {
	if _, err := db.getConnection(ctx).Model(model).WherePK().UpdateNotZero(); err != nil {
		return err
	}

	return nil
}

func (db *postgresDB) GetJiraAccounts(ctx context.Context) (map[string]model.JiraAccount, error) {
	var accounts []model.Account
	results := make(map[string]model.JiraAccount)
	query := db.getConnection(ctx).Model(&accounts)

	query.Where("resource = ?", "jira")

	if err := query.Select(); err != nil {
		return nil, err
	}

	for _, account := range accounts {
		var jiraAcc model.JiraAccount
		if props, err := json.Marshal(account.Props); err == nil {
			if err := json.Unmarshal(props, &jiraAcc); err == nil {
				results[account.SlackChannel] = jiraAcc
			}
		}
	}

	return results, nil
}

func (db *postgresDB) GetClickUpAccounts(ctx context.Context) (map[string]model.ClickUpAccount, error) {
	var accounts []model.Account
	results := make(map[string]model.ClickUpAccount)
	query := db.getConnection(ctx).Model(&accounts)

	query.Where("resource = ?", "clickup")

	if err := query.Select(); err != nil {
		return nil, err
	}

	for _, account := range accounts {
		var clickAcc model.ClickUpAccount
		if props, err := json.Marshal(account.Props); err == nil {
			if err := json.Unmarshal(props, &clickAcc); err == nil {
				results[account.SlackChannel] = clickAcc
			}
		}
	}

	return results, nil
}
