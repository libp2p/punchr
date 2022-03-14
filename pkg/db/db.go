package db

import (
	"context"
	"database/sql"
	"fmt"
	"sort"

	"github.com/dennis-tra/punchr/pkg/models"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/types"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type Client struct {
	*sql.DB
}

func NewClient(c *cli.Context) (*Client, error) {
	// Open database handle
	srcName := fmt.Sprintf(
		"host=%s port=%s dbname=%s user=%s password=%s sslmode=%s",
		c.String("db-host"),
		c.String("db-port"),
		c.String("db-name"),
		c.String("db-user"),
		c.String("db-password"),
		c.String("db-sslmode"),
	)
	dbh, err := sql.Open("postgres", srcName)
	if err != nil {
		return nil, errors.Wrap(err, "opening database")
	}

	// Ping database to verify connection.
	if err = dbh.Ping(); err != nil {
		return nil, errors.Wrap(err, "pinging database")
	}

	driver, err := postgres.WithInstance(dbh, &postgres.Config{})
	if err != nil {
		return nil, err
	}

	m, err := migrate.NewWithDatabaseInstance("file://./migrations", "optprov", driver)
	if err != nil {
		return nil, err
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return nil, err
	}

	boil.SetDB(dbh)

	return &Client{DB: dbh}, nil
}

func (c *Client) UpsertPeer(ctx context.Context, exec boil.ContextExecutor, pid peer.ID, av string, protocols []string) (*models.Peer, error) {
	sort.Strings(protocols)

	dbPeer := &models.Peer{
		MultiHash:    pid.String(),
		AgentVersion: null.NewString(av, av != ""),
		Protocols:    types.StringArray(protocols),
	}
	var prots interface{} = types.StringArray(protocols)
	if len(protocols) == 0 {
		prots = nil
	}

	rows, err := queries.Raw("SELECT upsert_peer($1, $2, $3)", dbPeer.MultiHash, dbPeer.AgentVersion.Ptr(), prots).QueryContext(ctx, exec)
	if err != nil {
		return nil, err
	}

	if !rows.Next() {
		return nil, rows.Err()
	}

	if err = rows.Scan(&dbPeer.ID); err != nil {
		return nil, err
	}

	return dbPeer, rows.Close()
}
