package db

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"

	"github.com/dennis-tra/punchr/pkg/maxmind"
	"github.com/dennis-tra/punchr/pkg/models"
	"github.com/dennis-tra/punchr/pkg/util"
)

type Client struct {
	*sql.DB

	MMClient *maxmind.Client
}

func NewClient(c *cli.Context) (*Client, error) {
	log.WithFields(log.Fields{
		"host":     c.String("db-host"),
		"port":     c.String("db-port"),
		"name":     c.String("db-name"),
		"user":     c.String("db-user"),
		"password": "****",
		"sslmode":  c.String("db-sslmode"),
	}).Infoln("Connecting to postgres database...")

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

	// Set global database handler
	boil.SetDB(dbh)

	// Create maxmind client to derive geo information
	mmClient, err := maxmind.NewClient()
	if err != nil {
		return nil, errors.Wrap(err, "new maxmind client")
	}

	// Apply migrations
	driver, err := postgres.WithInstance(dbh, &postgres.Config{})
	if err != nil {
		return nil, err
	}

	m, err := migrate.NewWithDatabaseInstance("file://./migrations", "punchr", driver)
	if err != nil {
		return nil, err
	}

	if err = m.Up(); err != nil && err != migrate.ErrNoChange {
		return nil, err
	}

	return &Client{DB: dbh, MMClient: mmClient}, nil
}

// DeferRollback calls rollback on the given transaction and logs the potential error.
func DeferRollback(txn *sql.Tx) {
	if err := txn.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
		log.WithError(err).Warnln("An error occurred when rolling back transaction")
	}
}

func (c *Client) UpsertPeer(ctx context.Context, exec boil.ContextExecutor, pid peer.ID, agentVersion *string, protocols []string) (*models.Peer, error) {
	dbPeer := &models.Peer{
		MultiHash:     pid.String(),
		AgentVersion:  null.StringFromPtr(agentVersion),
		Protocols:     protocols,
		SupportsDcutr: util.SupportDCUtR(protocols),
	}

	err := dbPeer.Upsert(
		ctx,
		exec,
		true,
		[]string{models.PeerColumns.MultiHash},
		boil.Whitelist( // which columns to update on conflict
			models.PeerColumns.UpdatedAt,
			models.PeerColumns.AgentVersion,
			models.PeerColumns.Protocols,
			models.PeerColumns.SupportsDcutr,
		),
		boil.Infer(),
	)
	return dbPeer, err
}

func (c *Client) GetAuthorization(ctx context.Context, exec boil.ContextExecutor, apiKey string) (*models.Authorization, error) {
	return models.Authorizations(models.AuthorizationWhere.APIKey.EQ(apiKey)).One(ctx, exec)
}

func (c *Client) UpsertMultiAddresses(ctx context.Context, exec boil.ContextExecutor, maddrs []ma.Multiaddr) (models.MultiAddressSlice, error) {
	// Sort maddrs to avoid dead-locks when inserting
	sort.SliceStable(maddrs, func(i, j int) bool {
		return maddrs[i].String() < maddrs[j].String()
	})

	dbMaddrs := make([]*models.MultiAddress, len(maddrs))
	for i, maddr := range maddrs {
		dbIPAddresses, country, continent, asn, err := c.UpsertIPAddresses(ctx, exec, maddr)
		if err != nil {
			return nil, errors.Wrap(err, "save ip addresses")
		}

		dbMaddr := &models.MultiAddress{
			Maddr:          maddr.String(),
			Country:        null.StringFromPtr(country),
			Continent:      null.StringFromPtr(continent),
			Asn:            null.IntFromPtr(asn),
			IsPublic:       manet.IsPublicAddr(maddr),
			IsRelay:        util.IsRelayedMaddr(maddr),
			IPAddressCount: len(dbIPAddresses),
		}
		if err = dbMaddr.Upsert(ctx, exec, true, []string{models.MultiAddressColumns.Maddr}, boil.Whitelist(models.MultiAddressColumns.UpdatedAt), boil.Infer()); err != nil {
			return nil, errors.Wrap(err, "upsert multi address")
		}

		if err := dbMaddr.SetIPAddresses(ctx, exec, false); err != nil {
			return nil, errors.Wrap(err, "removing ip addresses to multi addresses association")
		}

		if err = dbMaddr.AddIPAddresses(ctx, exec, false, dbIPAddresses...); err != nil {
			return nil, errors.Wrap(err, "adding ip addresses to multi address")
		}

		dbMaddrs[i] = dbMaddr
	}

	return dbMaddrs, nil
}

func (c *Client) UpsertMultiAddressesSet(ctx context.Context, exec boil.ContextExecutor, dbMaddrs models.MultiAddressSlice) (int, error) {
	ids := make([]string, len(dbMaddrs))
	for i, dbMaddr := range dbMaddrs {
		ids[i] = strconv.FormatInt(dbMaddr.ID, 10)
	}

	rows, err := queries.Raw("SELECT upsert_multi_addresses_sets('{"+strings.Join(ids, ",")+"}')").QueryContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "upsert multi addresses set")
	}

	if !rows.Next() {
		return 0, fmt.Errorf("no multi address set created")
	}

	var maddrSetID int
	if err = rows.Scan(&maddrSetID); err != nil {
		return 0, errors.Wrap(err, "scan multi address set id")
	}

	return maddrSetID, errors.Wrap(rows.Close(), "close multi address set query rows")
}

func (c *Client) UpsertIPAddresses(ctx context.Context, exec boil.ContextExecutor, maddr ma.Multiaddr) (models.IPAddressSlice, *string, *string, *int, error) {
	var (
		countries     []string
		continents    []string
		asns          []int
		dbIPAddresses []*models.IPAddress
	)

	addrInfos, err := c.MMClient.MaddrInfo(ctx, maddr)
	if err != nil {
		// That's not an error that should be returned
		log.WithError(err).Warnln("Could not get multi address information")
	}

	for ipAddress, addrInfo := range addrInfos {
		if addrInfo.Country != "" {
			countries = append(countries, addrInfo.Country)
		}

		if addrInfo.Continent != "" {
			continents = append(continents, addrInfo.Continent)
		}

		if addrInfo.ASN != 0 {
			asns = append(asns, int(addrInfo.ASN))
		}

		dbIPAddress := &models.IPAddress{
			Address:   ipAddress,
			Country:   null.NewString(addrInfo.Country, addrInfo.Country != ""),
			Continent: null.NewString(addrInfo.Continent, addrInfo.Continent != ""),
			Asn:       null.NewInt(int(addrInfo.ASN), addrInfo.ASN != 0),
			IsPublic:  manet.IsPublicAddr(maddr),
		}

		err = dbIPAddress.Upsert(
			ctx,
			exec,
			true,
			[]string{models.IPAddressColumns.Address},
			boil.Whitelist(models.IPAddressColumns.UpdatedAt),
			boil.Infer(),
		)
		if err != nil {
			return nil, nil, nil, nil, errors.Wrap(err, "upsert ip address")
		}

		dbIPAddresses = append(dbIPAddresses, dbIPAddress)
	}

	return dbIPAddresses, util.UniqueStr(countries), util.UniqueStr(continents), util.UniqueInt(asns), nil
}

// MapNetDirection maps the connection direction the corresponding database types.
func MapNetDirection(conn network.Conn) string {
	switch conn.Stat().Direction {
	case network.DirInbound:
		return models.ConnectionDirectionINBOUND
	case network.DirOutbound:
		return models.ConnectionDirectionOUTBOUND
	default:
		return models.ConnectionDirectionUNKNOWN
	}
}
