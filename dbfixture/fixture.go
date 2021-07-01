package dbfixture

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/j2gg0s/bun"
	"github.com/j2gg0s/bun/schema"
)

type FixtureOption func(l *Fixture)

func WithRecreateTables() FixtureOption {
	return func(l *Fixture) {
		if l.truncateTables {
			panic("don't use WithDropTables together with WithTruncateTables")
		}
		l.recreateTables = true
		l.seenTables = make(map[string]struct{})
	}
}

func WithTruncateTables() FixtureOption {
	return func(l *Fixture) {
		if l.truncateTables {
			panic("don't use WithTruncateTables together with WithRecreateTables")
		}
		l.truncateTables = true
		l.seenTables = make(map[string]struct{})
	}
}

func WithTemplateFuncs(funcMap template.FuncMap) FixtureOption {
	return func(l *Fixture) {
		for k, v := range funcMap {
			l.funcMap[k] = v
		}
	}
}

type Fixture struct {
	db *bun.DB

	recreateTables bool
	truncateTables bool
	seenTables     map[string]struct{}

	funcMap   template.FuncMap
	modelRows map[string]map[string]interface{}
}

func New(db *bun.DB, opts ...FixtureOption) *Fixture {
	l := &Fixture{
		db: db,

		funcMap:   defaultFuncs(),
		modelRows: make(map[string]map[string]interface{}),
	}
	for _, opt := range opts {
		opt(l)
	}
	return l
}

func (l *Fixture) Row(id string) (interface{}, error) {
	ss := strings.Split(id, ".")
	if len(ss) != 2 {
		return nil, fmt.Errorf("fixture: invalid row id: %q", id)
	}
	model, rowID := ss[0], ss[1]

	rows, ok := l.modelRows[model]
	if !ok {
		return nil, fmt.Errorf("fixture: unknown model=%q", model)
	}

	row, ok := rows[rowID]
	if !ok {
		return nil, fmt.Errorf("fixture: can't find row=%q for model=%q", rowID, model)
	}

	return row, nil
}

func (l *Fixture) MustRow(id string) interface{} {
	row, err := l.Row(id)
	if err != nil {
		panic(err)
	}
	return row
}

func (l *Fixture) Load(ctx context.Context, fsys fs.FS, names ...string) error {
	for _, name := range names {
		if err := l.load(ctx, fsys, name); err != nil {
			return err
		}
	}
	return nil
}

func (l *Fixture) load(ctx context.Context, fsys fs.FS, name string) error {
	fh, err := fsys.Open(name)
	if err != nil {
		return err
	}

	var fixtures []fixtureData

	dec := yaml.NewDecoder(fh)
	if err := dec.Decode(&fixtures); err != nil {
		return err
	}

	for i := range fixtures {
		if err := l.addFixture(ctx, &fixtures[i]); err != nil {
			return err
		}
	}

	return nil
}

func (l *Fixture) addFixture(ctx context.Context, data *fixtureData) error {
	table := l.db.Dialect().Tables().ByModel(data.Model)
	if table == nil {
		return fmt.Errorf("fixture: can't find model=%q (use db.RegisterModel)", data.Model)
	}

	if l.recreateTables {
		if err := l.dropTable(ctx, table); err != nil {
			return err
		}
	} else if l.truncateTables {
		if err := l.truncateTable(ctx, table); err != nil {
			return err
		}
	}

	for _, row := range data.Rows {
		if err := l.addRow(ctx, table, row); err != nil {
			return err
		}
	}

	return nil
}

func (l *Fixture) addRow(ctx context.Context, table *schema.Table, row row) error {
	var rowID string
	strct := reflect.New(table.Type).Elem()

	for key, value := range row {
		if key == "_id" {
			if err := value.Decode(&rowID); err != nil {
				return err
			}
			continue
		}

		field, err := table.Field(key)
		if err != nil {
			return err
		}

		if value.Tag == "!!str" && isTemplate(value.Value) {
			res, err := l.eval(value.Value)
			if err != nil {
				return err
			}

			if res != value.Value {
				if err := field.ScanValue(strct, res); err != nil {
					return err
				}
				continue
			}
		}

		fv := field.Value(strct)
		if err := value.Decode(fv.Addr().Interface()); err != nil {
			return err
		}
	}

	iface := strct.Addr().Interface()
	if _, err := l.db.NewInsert().
		Model(iface).
		Exec(ctx); err != nil {
		return err
	}

	if rowID == "" && len(table.PKs) == 1 {
		pk := table.PKs[0]
		fv := pk.Value(strct)
		rowID = "pk" + asString(fv)
	}

	if rowID != "" {
		rows, ok := l.modelRows[table.TypeName]
		if !ok {
			rows = make(map[string]interface{})
			l.modelRows[table.TypeName] = rows
		}
		rows[rowID] = iface
	}

	return nil
}

func (l *Fixture) dropTable(ctx context.Context, table *schema.Table) error {
	if _, ok := l.seenTables[table.Name]; ok {
		return nil
	}
	l.seenTables[table.Name] = struct{}{}

	if _, err := l.db.NewDropTable().
		Model(table.ZeroIface).
		IfExists().
		Exec(ctx); err != nil {
		return err
	}

	if _, err := l.db.NewCreateTable().
		Model(table.ZeroIface).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

func (l *Fixture) truncateTable(ctx context.Context, table *schema.Table) error {
	if _, ok := l.seenTables[table.Name]; ok {
		return nil
	}
	l.seenTables[table.Name] = struct{}{}

	if _, err := l.db.NewTruncateTable().
		Model(table.ZeroIface).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

func (l *Fixture) eval(templ string) (string, error) {
	tpl, err := template.New("").Funcs(l.funcMap).Parse(templ)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer

	if err := tpl.Execute(&buf, l.modelRows); err != nil {
		return "", err
	}

	return buf.String(), nil
}

type fixtureData struct {
	Model string `yaml:"model"`
	Rows  []row  `yaml:"rows"`
}

type row map[string]yaml.Node

var tplRE = regexp.MustCompile(`\{\{ .+ \}\}`)

func isTemplate(s string) bool {
	return tplRE.MatchString(s)
}

func asString(rv reflect.Value) string {
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(rv.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(rv.Uint(), 10)
	case reflect.Float64:
		return strconv.FormatFloat(rv.Float(), 'g', -1, 64)
	case reflect.Float32:
		return strconv.FormatFloat(rv.Float(), 'g', -1, 32)
	case reflect.Bool:
		return strconv.FormatBool(rv.Bool())
	}
	return fmt.Sprintf("%v", rv.Interface())
}

func defaultFuncs() template.FuncMap {
	return template.FuncMap{
		"now": func() string {
			return time.Now().Format(time.RFC3339Nano)
		},
	}
}
