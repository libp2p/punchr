// Code generated by SQLBoiler 4.13.0 (https://github.com/volatiletech/sqlboiler). DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package models

import (
	"bytes"
	"context"
	"reflect"
	"testing"

	"github.com/volatiletech/randomize"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/strmangle"
)

var (
	// Relationships sometimes use the reflection helper queries.Equal/queries.Assign
	// so force a package dependency in case they don't.
	_ = queries.Equal
)

func testConnectionEvents(t *testing.T) {
	t.Parallel()

	query := ConnectionEvents()

	if query.Query == nil {
		t.Error("expected a query, got nothing")
	}
}

func testConnectionEventsDelete(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &ConnectionEvent{}
	if err = randomize.Struct(seed, o, connectionEventDBTypes, true, connectionEventColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize ConnectionEvent struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	if rowsAff, err := o.Delete(ctx, tx); err != nil {
		t.Error(err)
	} else if rowsAff != 1 {
		t.Error("should only have deleted one row, but affected:", rowsAff)
	}

	count, err := ConnectionEvents().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 0 {
		t.Error("want zero records, got:", count)
	}
}

func testConnectionEventsQueryDeleteAll(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &ConnectionEvent{}
	if err = randomize.Struct(seed, o, connectionEventDBTypes, true, connectionEventColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize ConnectionEvent struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	if rowsAff, err := ConnectionEvents().DeleteAll(ctx, tx); err != nil {
		t.Error(err)
	} else if rowsAff != 1 {
		t.Error("should only have deleted one row, but affected:", rowsAff)
	}

	count, err := ConnectionEvents().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 0 {
		t.Error("want zero records, got:", count)
	}
}

func testConnectionEventsSliceDeleteAll(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &ConnectionEvent{}
	if err = randomize.Struct(seed, o, connectionEventDBTypes, true, connectionEventColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize ConnectionEvent struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	slice := ConnectionEventSlice{o}

	if rowsAff, err := slice.DeleteAll(ctx, tx); err != nil {
		t.Error(err)
	} else if rowsAff != 1 {
		t.Error("should only have deleted one row, but affected:", rowsAff)
	}

	count, err := ConnectionEvents().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 0 {
		t.Error("want zero records, got:", count)
	}
}

func testConnectionEventsExists(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &ConnectionEvent{}
	if err = randomize.Struct(seed, o, connectionEventDBTypes, true, connectionEventColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize ConnectionEvent struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	e, err := ConnectionEventExists(ctx, tx, o.ID)
	if err != nil {
		t.Errorf("Unable to check if ConnectionEvent exists: %s", err)
	}
	if !e {
		t.Errorf("Expected ConnectionEventExists to return true, but got false.")
	}
}

func testConnectionEventsFind(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &ConnectionEvent{}
	if err = randomize.Struct(seed, o, connectionEventDBTypes, true, connectionEventColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize ConnectionEvent struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	connectionEventFound, err := FindConnectionEvent(ctx, tx, o.ID)
	if err != nil {
		t.Error(err)
	}

	if connectionEventFound == nil {
		t.Error("want a record, got nil")
	}
}

func testConnectionEventsBind(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &ConnectionEvent{}
	if err = randomize.Struct(seed, o, connectionEventDBTypes, true, connectionEventColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize ConnectionEvent struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	if err = ConnectionEvents().Bind(ctx, tx, o); err != nil {
		t.Error(err)
	}
}

func testConnectionEventsOne(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &ConnectionEvent{}
	if err = randomize.Struct(seed, o, connectionEventDBTypes, true, connectionEventColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize ConnectionEvent struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	if x, err := ConnectionEvents().One(ctx, tx); err != nil {
		t.Error(err)
	} else if x == nil {
		t.Error("expected to get a non nil record")
	}
}

func testConnectionEventsAll(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	connectionEventOne := &ConnectionEvent{}
	connectionEventTwo := &ConnectionEvent{}
	if err = randomize.Struct(seed, connectionEventOne, connectionEventDBTypes, false, connectionEventColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize ConnectionEvent struct: %s", err)
	}
	if err = randomize.Struct(seed, connectionEventTwo, connectionEventDBTypes, false, connectionEventColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize ConnectionEvent struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = connectionEventOne.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}
	if err = connectionEventTwo.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	slice, err := ConnectionEvents().All(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if len(slice) != 2 {
		t.Error("want 2 records, got:", len(slice))
	}
}

func testConnectionEventsCount(t *testing.T) {
	t.Parallel()

	var err error
	seed := randomize.NewSeed()
	connectionEventOne := &ConnectionEvent{}
	connectionEventTwo := &ConnectionEvent{}
	if err = randomize.Struct(seed, connectionEventOne, connectionEventDBTypes, false, connectionEventColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize ConnectionEvent struct: %s", err)
	}
	if err = randomize.Struct(seed, connectionEventTwo, connectionEventDBTypes, false, connectionEventColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize ConnectionEvent struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = connectionEventOne.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}
	if err = connectionEventTwo.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	count, err := ConnectionEvents().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 2 {
		t.Error("want 2 records, got:", count)
	}
}

func connectionEventBeforeInsertHook(ctx context.Context, e boil.ContextExecutor, o *ConnectionEvent) error {
	*o = ConnectionEvent{}
	return nil
}

func connectionEventAfterInsertHook(ctx context.Context, e boil.ContextExecutor, o *ConnectionEvent) error {
	*o = ConnectionEvent{}
	return nil
}

func connectionEventAfterSelectHook(ctx context.Context, e boil.ContextExecutor, o *ConnectionEvent) error {
	*o = ConnectionEvent{}
	return nil
}

func connectionEventBeforeUpdateHook(ctx context.Context, e boil.ContextExecutor, o *ConnectionEvent) error {
	*o = ConnectionEvent{}
	return nil
}

func connectionEventAfterUpdateHook(ctx context.Context, e boil.ContextExecutor, o *ConnectionEvent) error {
	*o = ConnectionEvent{}
	return nil
}

func connectionEventBeforeDeleteHook(ctx context.Context, e boil.ContextExecutor, o *ConnectionEvent) error {
	*o = ConnectionEvent{}
	return nil
}

func connectionEventAfterDeleteHook(ctx context.Context, e boil.ContextExecutor, o *ConnectionEvent) error {
	*o = ConnectionEvent{}
	return nil
}

func connectionEventBeforeUpsertHook(ctx context.Context, e boil.ContextExecutor, o *ConnectionEvent) error {
	*o = ConnectionEvent{}
	return nil
}

func connectionEventAfterUpsertHook(ctx context.Context, e boil.ContextExecutor, o *ConnectionEvent) error {
	*o = ConnectionEvent{}
	return nil
}

func testConnectionEventsHooks(t *testing.T) {
	t.Parallel()

	var err error

	ctx := context.Background()
	empty := &ConnectionEvent{}
	o := &ConnectionEvent{}

	seed := randomize.NewSeed()
	if err = randomize.Struct(seed, o, connectionEventDBTypes, false); err != nil {
		t.Errorf("Unable to randomize ConnectionEvent object: %s", err)
	}

	AddConnectionEventHook(boil.BeforeInsertHook, connectionEventBeforeInsertHook)
	if err = o.doBeforeInsertHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doBeforeInsertHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected BeforeInsertHook function to empty object, but got: %#v", o)
	}
	connectionEventBeforeInsertHooks = []ConnectionEventHook{}

	AddConnectionEventHook(boil.AfterInsertHook, connectionEventAfterInsertHook)
	if err = o.doAfterInsertHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doAfterInsertHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected AfterInsertHook function to empty object, but got: %#v", o)
	}
	connectionEventAfterInsertHooks = []ConnectionEventHook{}

	AddConnectionEventHook(boil.AfterSelectHook, connectionEventAfterSelectHook)
	if err = o.doAfterSelectHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doAfterSelectHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected AfterSelectHook function to empty object, but got: %#v", o)
	}
	connectionEventAfterSelectHooks = []ConnectionEventHook{}

	AddConnectionEventHook(boil.BeforeUpdateHook, connectionEventBeforeUpdateHook)
	if err = o.doBeforeUpdateHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doBeforeUpdateHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected BeforeUpdateHook function to empty object, but got: %#v", o)
	}
	connectionEventBeforeUpdateHooks = []ConnectionEventHook{}

	AddConnectionEventHook(boil.AfterUpdateHook, connectionEventAfterUpdateHook)
	if err = o.doAfterUpdateHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doAfterUpdateHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected AfterUpdateHook function to empty object, but got: %#v", o)
	}
	connectionEventAfterUpdateHooks = []ConnectionEventHook{}

	AddConnectionEventHook(boil.BeforeDeleteHook, connectionEventBeforeDeleteHook)
	if err = o.doBeforeDeleteHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doBeforeDeleteHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected BeforeDeleteHook function to empty object, but got: %#v", o)
	}
	connectionEventBeforeDeleteHooks = []ConnectionEventHook{}

	AddConnectionEventHook(boil.AfterDeleteHook, connectionEventAfterDeleteHook)
	if err = o.doAfterDeleteHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doAfterDeleteHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected AfterDeleteHook function to empty object, but got: %#v", o)
	}
	connectionEventAfterDeleteHooks = []ConnectionEventHook{}

	AddConnectionEventHook(boil.BeforeUpsertHook, connectionEventBeforeUpsertHook)
	if err = o.doBeforeUpsertHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doBeforeUpsertHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected BeforeUpsertHook function to empty object, but got: %#v", o)
	}
	connectionEventBeforeUpsertHooks = []ConnectionEventHook{}

	AddConnectionEventHook(boil.AfterUpsertHook, connectionEventAfterUpsertHook)
	if err = o.doAfterUpsertHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doAfterUpsertHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected AfterUpsertHook function to empty object, but got: %#v", o)
	}
	connectionEventAfterUpsertHooks = []ConnectionEventHook{}
}

func testConnectionEventsInsert(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &ConnectionEvent{}
	if err = randomize.Struct(seed, o, connectionEventDBTypes, true, connectionEventColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize ConnectionEvent struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	count, err := ConnectionEvents().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 1 {
		t.Error("want one record, got:", count)
	}
}

func testConnectionEventsInsertWhitelist(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &ConnectionEvent{}
	if err = randomize.Struct(seed, o, connectionEventDBTypes, true); err != nil {
		t.Errorf("Unable to randomize ConnectionEvent struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Whitelist(connectionEventColumnsWithoutDefault...)); err != nil {
		t.Error(err)
	}

	count, err := ConnectionEvents().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 1 {
		t.Error("want one record, got:", count)
	}
}

func testConnectionEventToManyMultiAddresses(t *testing.T) {
	var err error
	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()

	var a ConnectionEvent
	var b, c MultiAddress

	seed := randomize.NewSeed()
	if err = randomize.Struct(seed, &a, connectionEventDBTypes, true, connectionEventColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize ConnectionEvent struct: %s", err)
	}

	if err := a.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Fatal(err)
	}

	if err = randomize.Struct(seed, &b, multiAddressDBTypes, false, multiAddressColumnsWithDefault...); err != nil {
		t.Fatal(err)
	}
	if err = randomize.Struct(seed, &c, multiAddressDBTypes, false, multiAddressColumnsWithDefault...); err != nil {
		t.Fatal(err)
	}

	if err = b.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Fatal(err)
	}
	if err = c.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Fatal(err)
	}

	_, err = tx.Exec("insert into \"connection_events_x_multi_addresses\" (\"connection_event_id\", \"multi_address_id\") values ($1, $2)", a.ID, b.ID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = tx.Exec("insert into \"connection_events_x_multi_addresses\" (\"connection_event_id\", \"multi_address_id\") values ($1, $2)", a.ID, c.ID)
	if err != nil {
		t.Fatal(err)
	}

	check, err := a.MultiAddresses().All(ctx, tx)
	if err != nil {
		t.Fatal(err)
	}

	bFound, cFound := false, false
	for _, v := range check {
		if v.ID == b.ID {
			bFound = true
		}
		if v.ID == c.ID {
			cFound = true
		}
	}

	if !bFound {
		t.Error("expected to find b")
	}
	if !cFound {
		t.Error("expected to find c")
	}

	slice := ConnectionEventSlice{&a}
	if err = a.L.LoadMultiAddresses(ctx, tx, false, (*[]*ConnectionEvent)(&slice), nil); err != nil {
		t.Fatal(err)
	}
	if got := len(a.R.MultiAddresses); got != 2 {
		t.Error("number of eager loaded records wrong, got:", got)
	}

	a.R.MultiAddresses = nil
	if err = a.L.LoadMultiAddresses(ctx, tx, true, &a, nil); err != nil {
		t.Fatal(err)
	}
	if got := len(a.R.MultiAddresses); got != 2 {
		t.Error("number of eager loaded records wrong, got:", got)
	}

	if t.Failed() {
		t.Logf("%#v", check)
	}
}

func testConnectionEventToManyAddOpMultiAddresses(t *testing.T) {
	var err error

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()

	var a ConnectionEvent
	var b, c, d, e MultiAddress

	seed := randomize.NewSeed()
	if err = randomize.Struct(seed, &a, connectionEventDBTypes, false, strmangle.SetComplement(connectionEventPrimaryKeyColumns, connectionEventColumnsWithoutDefault)...); err != nil {
		t.Fatal(err)
	}
	foreigners := []*MultiAddress{&b, &c, &d, &e}
	for _, x := range foreigners {
		if err = randomize.Struct(seed, x, multiAddressDBTypes, false, strmangle.SetComplement(multiAddressPrimaryKeyColumns, multiAddressColumnsWithoutDefault)...); err != nil {
			t.Fatal(err)
		}
	}

	if err := a.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Fatal(err)
	}
	if err = b.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Fatal(err)
	}
	if err = c.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Fatal(err)
	}

	foreignersSplitByInsertion := [][]*MultiAddress{
		{&b, &c},
		{&d, &e},
	}

	for i, x := range foreignersSplitByInsertion {
		err = a.AddMultiAddresses(ctx, tx, i != 0, x...)
		if err != nil {
			t.Fatal(err)
		}

		first := x[0]
		second := x[1]

		if first.R.ConnectionEvents[0] != &a {
			t.Error("relationship was not added properly to the slice")
		}
		if second.R.ConnectionEvents[0] != &a {
			t.Error("relationship was not added properly to the slice")
		}

		if a.R.MultiAddresses[i*2] != first {
			t.Error("relationship struct slice not set to correct value")
		}
		if a.R.MultiAddresses[i*2+1] != second {
			t.Error("relationship struct slice not set to correct value")
		}

		count, err := a.MultiAddresses().Count(ctx, tx)
		if err != nil {
			t.Fatal(err)
		}
		if want := int64((i + 1) * 2); count != want {
			t.Error("want", want, "got", count)
		}
	}
}

func testConnectionEventToManySetOpMultiAddresses(t *testing.T) {
	var err error

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()

	var a ConnectionEvent
	var b, c, d, e MultiAddress

	seed := randomize.NewSeed()
	if err = randomize.Struct(seed, &a, connectionEventDBTypes, false, strmangle.SetComplement(connectionEventPrimaryKeyColumns, connectionEventColumnsWithoutDefault)...); err != nil {
		t.Fatal(err)
	}
	foreigners := []*MultiAddress{&b, &c, &d, &e}
	for _, x := range foreigners {
		if err = randomize.Struct(seed, x, multiAddressDBTypes, false, strmangle.SetComplement(multiAddressPrimaryKeyColumns, multiAddressColumnsWithoutDefault)...); err != nil {
			t.Fatal(err)
		}
	}

	if err = a.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Fatal(err)
	}
	if err = b.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Fatal(err)
	}
	if err = c.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Fatal(err)
	}

	err = a.SetMultiAddresses(ctx, tx, false, &b, &c)
	if err != nil {
		t.Fatal(err)
	}

	count, err := a.MultiAddresses().Count(ctx, tx)
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Error("count was wrong:", count)
	}

	err = a.SetMultiAddresses(ctx, tx, true, &d, &e)
	if err != nil {
		t.Fatal(err)
	}

	count, err = a.MultiAddresses().Count(ctx, tx)
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Error("count was wrong:", count)
	}

	// The following checks cannot be implemented since we have no handle
	// to these when we call Set(). Leaving them here as wishful thinking
	// and to let people know there's dragons.
	//
	// if len(b.R.ConnectionEvents) != 0 {
	// 	t.Error("relationship was not removed properly from the slice")
	// }
	// if len(c.R.ConnectionEvents) != 0 {
	// 	t.Error("relationship was not removed properly from the slice")
	// }
	if d.R.ConnectionEvents[0] != &a {
		t.Error("relationship was not added properly to the slice")
	}
	if e.R.ConnectionEvents[0] != &a {
		t.Error("relationship was not added properly to the slice")
	}

	if a.R.MultiAddresses[0] != &d {
		t.Error("relationship struct slice not set to correct value")
	}
	if a.R.MultiAddresses[1] != &e {
		t.Error("relationship struct slice not set to correct value")
	}
}

func testConnectionEventToManyRemoveOpMultiAddresses(t *testing.T) {
	var err error

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()

	var a ConnectionEvent
	var b, c, d, e MultiAddress

	seed := randomize.NewSeed()
	if err = randomize.Struct(seed, &a, connectionEventDBTypes, false, strmangle.SetComplement(connectionEventPrimaryKeyColumns, connectionEventColumnsWithoutDefault)...); err != nil {
		t.Fatal(err)
	}
	foreigners := []*MultiAddress{&b, &c, &d, &e}
	for _, x := range foreigners {
		if err = randomize.Struct(seed, x, multiAddressDBTypes, false, strmangle.SetComplement(multiAddressPrimaryKeyColumns, multiAddressColumnsWithoutDefault)...); err != nil {
			t.Fatal(err)
		}
	}

	if err := a.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Fatal(err)
	}

	err = a.AddMultiAddresses(ctx, tx, true, foreigners...)
	if err != nil {
		t.Fatal(err)
	}

	count, err := a.MultiAddresses().Count(ctx, tx)
	if err != nil {
		t.Fatal(err)
	}
	if count != 4 {
		t.Error("count was wrong:", count)
	}

	err = a.RemoveMultiAddresses(ctx, tx, foreigners[:2]...)
	if err != nil {
		t.Fatal(err)
	}

	count, err = a.MultiAddresses().Count(ctx, tx)
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Error("count was wrong:", count)
	}

	if len(b.R.ConnectionEvents) != 0 {
		t.Error("relationship was not removed properly from the slice")
	}
	if len(c.R.ConnectionEvents) != 0 {
		t.Error("relationship was not removed properly from the slice")
	}
	if d.R.ConnectionEvents[0] != &a {
		t.Error("relationship was not added properly to the foreign struct")
	}
	if e.R.ConnectionEvents[0] != &a {
		t.Error("relationship was not added properly to the foreign struct")
	}

	if len(a.R.MultiAddresses) != 2 {
		t.Error("should have preserved two relationships")
	}

	// Removal doesn't do a stable deletion for performance so we have to flip the order
	if a.R.MultiAddresses[1] != &d {
		t.Error("relationship to d should have been preserved")
	}
	if a.R.MultiAddresses[0] != &e {
		t.Error("relationship to e should have been preserved")
	}
}

func testConnectionEventToOnePeerUsingLocal(t *testing.T) {
	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()

	var local ConnectionEvent
	var foreign Peer

	seed := randomize.NewSeed()
	if err := randomize.Struct(seed, &local, connectionEventDBTypes, false, connectionEventColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize ConnectionEvent struct: %s", err)
	}
	if err := randomize.Struct(seed, &foreign, peerDBTypes, false, peerColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Peer struct: %s", err)
	}

	if err := foreign.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Fatal(err)
	}

	local.LocalID = foreign.ID
	if err := local.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Fatal(err)
	}

	check, err := local.Local().One(ctx, tx)
	if err != nil {
		t.Fatal(err)
	}

	if check.ID != foreign.ID {
		t.Errorf("want: %v, got %v", foreign.ID, check.ID)
	}

	slice := ConnectionEventSlice{&local}
	if err = local.L.LoadLocal(ctx, tx, false, (*[]*ConnectionEvent)(&slice), nil); err != nil {
		t.Fatal(err)
	}
	if local.R.Local == nil {
		t.Error("struct should have been eager loaded")
	}

	local.R.Local = nil
	if err = local.L.LoadLocal(ctx, tx, true, &local, nil); err != nil {
		t.Fatal(err)
	}
	if local.R.Local == nil {
		t.Error("struct should have been eager loaded")
	}
}

func testConnectionEventToOneMultiAddressUsingConnMultiAddress(t *testing.T) {
	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()

	var local ConnectionEvent
	var foreign MultiAddress

	seed := randomize.NewSeed()
	if err := randomize.Struct(seed, &local, connectionEventDBTypes, false, connectionEventColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize ConnectionEvent struct: %s", err)
	}
	if err := randomize.Struct(seed, &foreign, multiAddressDBTypes, false, multiAddressColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize MultiAddress struct: %s", err)
	}

	if err := foreign.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Fatal(err)
	}

	local.ConnMultiAddressID = foreign.ID
	if err := local.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Fatal(err)
	}

	check, err := local.ConnMultiAddress().One(ctx, tx)
	if err != nil {
		t.Fatal(err)
	}

	if check.ID != foreign.ID {
		t.Errorf("want: %v, got %v", foreign.ID, check.ID)
	}

	slice := ConnectionEventSlice{&local}
	if err = local.L.LoadConnMultiAddress(ctx, tx, false, (*[]*ConnectionEvent)(&slice), nil); err != nil {
		t.Fatal(err)
	}
	if local.R.ConnMultiAddress == nil {
		t.Error("struct should have been eager loaded")
	}

	local.R.ConnMultiAddress = nil
	if err = local.L.LoadConnMultiAddress(ctx, tx, true, &local, nil); err != nil {
		t.Fatal(err)
	}
	if local.R.ConnMultiAddress == nil {
		t.Error("struct should have been eager loaded")
	}
}

func testConnectionEventToOnePeerUsingRemote(t *testing.T) {
	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()

	var local ConnectionEvent
	var foreign Peer

	seed := randomize.NewSeed()
	if err := randomize.Struct(seed, &local, connectionEventDBTypes, false, connectionEventColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize ConnectionEvent struct: %s", err)
	}
	if err := randomize.Struct(seed, &foreign, peerDBTypes, false, peerColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Peer struct: %s", err)
	}

	if err := foreign.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Fatal(err)
	}

	local.RemoteID = foreign.ID
	if err := local.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Fatal(err)
	}

	check, err := local.Remote().One(ctx, tx)
	if err != nil {
		t.Fatal(err)
	}

	if check.ID != foreign.ID {
		t.Errorf("want: %v, got %v", foreign.ID, check.ID)
	}

	slice := ConnectionEventSlice{&local}
	if err = local.L.LoadRemote(ctx, tx, false, (*[]*ConnectionEvent)(&slice), nil); err != nil {
		t.Fatal(err)
	}
	if local.R.Remote == nil {
		t.Error("struct should have been eager loaded")
	}

	local.R.Remote = nil
	if err = local.L.LoadRemote(ctx, tx, true, &local, nil); err != nil {
		t.Fatal(err)
	}
	if local.R.Remote == nil {
		t.Error("struct should have been eager loaded")
	}
}

func testConnectionEventToOneSetOpPeerUsingLocal(t *testing.T) {
	var err error

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()

	var a ConnectionEvent
	var b, c Peer

	seed := randomize.NewSeed()
	if err = randomize.Struct(seed, &a, connectionEventDBTypes, false, strmangle.SetComplement(connectionEventPrimaryKeyColumns, connectionEventColumnsWithoutDefault)...); err != nil {
		t.Fatal(err)
	}
	if err = randomize.Struct(seed, &b, peerDBTypes, false, strmangle.SetComplement(peerPrimaryKeyColumns, peerColumnsWithoutDefault)...); err != nil {
		t.Fatal(err)
	}
	if err = randomize.Struct(seed, &c, peerDBTypes, false, strmangle.SetComplement(peerPrimaryKeyColumns, peerColumnsWithoutDefault)...); err != nil {
		t.Fatal(err)
	}

	if err := a.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Fatal(err)
	}
	if err = b.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Fatal(err)
	}

	for i, x := range []*Peer{&b, &c} {
		err = a.SetLocal(ctx, tx, i != 0, x)
		if err != nil {
			t.Fatal(err)
		}

		if a.R.Local != x {
			t.Error("relationship struct not set to correct value")
		}

		if x.R.LocalConnectionEvents[0] != &a {
			t.Error("failed to append to foreign relationship struct")
		}
		if a.LocalID != x.ID {
			t.Error("foreign key was wrong value", a.LocalID)
		}

		zero := reflect.Zero(reflect.TypeOf(a.LocalID))
		reflect.Indirect(reflect.ValueOf(&a.LocalID)).Set(zero)

		if err = a.Reload(ctx, tx); err != nil {
			t.Fatal("failed to reload", err)
		}

		if a.LocalID != x.ID {
			t.Error("foreign key was wrong value", a.LocalID, x.ID)
		}
	}
}
func testConnectionEventToOneSetOpMultiAddressUsingConnMultiAddress(t *testing.T) {
	var err error

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()

	var a ConnectionEvent
	var b, c MultiAddress

	seed := randomize.NewSeed()
	if err = randomize.Struct(seed, &a, connectionEventDBTypes, false, strmangle.SetComplement(connectionEventPrimaryKeyColumns, connectionEventColumnsWithoutDefault)...); err != nil {
		t.Fatal(err)
	}
	if err = randomize.Struct(seed, &b, multiAddressDBTypes, false, strmangle.SetComplement(multiAddressPrimaryKeyColumns, multiAddressColumnsWithoutDefault)...); err != nil {
		t.Fatal(err)
	}
	if err = randomize.Struct(seed, &c, multiAddressDBTypes, false, strmangle.SetComplement(multiAddressPrimaryKeyColumns, multiAddressColumnsWithoutDefault)...); err != nil {
		t.Fatal(err)
	}

	if err := a.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Fatal(err)
	}
	if err = b.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Fatal(err)
	}

	for i, x := range []*MultiAddress{&b, &c} {
		err = a.SetConnMultiAddress(ctx, tx, i != 0, x)
		if err != nil {
			t.Fatal(err)
		}

		if a.R.ConnMultiAddress != x {
			t.Error("relationship struct not set to correct value")
		}

		if x.R.ConnMultiAddressConnectionEvents[0] != &a {
			t.Error("failed to append to foreign relationship struct")
		}
		if a.ConnMultiAddressID != x.ID {
			t.Error("foreign key was wrong value", a.ConnMultiAddressID)
		}

		zero := reflect.Zero(reflect.TypeOf(a.ConnMultiAddressID))
		reflect.Indirect(reflect.ValueOf(&a.ConnMultiAddressID)).Set(zero)

		if err = a.Reload(ctx, tx); err != nil {
			t.Fatal("failed to reload", err)
		}

		if a.ConnMultiAddressID != x.ID {
			t.Error("foreign key was wrong value", a.ConnMultiAddressID, x.ID)
		}
	}
}
func testConnectionEventToOneSetOpPeerUsingRemote(t *testing.T) {
	var err error

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()

	var a ConnectionEvent
	var b, c Peer

	seed := randomize.NewSeed()
	if err = randomize.Struct(seed, &a, connectionEventDBTypes, false, strmangle.SetComplement(connectionEventPrimaryKeyColumns, connectionEventColumnsWithoutDefault)...); err != nil {
		t.Fatal(err)
	}
	if err = randomize.Struct(seed, &b, peerDBTypes, false, strmangle.SetComplement(peerPrimaryKeyColumns, peerColumnsWithoutDefault)...); err != nil {
		t.Fatal(err)
	}
	if err = randomize.Struct(seed, &c, peerDBTypes, false, strmangle.SetComplement(peerPrimaryKeyColumns, peerColumnsWithoutDefault)...); err != nil {
		t.Fatal(err)
	}

	if err := a.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Fatal(err)
	}
	if err = b.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Fatal(err)
	}

	for i, x := range []*Peer{&b, &c} {
		err = a.SetRemote(ctx, tx, i != 0, x)
		if err != nil {
			t.Fatal(err)
		}

		if a.R.Remote != x {
			t.Error("relationship struct not set to correct value")
		}

		if x.R.RemoteConnectionEvents[0] != &a {
			t.Error("failed to append to foreign relationship struct")
		}
		if a.RemoteID != x.ID {
			t.Error("foreign key was wrong value", a.RemoteID)
		}

		zero := reflect.Zero(reflect.TypeOf(a.RemoteID))
		reflect.Indirect(reflect.ValueOf(&a.RemoteID)).Set(zero)

		if err = a.Reload(ctx, tx); err != nil {
			t.Fatal("failed to reload", err)
		}

		if a.RemoteID != x.ID {
			t.Error("foreign key was wrong value", a.RemoteID, x.ID)
		}
	}
}

func testConnectionEventsReload(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &ConnectionEvent{}
	if err = randomize.Struct(seed, o, connectionEventDBTypes, true, connectionEventColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize ConnectionEvent struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	if err = o.Reload(ctx, tx); err != nil {
		t.Error(err)
	}
}

func testConnectionEventsReloadAll(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &ConnectionEvent{}
	if err = randomize.Struct(seed, o, connectionEventDBTypes, true, connectionEventColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize ConnectionEvent struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	slice := ConnectionEventSlice{o}

	if err = slice.ReloadAll(ctx, tx); err != nil {
		t.Error(err)
	}
}

func testConnectionEventsSelect(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &ConnectionEvent{}
	if err = randomize.Struct(seed, o, connectionEventDBTypes, true, connectionEventColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize ConnectionEvent struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	slice, err := ConnectionEvents().All(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if len(slice) != 1 {
		t.Error("want one record, got:", len(slice))
	}
}

var (
	connectionEventDBTypes = map[string]string{`ID`: `integer`, `LocalID`: `bigint`, `RemoteID`: `bigint`, `ConnMultiAddressID`: `bigint`, `OpenedAt`: `timestamp with time zone`, `CreatedAt`: `timestamp with time zone`}
	_                      = bytes.MinRead
)

func testConnectionEventsUpdate(t *testing.T) {
	t.Parallel()

	if 0 == len(connectionEventPrimaryKeyColumns) {
		t.Skip("Skipping table with no primary key columns")
	}
	if len(connectionEventAllColumns) == len(connectionEventPrimaryKeyColumns) {
		t.Skip("Skipping table with only primary key columns")
	}

	seed := randomize.NewSeed()
	var err error
	o := &ConnectionEvent{}
	if err = randomize.Struct(seed, o, connectionEventDBTypes, true, connectionEventColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize ConnectionEvent struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	count, err := ConnectionEvents().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 1 {
		t.Error("want one record, got:", count)
	}

	if err = randomize.Struct(seed, o, connectionEventDBTypes, true, connectionEventPrimaryKeyColumns...); err != nil {
		t.Errorf("Unable to randomize ConnectionEvent struct: %s", err)
	}

	if rowsAff, err := o.Update(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	} else if rowsAff != 1 {
		t.Error("should only affect one row but affected", rowsAff)
	}
}

func testConnectionEventsSliceUpdateAll(t *testing.T) {
	t.Parallel()

	if len(connectionEventAllColumns) == len(connectionEventPrimaryKeyColumns) {
		t.Skip("Skipping table with only primary key columns")
	}

	seed := randomize.NewSeed()
	var err error
	o := &ConnectionEvent{}
	if err = randomize.Struct(seed, o, connectionEventDBTypes, true, connectionEventColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize ConnectionEvent struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	count, err := ConnectionEvents().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 1 {
		t.Error("want one record, got:", count)
	}

	if err = randomize.Struct(seed, o, connectionEventDBTypes, true, connectionEventPrimaryKeyColumns...); err != nil {
		t.Errorf("Unable to randomize ConnectionEvent struct: %s", err)
	}

	// Remove Primary keys and unique columns from what we plan to update
	var fields []string
	if strmangle.StringSliceMatch(connectionEventAllColumns, connectionEventPrimaryKeyColumns) {
		fields = connectionEventAllColumns
	} else {
		fields = strmangle.SetComplement(
			connectionEventAllColumns,
			connectionEventPrimaryKeyColumns,
		)
		fields = strmangle.SetComplement(fields, connectionEventGeneratedColumns)
	}

	value := reflect.Indirect(reflect.ValueOf(o))
	typ := reflect.TypeOf(o).Elem()
	n := typ.NumField()

	updateMap := M{}
	for _, col := range fields {
		for i := 0; i < n; i++ {
			f := typ.Field(i)
			if f.Tag.Get("boil") == col {
				updateMap[col] = value.Field(i).Interface()
			}
		}
	}

	slice := ConnectionEventSlice{o}
	if rowsAff, err := slice.UpdateAll(ctx, tx, updateMap); err != nil {
		t.Error(err)
	} else if rowsAff != 1 {
		t.Error("wanted one record updated but got", rowsAff)
	}
}

func testConnectionEventsUpsert(t *testing.T) {
	t.Parallel()

	if len(connectionEventAllColumns) == len(connectionEventPrimaryKeyColumns) {
		t.Skip("Skipping table with only primary key columns")
	}

	seed := randomize.NewSeed()
	var err error
	// Attempt the INSERT side of an UPSERT
	o := ConnectionEvent{}
	if err = randomize.Struct(seed, &o, connectionEventDBTypes, true); err != nil {
		t.Errorf("Unable to randomize ConnectionEvent struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Upsert(ctx, tx, false, nil, boil.Infer(), boil.Infer()); err != nil {
		t.Errorf("Unable to upsert ConnectionEvent: %s", err)
	}

	count, err := ConnectionEvents().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}
	if count != 1 {
		t.Error("want one record, got:", count)
	}

	// Attempt the UPDATE side of an UPSERT
	if err = randomize.Struct(seed, &o, connectionEventDBTypes, false, connectionEventPrimaryKeyColumns...); err != nil {
		t.Errorf("Unable to randomize ConnectionEvent struct: %s", err)
	}

	if err = o.Upsert(ctx, tx, true, nil, boil.Infer(), boil.Infer()); err != nil {
		t.Errorf("Unable to upsert ConnectionEvent: %s", err)
	}

	count, err = ConnectionEvents().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}
	if count != 1 {
		t.Error("want one record, got:", count)
	}
}
