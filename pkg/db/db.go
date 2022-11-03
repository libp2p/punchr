package db

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"github.com/dennis-tra/punchr/pkg/udger"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/dennis-tra/punchr/pkg/maxmind"
	"github.com/dennis-tra/punchr/pkg/models"
	"github.com/dennis-tra/punchr/pkg/util"
)

//go:embed migrations
var migrations embed.FS

type Client struct {
	*sql.DB

	MMClient *maxmind.Client

	UdgerClient *udger.Client
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

	// Create udger client
	uclient, err := udger.NewClient(c.String("udger-db"))
	if err != nil {
		return nil, errors.Wrap(err, "new udger client")
	}

	client := &Client{DB: dbh, MMClient: mmClient, UdgerClient: uclient}
	client.applyMigrations(c, dbh)

	return client, nil
}

func (c *Client) applyMigrations(ctx *cli.Context, dbh *sql.DB) {
	tmpDir, err := os.MkdirTemp("", "punchr-"+ctx.App.Version)
	if err != nil {
		log.WithError(err).WithField("pattern", "punchr-"+ctx.App.Version).Warnln("Could not create tmp directory for migrations")
		return
	}
	defer func() {
		if err = os.RemoveAll(tmpDir); err != nil {
			log.WithError(err).WithField("tmpDir", tmpDir).Warnln("Could not clean up tmp directory")
		}
	}()
	log.WithField("dir", tmpDir).Debugln("Created temporary directory")

	err = fs.WalkDir(migrations, ".", func(path string, d fs.DirEntry, err error) error {
		join := filepath.Join(tmpDir, path)
		if d.IsDir() {
			return os.MkdirAll(join, 0o755)
		}

		data, err := migrations.ReadFile(path)
		if err != nil {
			return errors.Wrap(err, "read file")
		}

		return os.WriteFile(join, data, 0o644)
	})
	if err != nil {
		log.WithError(err).Warnln("Could not create migrations files")
		return
	}

	// Apply migrations
	driver, err := postgres.WithInstance(dbh, &postgres.Config{})
	if err != nil {
		log.WithError(err).Warnln("Could not create driver instance")
		return
	}

	m, err := migrate.NewWithDatabaseInstance("file://"+filepath.Join(tmpDir, "migrations"), ctx.String("db-name"), driver)
	if err != nil {
		log.WithError(err).Warnln("Could not create migrate instance")
		return
	}

	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.WithError(err).Warnln("Couldn't apply migrations")
		return
	}
}

// DeferRollback calls rollback on the given transaction and logs the potential error.
func DeferRollback(txn *sql.Tx) {
	if err := txn.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
		log.WithError(err).Warnln("An error occurred when rolling back transaction")
	}
}

func (c *Client) UpsertPeer(ctx context.Context, exec boil.ContextExecutor, pid peer.ID, agentVersion *string, protocols []string) (*models.Peer, error) {
	dbPeer := &models.Peer{
		MultiHash:    pid.String(),
		AgentVersion: null.StringFromPtr(agentVersion),
		Protocols:    protocols,
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

		dbMaddr, err := c.UpsertMultiAddress(ctx, exec, maddr)
		if err != nil {
			return nil, errors.Wrap(err, "upsert multi address")
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

func (c *Client) UpsertIPAddresses(ctx context.Context, exec boil.ContextExecutor, addrInfos map[string]*maxmind.AddrInfo) (models.IPAddressSlice, *string, *string, *int, error) {
	var (
		countries     []string
		continents    []string
		asns          []int
		dbIPAddresses []*models.IPAddress
	)

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

		dc, err := c.UdgerClient.Datacenter(ipAddress)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			log.WithError(err).WithField("addr", ipAddress).Warnln("could not extract data center")
		}

		dbIPAddress := &models.IPAddress{
			Address:   ipAddress,
			Country:   null.NewString(addrInfo.Country, addrInfo.Country != ""),
			Continent: null.NewString(addrInfo.Continent, addrInfo.Continent != ""),
			Asn:       null.NewInt(int(addrInfo.ASN), addrInfo.ASN != 0),
			IsCloud:   null.NewInt(dc, dc != 0),
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

	return dbIPAddresses, util.Unique(countries), util.Unique(continents), util.Unique(asns), nil
}

func (c *Client) UpsertMultiAddress(ctx context.Context, exec boil.ContextExecutor, maddr ma.Multiaddr) (*models.MultiAddress, error) {
	addrInfos, err := c.MMClient.MaddrInfo(ctx, maddr)
	if err != nil {
		// That's not an error that should be returned
		log.WithError(err).Warnln("Could not get multi address information")
	}

	dbMaddr := &models.MultiAddress{
		Maddr:    maddr.String(),
		IsRelay:  null.BoolFrom(util.IsRelayedMaddr(maddr)),
		IsPublic: null.BoolFrom(manet.IsPublicAddr(maddr)),
	}

	if len(addrInfos) == 1 {
		var addrInfo *maxmind.AddrInfo
		var ipAddress string
		for ipAddress, addrInfo = range addrInfos {
			break
		}

		dc, err := c.UdgerClient.Datacenter(ipAddress)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			log.WithError(err).WithField("addr", ipAddress).Warnln("could not extract data center")
		}

		dbMaddr.Asn = null.NewInt(int(addrInfo.ASN), addrInfo.ASN != 0)
		dbMaddr.IsCloud = null.NewInt(dc, dc != 0)
		dbMaddr.Addr = null.NewString(ipAddress, ipAddress != "")
		dbMaddr.Country = null.NewString(addrInfo.Country, addrInfo.Country != "")
		dbMaddr.Continent = null.NewString(addrInfo.Continent, addrInfo.Continent != "")
	} else if len(addrInfos) > 1 {
		dbMaddr.HasManyAddrs = null.NewBool(true, true)
		dbipAddresses, country, continent, asn, err := c.UpsertIPAddresses(ctx, exec, addrInfos)
		if err != nil {
			return nil, errors.Wrap(err, "upsert ip addresses")
		}
		dbMaddr.Country = null.StringFromPtr(country)
		dbMaddr.Continent = null.StringFromPtr(continent)
		dbMaddr.Asn = null.IntFromPtr(asn)

		for _, ipAddress := range dbipAddresses {
			if err := ipAddress.SetMultiAddress(ctx, exec, false, dbMaddr); err != nil {
				return nil, errors.Wrap(err, "assign multi address to ip address")
			}
		}
	}

	return dbMaddr, dbMaddr.Upsert(ctx, exec, true, []string{models.MultiAddressColumns.Maddr}, boil.Whitelist(models.MultiAddressColumns.UpdatedAt), boil.Infer())
}
