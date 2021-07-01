package bun

import (
	"context"
	"database/sql"
	"reflect"
	"sync/atomic"
	"time"

	"github.com/j2gg0s/bun/schema"
)

type QueryEvent struct {
	DB *DB

	QueryAppender schema.QueryAppender
	Query         string
	QueryArgs     []interface{}

	StartTime time.Time
	Result    sql.Result
	Err       error

	Stash map[interface{}]interface{}
}

type QueryHook interface {
	BeforeQuery(context.Context, *QueryEvent) context.Context
	AfterQuery(context.Context, *QueryEvent)
}

func (db *DB) beforeQuery(
	ctx context.Context,
	queryApp schema.QueryAppender,
	query string,
	queryArgs []interface{},
) (context.Context, *QueryEvent) {
	atomic.AddUint64(&db.stats.Queries, 1)

	if len(db.queryHooks) == 0 {
		return ctx, nil
	}

	event := &QueryEvent{
		DB: db,

		QueryAppender: queryApp,
		Query:         query,
		QueryArgs:     queryArgs,

		StartTime: time.Now(),
	}

	for _, hook := range db.queryHooks {
		ctx = hook.BeforeQuery(ctx, event)
	}

	return ctx, event
}

func (db *DB) afterQuery(
	ctx context.Context,
	event *QueryEvent,
	res sql.Result,
	err error,
) {
	switch err {
	case nil, sql.ErrNoRows:
		// nothing
	default:
		atomic.AddUint64(&db.stats.Errors, 1)
	}

	if event == nil {
		return
	}

	event.Result = res
	event.Err = err

	db.afterQueryFromIndex(ctx, event, len(db.queryHooks)-1)
}

func (db *DB) afterQueryFromIndex(ctx context.Context, event *QueryEvent, hookIndex int) {
	for ; hookIndex >= 0; hookIndex-- {
		db.queryHooks[hookIndex].AfterQuery(ctx, event)
	}
}

//------------------------------------------------------------------------------

func callBeforeScanHook(ctx context.Context, v reflect.Value) error {
	return v.Interface().(schema.BeforeScanHook).BeforeScan(ctx)
}

func callAfterScanHook(ctx context.Context, v reflect.Value) error {
	return v.Interface().(schema.AfterScanHook).AfterScan(ctx)
}
