# protoactor-go-persistence-mysql

Go package with persistence provider for Proto Actor (Go) based on MySQL.

# Usage

```go
package main

import (
	"database/sql"
	
	"github.com/asynkron/protoactor-go/actor"
	"github.com/asynkron/protoactor-go/persistence"
	"github.com/go-sql-driver/mysql"
	"github.com/ytake/protoactor-go-persistence-mysql/persistencemysql"
)

type Actor struct {
	persistence.Mixin
}

func (a *Actor) Receive(ctx actor.Context) {
	// example
}

func main() {

	conf := &mysql.Config{
		// example config
	}
	system := actor.NewActorSystem()
	db, _ := sql.Open("mysql", conf.FormatDSN())
	provider, _ := persistencemysql.New(3, persistencemysql.NewTable(), db, system.Logger())

	props := actor.PropsFromProducer(func() actor.Actor { return &Actor{} },
		actor.WithReceiverMiddleware(persistence.Using(provider)))

	pid, _ := system.Root.SpawnNamed(props, "persistent")
}

```

# Default table schema

use ulid as id(varchar(26)) and json as payload

```sql
CREATE TABLE `journals`
(
    `id`              varchar(26) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL,
    `payload`         json                                                  NOT NULL,
    `sequence_number` bigint                                                 DEFAULT NULL,
    `actor_name`      varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin DEFAULT NULL,
    `created_at`      timestamp                                              DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uidx_id` (`id`),
    UNIQUE KEY `uidx_names` (`actor_name`,`sequence_number`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

CREATE TABLE `snapshots`
(
    `id`              varchar(26) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL,
    `payload`         json                                                  NOT NULL,
    `sequence_number` bigint                                                 DEFAULT NULL,
    `actor_name`      varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin DEFAULT NULL,
    `created_at`      timestamp                                              DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uidx_id` (`id`),
    UNIQUE KEY `uidx_names` (`actor_name`,`sequence_number`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

```

## change table name

use the interface to change the table name.

for journal table and snapshot table.

```go 
// Schemaer is the interface that wraps the basic methods for a schema.
type Schemaer interface {
    // JournalTableName returns the name of the journal table.
    JournalTableName() string
    // SnapshotTableName returns the name of the snapshot table.
    SnapshotTableName() string
    // ID returns the name of the id column.
    ID() string
    // Payload returns the name of the payload column.
    Payload() string
    // ActorName returns the name of the actor name column.
    ActorName() string
    // SequenceNumber returns the name of the sequence number column.
    SequenceNumber() string
    // Created returns the name of the created at column.
    Created() string
    // CreateTable returns the sql statement to create the table.
    CreateTable() []string
}

``` 
