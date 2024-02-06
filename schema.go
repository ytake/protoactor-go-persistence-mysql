package persistencemysql

type Table struct {
	name   string
	schema Schemaer
}

func NewTable(name string) *DefaultSchema {
	return &DefaultSchema{
		tableName: name,
	}
}

// Schemaer is the interface that wraps the basic methods for a schema.
type Schemaer interface {
	// TableName returns the name of the table.
	TableName() string
	// ID returns the name of the id column.
	ID() string
	// Payload returns the name of the payload column.
	Payload() string
	// ActorName returns the name of the actor name column.
	ActorName() string
	// EventIndex returns the name of the event index column.
	EventIndex() string
	// EventName returns the name of the event name column.
	EventName() string
	// Created returns the name of the created at column.
	Created() string
	// CreateTable returns the sql statement to create the table.
	CreateTable() string
}

// DefaultSchema is the default schema for the mysql provider.
type DefaultSchema struct {
	tableName string
}

// TableName returns the name of the table.
func (d *DefaultSchema) TableName() string {
	return d.tableName
}

// ID returns the name of the id column.
func (d *DefaultSchema) ID() string {
	return "id"
}

// Payload returns the name of the payload column.
func (d *DefaultSchema) Payload() string {
	return "payload"
}

// ActorName returns the name of the actor name column.
func (d *DefaultSchema) ActorName() string {
	return "actor_name"
}

// EventIndex returns the name of the event index column.
func (d *DefaultSchema) EventIndex() string {
	return "event_index"
}

// EventName returns the name of the event name column.
func (d *DefaultSchema) EventName() string {
	return "event_name"
}

// Created returns the name of the created at column.
func (d *DefaultSchema) Created() string {
	return "created_at"
}

// CreateTable returns the sql statement to create the table.
func (d *DefaultSchema) CreateTable() string {
	return "CREATE TABLE `" + d.tableName + "` (" +
		"`" + d.ID() + "` varchar(26) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL," +
		"`" + d.Payload() + "` json NOT NULL," +
		"`" + d.EventIndex() + "` bigint DEFAULT NULL," +
		"`" + d.ActorName() + "` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin DEFAULT NULL," +
		"`" + d.EventName() + "` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin DEFAULT NULL," +
		"`" + d.Created() + "` timestamp DEFAULT CURRENT_TIMESTAMP," +
		"PRIMARY KEY (`" + d.ID() + "`)," +
		"UNIQUE KEY `uidx_id` (`" + d.ID() + "`)," +
		"UNIQUE KEY `uidx_names` (`" + d.ActorName() + "`,`" + d.EventName() + "`,`" + d.EventIndex() + "`)" +
		") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin"
}
