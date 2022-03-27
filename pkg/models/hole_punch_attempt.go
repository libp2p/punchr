// Code generated by SQLBoiler 4.6.0 (https://github.com/volatiletech/sqlboiler). DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package models

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/friendsofgo/errors"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/queries/qmhelper"
	"github.com/volatiletech/strmangle"
)

// HolePunchAttempt is an object representing the database table.
type HolePunchAttempt struct {
	ID                int         `boil:"id" json:"id" toml:"id" yaml:"id"`
	HolePunchResultID int         `boil:"hole_punch_result_id" json:"hole_punch_result_id" toml:"hole_punch_result_id" yaml:"hole_punch_result_id"`
	OpenedAt          time.Time   `boil:"opened_at" json:"opened_at" toml:"opened_at" yaml:"opened_at"`
	StartedAt         time.Time   `boil:"started_at" json:"started_at" toml:"started_at" yaml:"started_at"`
	EndedAt           time.Time   `boil:"ended_at" json:"ended_at" toml:"ended_at" yaml:"ended_at"`
	StartRTT          null.String `boil:"start_rtt" json:"start_rtt,omitempty" toml:"start_rtt" yaml:"start_rtt,omitempty"`
	ElapsedTime       string      `boil:"elapsed_time" json:"elapsed_time" toml:"elapsed_time" yaml:"elapsed_time"`
	Outcome           string      `boil:"outcome" json:"outcome" toml:"outcome" yaml:"outcome"`
	Error             null.String `boil:"error" json:"error,omitempty" toml:"error" yaml:"error,omitempty"`
	DirectDialError   null.String `boil:"direct_dial_error" json:"direct_dial_error,omitempty" toml:"direct_dial_error" yaml:"direct_dial_error,omitempty"`
	UpdatedAt         time.Time   `boil:"updated_at" json:"updated_at" toml:"updated_at" yaml:"updated_at"`
	CreatedAt         time.Time   `boil:"created_at" json:"created_at" toml:"created_at" yaml:"created_at"`

	R *holePunchAttemptR `boil:"-" json:"-" toml:"-" yaml:"-"`
	L holePunchAttemptL  `boil:"-" json:"-" toml:"-" yaml:"-"`
}

var HolePunchAttemptColumns = struct {
	ID                string
	HolePunchResultID string
	OpenedAt          string
	StartedAt         string
	EndedAt           string
	StartRTT          string
	ElapsedTime       string
	Outcome           string
	Error             string
	DirectDialError   string
	UpdatedAt         string
	CreatedAt         string
}{
	ID:                "id",
	HolePunchResultID: "hole_punch_result_id",
	OpenedAt:          "opened_at",
	StartedAt:         "started_at",
	EndedAt:           "ended_at",
	StartRTT:          "start_rtt",
	ElapsedTime:       "elapsed_time",
	Outcome:           "outcome",
	Error:             "error",
	DirectDialError:   "direct_dial_error",
	UpdatedAt:         "updated_at",
	CreatedAt:         "created_at",
}

var HolePunchAttemptTableColumns = struct {
	ID                string
	HolePunchResultID string
	OpenedAt          string
	StartedAt         string
	EndedAt           string
	StartRTT          string
	ElapsedTime       string
	Outcome           string
	Error             string
	DirectDialError   string
	UpdatedAt         string
	CreatedAt         string
}{
	ID:                "hole_punch_attempt.id",
	HolePunchResultID: "hole_punch_attempt.hole_punch_result_id",
	OpenedAt:          "hole_punch_attempt.opened_at",
	StartedAt:         "hole_punch_attempt.started_at",
	EndedAt:           "hole_punch_attempt.ended_at",
	StartRTT:          "hole_punch_attempt.start_rtt",
	ElapsedTime:       "hole_punch_attempt.elapsed_time",
	Outcome:           "hole_punch_attempt.outcome",
	Error:             "hole_punch_attempt.error",
	DirectDialError:   "hole_punch_attempt.direct_dial_error",
	UpdatedAt:         "hole_punch_attempt.updated_at",
	CreatedAt:         "hole_punch_attempt.created_at",
}

// Generated where

type whereHelpernull_String struct{ field string }

func (w whereHelpernull_String) EQ(x null.String) qm.QueryMod {
	return qmhelper.WhereNullEQ(w.field, false, x)
}
func (w whereHelpernull_String) NEQ(x null.String) qm.QueryMod {
	return qmhelper.WhereNullEQ(w.field, true, x)
}
func (w whereHelpernull_String) IsNull() qm.QueryMod    { return qmhelper.WhereIsNull(w.field) }
func (w whereHelpernull_String) IsNotNull() qm.QueryMod { return qmhelper.WhereIsNotNull(w.field) }
func (w whereHelpernull_String) LT(x null.String) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LT, x)
}
func (w whereHelpernull_String) LTE(x null.String) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LTE, x)
}
func (w whereHelpernull_String) GT(x null.String) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GT, x)
}
func (w whereHelpernull_String) GTE(x null.String) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GTE, x)
}

var HolePunchAttemptWhere = struct {
	ID                whereHelperint
	HolePunchResultID whereHelperint
	OpenedAt          whereHelpertime_Time
	StartedAt         whereHelpertime_Time
	EndedAt           whereHelpertime_Time
	StartRTT          whereHelpernull_String
	ElapsedTime       whereHelperstring
	Outcome           whereHelperstring
	Error             whereHelpernull_String
	DirectDialError   whereHelpernull_String
	UpdatedAt         whereHelpertime_Time
	CreatedAt         whereHelpertime_Time
}{
	ID:                whereHelperint{field: "\"hole_punch_attempt\".\"id\""},
	HolePunchResultID: whereHelperint{field: "\"hole_punch_attempt\".\"hole_punch_result_id\""},
	OpenedAt:          whereHelpertime_Time{field: "\"hole_punch_attempt\".\"opened_at\""},
	StartedAt:         whereHelpertime_Time{field: "\"hole_punch_attempt\".\"started_at\""},
	EndedAt:           whereHelpertime_Time{field: "\"hole_punch_attempt\".\"ended_at\""},
	StartRTT:          whereHelpernull_String{field: "\"hole_punch_attempt\".\"start_rtt\""},
	ElapsedTime:       whereHelperstring{field: "\"hole_punch_attempt\".\"elapsed_time\""},
	Outcome:           whereHelperstring{field: "\"hole_punch_attempt\".\"outcome\""},
	Error:             whereHelpernull_String{field: "\"hole_punch_attempt\".\"error\""},
	DirectDialError:   whereHelpernull_String{field: "\"hole_punch_attempt\".\"direct_dial_error\""},
	UpdatedAt:         whereHelpertime_Time{field: "\"hole_punch_attempt\".\"updated_at\""},
	CreatedAt:         whereHelpertime_Time{field: "\"hole_punch_attempt\".\"created_at\""},
}

// HolePunchAttemptRels is where relationship names are stored.
var HolePunchAttemptRels = struct {
	HolePunchResult string
}{
	HolePunchResult: "HolePunchResult",
}

// holePunchAttemptR is where relationships are stored.
type holePunchAttemptR struct {
	HolePunchResult *HolePunchResult `boil:"HolePunchResult" json:"HolePunchResult" toml:"HolePunchResult" yaml:"HolePunchResult"`
}

// NewStruct creates a new relationship struct
func (*holePunchAttemptR) NewStruct() *holePunchAttemptR {
	return &holePunchAttemptR{}
}

// holePunchAttemptL is where Load methods for each relationship are stored.
type holePunchAttemptL struct{}

var (
	holePunchAttemptAllColumns            = []string{"id", "hole_punch_result_id", "opened_at", "started_at", "ended_at", "start_rtt", "elapsed_time", "outcome", "error", "direct_dial_error", "updated_at", "created_at"}
	holePunchAttemptColumnsWithoutDefault = []string{"hole_punch_result_id", "opened_at", "started_at", "ended_at", "start_rtt", "elapsed_time", "outcome", "error", "direct_dial_error", "updated_at", "created_at"}
	holePunchAttemptColumnsWithDefault    = []string{"id"}
	holePunchAttemptPrimaryKeyColumns     = []string{"id"}
)

type (
	// HolePunchAttemptSlice is an alias for a slice of pointers to HolePunchAttempt.
	// This should almost always be used instead of []HolePunchAttempt.
	HolePunchAttemptSlice []*HolePunchAttempt
	// HolePunchAttemptHook is the signature for custom HolePunchAttempt hook methods
	HolePunchAttemptHook func(context.Context, boil.ContextExecutor, *HolePunchAttempt) error

	holePunchAttemptQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	holePunchAttemptType                 = reflect.TypeOf(&HolePunchAttempt{})
	holePunchAttemptMapping              = queries.MakeStructMapping(holePunchAttemptType)
	holePunchAttemptPrimaryKeyMapping, _ = queries.BindMapping(holePunchAttemptType, holePunchAttemptMapping, holePunchAttemptPrimaryKeyColumns)
	holePunchAttemptInsertCacheMut       sync.RWMutex
	holePunchAttemptInsertCache          = make(map[string]insertCache)
	holePunchAttemptUpdateCacheMut       sync.RWMutex
	holePunchAttemptUpdateCache          = make(map[string]updateCache)
	holePunchAttemptUpsertCacheMut       sync.RWMutex
	holePunchAttemptUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var holePunchAttemptBeforeInsertHooks []HolePunchAttemptHook
var holePunchAttemptBeforeUpdateHooks []HolePunchAttemptHook
var holePunchAttemptBeforeDeleteHooks []HolePunchAttemptHook
var holePunchAttemptBeforeUpsertHooks []HolePunchAttemptHook

var holePunchAttemptAfterInsertHooks []HolePunchAttemptHook
var holePunchAttemptAfterSelectHooks []HolePunchAttemptHook
var holePunchAttemptAfterUpdateHooks []HolePunchAttemptHook
var holePunchAttemptAfterDeleteHooks []HolePunchAttemptHook
var holePunchAttemptAfterUpsertHooks []HolePunchAttemptHook

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *HolePunchAttempt) doBeforeInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range holePunchAttemptBeforeInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *HolePunchAttempt) doBeforeUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range holePunchAttemptBeforeUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *HolePunchAttempt) doBeforeDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range holePunchAttemptBeforeDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *HolePunchAttempt) doBeforeUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range holePunchAttemptBeforeUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *HolePunchAttempt) doAfterInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range holePunchAttemptAfterInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterSelectHooks executes all "after Select" hooks.
func (o *HolePunchAttempt) doAfterSelectHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range holePunchAttemptAfterSelectHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *HolePunchAttempt) doAfterUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range holePunchAttemptAfterUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *HolePunchAttempt) doAfterDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range holePunchAttemptAfterDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *HolePunchAttempt) doAfterUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range holePunchAttemptAfterUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddHolePunchAttemptHook registers your hook function for all future operations.
func AddHolePunchAttemptHook(hookPoint boil.HookPoint, holePunchAttemptHook HolePunchAttemptHook) {
	switch hookPoint {
	case boil.BeforeInsertHook:
		holePunchAttemptBeforeInsertHooks = append(holePunchAttemptBeforeInsertHooks, holePunchAttemptHook)
	case boil.BeforeUpdateHook:
		holePunchAttemptBeforeUpdateHooks = append(holePunchAttemptBeforeUpdateHooks, holePunchAttemptHook)
	case boil.BeforeDeleteHook:
		holePunchAttemptBeforeDeleteHooks = append(holePunchAttemptBeforeDeleteHooks, holePunchAttemptHook)
	case boil.BeforeUpsertHook:
		holePunchAttemptBeforeUpsertHooks = append(holePunchAttemptBeforeUpsertHooks, holePunchAttemptHook)
	case boil.AfterInsertHook:
		holePunchAttemptAfterInsertHooks = append(holePunchAttemptAfterInsertHooks, holePunchAttemptHook)
	case boil.AfterSelectHook:
		holePunchAttemptAfterSelectHooks = append(holePunchAttemptAfterSelectHooks, holePunchAttemptHook)
	case boil.AfterUpdateHook:
		holePunchAttemptAfterUpdateHooks = append(holePunchAttemptAfterUpdateHooks, holePunchAttemptHook)
	case boil.AfterDeleteHook:
		holePunchAttemptAfterDeleteHooks = append(holePunchAttemptAfterDeleteHooks, holePunchAttemptHook)
	case boil.AfterUpsertHook:
		holePunchAttemptAfterUpsertHooks = append(holePunchAttemptAfterUpsertHooks, holePunchAttemptHook)
	}
}

// One returns a single holePunchAttempt record from the query.
func (q holePunchAttemptQuery) One(ctx context.Context, exec boil.ContextExecutor) (*HolePunchAttempt, error) {
	o := &HolePunchAttempt{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(ctx, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: failed to execute a one query for hole_punch_attempt")
	}

	if err := o.doAfterSelectHooks(ctx, exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all HolePunchAttempt records from the query.
func (q holePunchAttemptQuery) All(ctx context.Context, exec boil.ContextExecutor) (HolePunchAttemptSlice, error) {
	var o []*HolePunchAttempt

	err := q.Bind(ctx, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "models: failed to assign all query results to HolePunchAttempt slice")
	}

	if len(holePunchAttemptAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(ctx, exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all HolePunchAttempt records in the query.
func (q holePunchAttemptQuery) Count(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to count hole_punch_attempt rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q holePunchAttemptQuery) Exists(ctx context.Context, exec boil.ContextExecutor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "models: failed to check if hole_punch_attempt exists")
	}

	return count > 0, nil
}

// HolePunchResult pointed to by the foreign key.
func (o *HolePunchAttempt) HolePunchResult(mods ...qm.QueryMod) holePunchResultQuery {
	queryMods := []qm.QueryMod{
		qm.Where("\"id\" = ?", o.HolePunchResultID),
	}

	queryMods = append(queryMods, mods...)

	query := HolePunchResults(queryMods...)
	queries.SetFrom(query.Query, "\"hole_punch_results\"")

	return query
}

// LoadHolePunchResult allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (holePunchAttemptL) LoadHolePunchResult(ctx context.Context, e boil.ContextExecutor, singular bool, maybeHolePunchAttempt interface{}, mods queries.Applicator) error {
	var slice []*HolePunchAttempt
	var object *HolePunchAttempt

	if singular {
		object = maybeHolePunchAttempt.(*HolePunchAttempt)
	} else {
		slice = *maybeHolePunchAttempt.(*[]*HolePunchAttempt)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &holePunchAttemptR{}
		}
		args = append(args, object.HolePunchResultID)

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &holePunchAttemptR{}
			}

			for _, a := range args {
				if a == obj.HolePunchResultID {
					continue Outer
				}
			}

			args = append(args, obj.HolePunchResultID)

		}
	}

	if len(args) == 0 {
		return nil
	}

	query := NewQuery(
		qm.From(`hole_punch_results`),
		qm.WhereIn(`hole_punch_results.id in ?`, args...),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.QueryContext(ctx, e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load HolePunchResult")
	}

	var resultSlice []*HolePunchResult
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice HolePunchResult")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results of eager load for hole_punch_results")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for hole_punch_results")
	}

	if len(holePunchAttemptAfterSelectHooks) != 0 {
		for _, obj := range resultSlice {
			if err := obj.doAfterSelectHooks(ctx, e); err != nil {
				return err
			}
		}
	}

	if len(resultSlice) == 0 {
		return nil
	}

	if singular {
		foreign := resultSlice[0]
		object.R.HolePunchResult = foreign
		if foreign.R == nil {
			foreign.R = &holePunchResultR{}
		}
		foreign.R.HolePunchAttempts = append(foreign.R.HolePunchAttempts, object)
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.HolePunchResultID == foreign.ID {
				local.R.HolePunchResult = foreign
				if foreign.R == nil {
					foreign.R = &holePunchResultR{}
				}
				foreign.R.HolePunchAttempts = append(foreign.R.HolePunchAttempts, local)
				break
			}
		}
	}

	return nil
}

// SetHolePunchResult of the holePunchAttempt to the related item.
// Sets o.R.HolePunchResult to related.
// Adds o to related.R.HolePunchAttempts.
func (o *HolePunchAttempt) SetHolePunchResult(ctx context.Context, exec boil.ContextExecutor, insert bool, related *HolePunchResult) error {
	var err error
	if insert {
		if err = related.Insert(ctx, exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"hole_punch_attempt\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"hole_punch_result_id"}),
		strmangle.WhereClause("\"", "\"", 2, holePunchAttemptPrimaryKeyColumns),
	)
	values := []interface{}{related.ID, o.ID}

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, updateQuery)
		fmt.Fprintln(writer, values)
	}
	if _, err = exec.ExecContext(ctx, updateQuery, values...); err != nil {
		return errors.Wrap(err, "failed to update local table")
	}

	o.HolePunchResultID = related.ID
	if o.R == nil {
		o.R = &holePunchAttemptR{
			HolePunchResult: related,
		}
	} else {
		o.R.HolePunchResult = related
	}

	if related.R == nil {
		related.R = &holePunchResultR{
			HolePunchAttempts: HolePunchAttemptSlice{o},
		}
	} else {
		related.R.HolePunchAttempts = append(related.R.HolePunchAttempts, o)
	}

	return nil
}

// HolePunchAttempts retrieves all the records using an executor.
func HolePunchAttempts(mods ...qm.QueryMod) holePunchAttemptQuery {
	mods = append(mods, qm.From("\"hole_punch_attempt\""))
	return holePunchAttemptQuery{NewQuery(mods...)}
}

// FindHolePunchAttempt retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindHolePunchAttempt(ctx context.Context, exec boil.ContextExecutor, iD int, selectCols ...string) (*HolePunchAttempt, error) {
	holePunchAttemptObj := &HolePunchAttempt{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"hole_punch_attempt\" where \"id\"=$1", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(ctx, exec, holePunchAttemptObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: unable to select from hole_punch_attempt")
	}

	if err = holePunchAttemptObj.doAfterSelectHooks(ctx, exec); err != nil {
		return holePunchAttemptObj, err
	}

	return holePunchAttemptObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *HolePunchAttempt) Insert(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) error {
	if o == nil {
		return errors.New("models: no hole_punch_attempt provided for insertion")
	}

	var err error
	if !boil.TimestampsAreSkipped(ctx) {
		currTime := time.Now().In(boil.GetLocation())

		if o.UpdatedAt.IsZero() {
			o.UpdatedAt = currTime
		}
		if o.CreatedAt.IsZero() {
			o.CreatedAt = currTime
		}
	}

	if err := o.doBeforeInsertHooks(ctx, exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(holePunchAttemptColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	holePunchAttemptInsertCacheMut.RLock()
	cache, cached := holePunchAttemptInsertCache[key]
	holePunchAttemptInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			holePunchAttemptAllColumns,
			holePunchAttemptColumnsWithDefault,
			holePunchAttemptColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(holePunchAttemptType, holePunchAttemptMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(holePunchAttemptType, holePunchAttemptMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"hole_punch_attempt\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"hole_punch_attempt\" %sDEFAULT VALUES%s"
		}

		var queryOutput, queryReturning string

		if len(cache.retMapping) != 0 {
			queryReturning = fmt.Sprintf(" RETURNING \"%s\"", strings.Join(returnColumns, "\",\""))
		}

		cache.query = fmt.Sprintf(cache.query, queryOutput, queryReturning)
	}

	value := reflect.Indirect(reflect.ValueOf(o))
	vals := queries.ValuesFromMapping(value, cache.valueMapping)

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, cache.query)
		fmt.Fprintln(writer, vals)
	}

	if len(cache.retMapping) != 0 {
		err = exec.QueryRowContext(ctx, cache.query, vals...).Scan(queries.PtrsFromMapping(value, cache.retMapping)...)
	} else {
		_, err = exec.ExecContext(ctx, cache.query, vals...)
	}

	if err != nil {
		return errors.Wrap(err, "models: unable to insert into hole_punch_attempt")
	}

	if !cached {
		holePunchAttemptInsertCacheMut.Lock()
		holePunchAttemptInsertCache[key] = cache
		holePunchAttemptInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(ctx, exec)
}

// Update uses an executor to update the HolePunchAttempt.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *HolePunchAttempt) Update(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) (int64, error) {
	if !boil.TimestampsAreSkipped(ctx) {
		currTime := time.Now().In(boil.GetLocation())

		o.UpdatedAt = currTime
	}

	var err error
	if err = o.doBeforeUpdateHooks(ctx, exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	holePunchAttemptUpdateCacheMut.RLock()
	cache, cached := holePunchAttemptUpdateCache[key]
	holePunchAttemptUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			holePunchAttemptAllColumns,
			holePunchAttemptPrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("models: unable to update hole_punch_attempt, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"hole_punch_attempt\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, holePunchAttemptPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(holePunchAttemptType, holePunchAttemptMapping, append(wl, holePunchAttemptPrimaryKeyColumns...))
		if err != nil {
			return 0, err
		}
	}

	values := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), cache.valueMapping)

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, cache.query)
		fmt.Fprintln(writer, values)
	}
	var result sql.Result
	result, err = exec.ExecContext(ctx, cache.query, values...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update hole_punch_attempt row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by update for hole_punch_attempt")
	}

	if !cached {
		holePunchAttemptUpdateCacheMut.Lock()
		holePunchAttemptUpdateCache[key] = cache
		holePunchAttemptUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(ctx, exec)
}

// UpdateAll updates all rows with the specified column values.
func (q holePunchAttemptQuery) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all for hole_punch_attempt")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected for hole_punch_attempt")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o HolePunchAttemptSlice) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	ln := int64(len(o))
	if ln == 0 {
		return 0, nil
	}

	if len(cols) == 0 {
		return 0, errors.New("models: update all requires at least one column argument")
	}

	colNames := make([]string, len(cols))
	args := make([]interface{}, len(cols))

	i := 0
	for name, value := range cols {
		colNames[i] = name
		args[i] = value
		i++
	}

	// Append all of the primary key values for each column
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), holePunchAttemptPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"hole_punch_attempt\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, holePunchAttemptPrimaryKeyColumns, len(o)))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all in holePunchAttempt slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected all in update all holePunchAttempt")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *HolePunchAttempt) Upsert(ctx context.Context, exec boil.ContextExecutor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("models: no hole_punch_attempt provided for upsert")
	}
	if !boil.TimestampsAreSkipped(ctx) {
		currTime := time.Now().In(boil.GetLocation())

		o.UpdatedAt = currTime
		if o.CreatedAt.IsZero() {
			o.CreatedAt = currTime
		}
	}

	if err := o.doBeforeUpsertHooks(ctx, exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(holePunchAttemptColumnsWithDefault, o)

	// Build cache key in-line uglily - mysql vs psql problems
	buf := strmangle.GetBuffer()
	if updateOnConflict {
		buf.WriteByte('t')
	} else {
		buf.WriteByte('f')
	}
	buf.WriteByte('.')
	for _, c := range conflictColumns {
		buf.WriteString(c)
	}
	buf.WriteByte('.')
	buf.WriteString(strconv.Itoa(updateColumns.Kind))
	for _, c := range updateColumns.Cols {
		buf.WriteString(c)
	}
	buf.WriteByte('.')
	buf.WriteString(strconv.Itoa(insertColumns.Kind))
	for _, c := range insertColumns.Cols {
		buf.WriteString(c)
	}
	buf.WriteByte('.')
	for _, c := range nzDefaults {
		buf.WriteString(c)
	}
	key := buf.String()
	strmangle.PutBuffer(buf)

	holePunchAttemptUpsertCacheMut.RLock()
	cache, cached := holePunchAttemptUpsertCache[key]
	holePunchAttemptUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			holePunchAttemptAllColumns,
			holePunchAttemptColumnsWithDefault,
			holePunchAttemptColumnsWithoutDefault,
			nzDefaults,
		)
		update := updateColumns.UpdateColumnSet(
			holePunchAttemptAllColumns,
			holePunchAttemptPrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("models: unable to upsert hole_punch_attempt, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(holePunchAttemptPrimaryKeyColumns))
			copy(conflict, holePunchAttemptPrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"hole_punch_attempt\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(holePunchAttemptType, holePunchAttemptMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(holePunchAttemptType, holePunchAttemptMapping, ret)
			if err != nil {
				return err
			}
		}
	}

	value := reflect.Indirect(reflect.ValueOf(o))
	vals := queries.ValuesFromMapping(value, cache.valueMapping)
	var returns []interface{}
	if len(cache.retMapping) != 0 {
		returns = queries.PtrsFromMapping(value, cache.retMapping)
	}

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, cache.query)
		fmt.Fprintln(writer, vals)
	}
	if len(cache.retMapping) != 0 {
		err = exec.QueryRowContext(ctx, cache.query, vals...).Scan(returns...)
		if err == sql.ErrNoRows {
			err = nil // Postgres doesn't return anything when there's no update
		}
	} else {
		_, err = exec.ExecContext(ctx, cache.query, vals...)
	}
	if err != nil {
		return errors.Wrap(err, "models: unable to upsert hole_punch_attempt")
	}

	if !cached {
		holePunchAttemptUpsertCacheMut.Lock()
		holePunchAttemptUpsertCache[key] = cache
		holePunchAttemptUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(ctx, exec)
}

// Delete deletes a single HolePunchAttempt record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *HolePunchAttempt) Delete(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if o == nil {
		return 0, errors.New("models: no HolePunchAttempt provided for delete")
	}

	if err := o.doBeforeDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), holePunchAttemptPrimaryKeyMapping)
	sql := "DELETE FROM \"hole_punch_attempt\" WHERE \"id\"=$1"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete from hole_punch_attempt")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by delete for hole_punch_attempt")
	}

	if err := o.doAfterDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q holePunchAttemptQuery) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("models: no holePunchAttemptQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from hole_punch_attempt")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for hole_punch_attempt")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o HolePunchAttemptSlice) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(holePunchAttemptBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(ctx, exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), holePunchAttemptPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"hole_punch_attempt\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, holePunchAttemptPrimaryKeyColumns, len(o))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from holePunchAttempt slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for hole_punch_attempt")
	}

	if len(holePunchAttemptAfterDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterDeleteHooks(ctx, exec); err != nil {
				return 0, err
			}
		}
	}

	return rowsAff, nil
}

// Reload refetches the object from the database
// using the primary keys with an executor.
func (o *HolePunchAttempt) Reload(ctx context.Context, exec boil.ContextExecutor) error {
	ret, err := FindHolePunchAttempt(ctx, exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *HolePunchAttemptSlice) ReloadAll(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := HolePunchAttemptSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), holePunchAttemptPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"hole_punch_attempt\".* FROM \"hole_punch_attempt\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, holePunchAttemptPrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(ctx, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "models: unable to reload all in HolePunchAttemptSlice")
	}

	*o = slice

	return nil
}

// HolePunchAttemptExists checks if the HolePunchAttempt row exists.
func HolePunchAttemptExists(ctx context.Context, exec boil.ContextExecutor, iD int) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"hole_punch_attempt\" where \"id\"=$1 limit 1)"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, iD)
	}
	row := exec.QueryRowContext(ctx, sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "models: unable to check if hole_punch_attempt exists")
	}

	return exists, nil
}