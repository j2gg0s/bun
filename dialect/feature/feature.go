package feature

import "github.com/j2gg0s/bun/internal"

type Feature = internal.Flag

const DefaultFeatures = Returning | TableCascade

const (
	Returning Feature = 1 << iota
	DefaultPlaceholder
	DoubleColonCast
	ValuesRow
	UpdateMultiTable
	InsertTableAlias
	AutoIncrement
	TableCascade
	TableIdentity
	TableTruncate
	OnDuplicateKey
)
