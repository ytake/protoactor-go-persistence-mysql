package persistencemysql

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/asynkron/protoactor-go/persistence"
	"github.com/oklog/ulid/v2"
	"google.golang.org/protobuf/proto"
)

// Provider is the abstraction used for persistence
type Provider struct {
	tableSchema      Schemaer
	snapshotInterval int
	db               *sql.DB
	logger           *slog.Logger
}

// New creates a new mysql provider
func New(snapshotInterval int, table Schemaer, db *sql.DB, logger *slog.Logger) (*Provider, error) {
	return &Provider{
		tableSchema:      table,
		snapshotInterval: snapshotInterval,
		db:               db,
		logger:           logger,
	}, nil
}

// DeleteEvents removes all events from the provider
func (provider *Provider) DeleteEvents(_ string, _ int) {
}

func (provider *Provider) Restart() {
}

func (provider *Provider) GetSnapshotInterval() int {
	return provider.snapshotInterval
}

func (provider *Provider) selectColumns() string {
	return strings.Join([]string{
		provider.tableSchema.ID(),
		provider.tableSchema.Payload(),
		provider.tableSchema.SequenceNumber(),
		provider.tableSchema.ActorName(),
	}, ",")
}

func (provider *Provider) GetEvents(actorName string, eventIndexStart int, eventIndexEnd int, callback func(e interface{})) {
	tx, _ := provider.db.Begin()
	defer tx.Commit()
	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE %s = ? AND %s BETWEEN ? AND ? ORDER BY %s ASC",
		provider.selectColumns(),
		provider.tableSchema.JournalTableName(),
		provider.tableSchema.ActorName(),
		provider.tableSchema.SequenceNumber(),
		provider.tableSchema.SequenceNumber(),
	)
	args := []interface{}{actorName, eventIndexStart, eventIndexEnd}
	if eventIndexEnd == 0 {
		query = fmt.Sprintf(
			"SELECT %s FROM %s WHERE %s = ? AND %s >= ? ORDER BY %s ASC",
			provider.selectColumns(),
			provider.tableSchema.JournalTableName(),
			provider.tableSchema.ActorName(),
			provider.tableSchema.SequenceNumber(),
			provider.tableSchema.SequenceNumber())
		args = []interface{}{actorName, eventIndexStart}
	}
	rows, err := tx.Query(query, args...)
	if !errors.Is(err, sql.ErrNoRows) && err != nil {
		provider.logger.Error(err.Error(), slog.String("actor_name", actorName))
		return
	}
	for rows.Next() {
		env := &envelope{}
		if err := rows.Scan(&env.ID, &env.Message, &env.EventIndex, &env.ActorName); err != nil {
			return
		}
		m, err := env.message()
		if err != nil {
			provider.logger.Error(err.Error(), slog.String("actor_name", actorName))
			return
		}
		callback(m)
	}
}

// 'executeTx' is a function that manages the lifecycle of a DB transaction.
// It takes a function 'op' that contains DB transaction operation to be executed.
func (provider *Provider) executeTx(op func(tx *sql.Tx) error) (err error) {
	// Start transaction
	var tx *sql.Tx
	tx, err = provider.db.Begin()
	defer tx.Rollback()
	if err != nil {
		return err
	}
	// Execute operation
	err = op(tx)
	if err != nil {
		return err
	}
	// Everything went fine
	return tx.Commit()
}

func (provider *Provider) PersistEvent(actorName string, eventIndex int, snapshot proto.Message) {
	bytes, err := newByte(snapshot)
	if err != nil {
		provider.logger.Error(
			fmt.Sprintf("persistence error: %s", err), slog.String("actor_name", actorName))
		return
	}
	err = provider.executeTx(func(tx *sql.Tx) error {
		stmt, err := tx.Prepare(
			fmt.Sprintf(
				"INSERT INTO %s (%s) VALUES (?, ?, ?, ?)",
				provider.tableSchema.JournalTableName(), provider.selectColumns()))
		if err != nil {
			return err
		}
		_, err = stmt.Exec(ulid.Make().String(), string(bytes), eventIndex, actorName)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		provider.logger.Error(
			fmt.Sprintf("persistence event / sql error: %s", err.Error()),
			slog.String("actor_name", actorName))
		return
	}
}

func (provider *Provider) PersistSnapshot(actorName string, eventIndex int, snapshot proto.Message) {
	bytes, err := newByte(snapshot)
	if err != nil {
		provider.logger.Error(
			fmt.Sprintf("persistence error: %s", err), slog.String("actor_name", actorName))
		return
	}
	err = provider.executeTx(func(tx *sql.Tx) error {
		stmt, err := tx.Prepare(
			fmt.Sprintf(
				"INSERT INTO %s (%s) VALUES (?, ?, ?, ?)",
				provider.tableSchema.SnapshotTableName(), provider.selectColumns()))
		if err != nil {
			return err
		}
		_, err = stmt.Exec(ulid.Make().String(), string(bytes), eventIndex, actorName)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		provider.logger.Error(
			fmt.Sprintf("persistence snapshot / sql error: %s", err.Error()),
			slog.String("actor_name", actorName))
		return
	}
}

func (provider *Provider) GetSnapshot(actorName string) (snapshot interface{}, eventIndex int, ok bool) {
	tx, _ := provider.db.Begin()
	defer tx.Commit()
	rows, err := tx.Query(
		fmt.Sprintf(
			"SELECT %s FROM %s WHERE %s = ? ORDER BY %s DESC",
			provider.selectColumns(),
			provider.tableSchema.SnapshotTableName(),
			provider.tableSchema.ActorName(),
			provider.tableSchema.SequenceNumber(),
		),
		actorName)
	defer rows.Close()
	if !errors.Is(err, sql.ErrNoRows) && err != nil {
		provider.logger.Error(err.Error(), slog.String("actor_name", actorName))
		return nil, 0, false
	}
	for rows.Next() {
		env := envelope{}
		if err := rows.Scan(&env.ID, &env.Message, &env.EventIndex, &env.ActorName); err != nil {
			return nil, 0, false
		}
		m, err := env.message()
		if err != nil {
			provider.logger.Error(err.Error(), slog.String("actor_name", actorName))
			return nil, 0, false
		}
		return m, env.EventIndex, true
	}
	return nil, 0, false
}

func (provider *Provider) DeleteSnapshots(_ string, _ int) {
}

func (provider *Provider) GetState() persistence.ProviderState {
	return provider
}
