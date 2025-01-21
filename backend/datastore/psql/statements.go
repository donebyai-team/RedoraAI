package psql

import (
	"bytes"
	"context"
	"database/sql"
	"embed"
	_ "embed"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"text/template"

	"github.com/bobg/go-generics/v2/maps"
	"github.com/bobg/go-generics/v2/slices"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/shank318/doota/datastore"
	"go.uber.org/multierr"
)

//go:embed sql
var statements embed.FS

var templates *template.Template

func initTemplates() {
	var err error

	templates = template.New("sql").Funcs(template.FuncMap{
		"increment": func(i int) int {
			return i + 1
		},
		"is_last": func(i int, elements any) bool {
			rv := reflect.ValueOf(elements)
			rvKind := rv.Kind()

			if rvKind == reflect.Slice || rvKind == reflect.Array || rvKind == reflect.Map || rvKind == reflect.String {
				return i == rv.Len()-1
			}

			return false
		},
		"map": func(pairs ...any) (map[string]any, error) {
			if len(pairs)%2 != 0 {
				return nil, errors.New("misaligned map")
			}

			m := make(map[string]any, len(pairs)/2)

			for i := 0; i < len(pairs); i += 2 {
				key, ok := pairs[i].(string)

				if !ok {
					return nil, fmt.Errorf("cannot use type %T as map key", pairs[i])
				}
				m[key] = pairs[i+1]
			}
			return m, nil
		},
	})

	templates, err = templates.ParseFS(statements, "sql/*/*.sql")
	if err != nil {
		panic(fmt.Errorf("unable to parse embedded sql statements: %w", err))
	}
}

// onDiskStatement returns the content of the file located at `file` in the `sql` folder.
//
// **Important** Not safe for concurrent usage, should be all used from the same goroutine!
func onDiskStatement(file string) string {
	if templates == nil {
		initTemplates()
	}

	_, name, found := strings.Cut(file, "/")
	if !found {
		panic(fmt.Errorf("unable to find 'folder/name' in %q", file))
	}

	buffer := bytes.NewBuffer(make([]byte, 0, 1024))
	if err := templates.ExecuteTemplate(buffer, name, map[string]any{}); err != nil {
		panic(fmt.Errorf("unable to execute embedded sql statements: %w", err))
	}

	return buffer.String()
}

func getOne[T any](ctx context.Context, db *Database, statement string, args map[string]any) (out *T, err error) {
	stmt := db.mustGetStmt(statement)
	var model T
	err = stmt.GetContext(ctx, &model, args)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, datastore.NotFound
		}

		wheres := make([]string, 0, len(args))
		for key, value := range args {
			wheres = append(wheres, fmt.Sprintf("%s = %v", key, value))
		}

		return nil, fmt.Errorf("failed %s with %s: %w", strings.ReplaceAll(statement, "_", " "), strings.Join(wheres, ", "), err)
	}

	return &model, nil
}

func getOneMapped[T any, R any](ctx context.Context, db *Database, statement string, args map[string]any, mapper func(in *T) R) (out R, err error) {
	model, err := getOne[T](ctx, db, statement, args)
	if err == nil {
		return mapper(model), nil
	}

	return out, err
}

func getMany[T any](ctx context.Context, db *Database, statement string, args map[string]any) (out []*T, err error) {
	stmt := db.mustGetStmt(statement)
	var models []*T
	err = stmt.SelectContext(ctx, &models, args)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		wheres := make([]string, 0, len(args))
		for key, value := range args {
			wheres = append(wheres, fmt.Sprintf("%s = %v", key, value))
		}

		return nil, fmt.Errorf("failed %s with %s: %w", strings.ReplaceAll(statement, "_", " "), strings.Join(wheres, ", "), err)
	}

	return models, nil
}

func getManyMapped[T any, R any](ctx context.Context, db *Database, statement string, args map[string]any, mapper func(in *T) R) (out []R, err error) {
	models, err := getMany[T](ctx, db, statement, args)
	if err == nil {
		return slices.Map(models, mapper), nil
	}

	return out, err
}

type Identifiable interface {
	GetIdentifier() string
}

// getAll returns all models identified by the given identifiers in the exact sequence they were requested. If a given ID
// couldn't be found, the function will return a [datastore.NotFound] error, the internal message will contain the missing
// identifiers.
//
// This method differs from [getMany] in that it guarantees the order of the returned models to be in the same sequence as
// the received identifiers.
func getAll[T Identifiable](ctx context.Context, db *Database, statement string, key string, identifiers []string) (out []*T, err error) {
	models, err := getMany[T](ctx, db, statement, map[string]any{
		key: pq.Array(unique(identifiers)),
	})
	if err == nil {
		byIds := make(map[string]*T, len(models))
		for _, model := range models {
			byIds[(*model).GetIdentifier()] = model
		}

		out = make([]*T, len(identifiers))
		var missing []string

		for i, identifier := range identifiers {
			result, found := byIds[identifier]
			if !found {
				missing = append(missing, identifier)
				continue
			}

			out[i] = result
		}

		if len(missing) > 0 {
			return nil, fmt.Errorf("failed to find %s %s: %w", key, strings.Join(missing, ", "), datastore.NotFound)
		}
	}

	return out, err
}

// getAllMapped returns all models identified by the given identifiers in the exact sequence they were requested. It's
// essentially a simple wrapper around [getAll] that maps the models to a different type.
func getAllMapped[T Identifiable, R any](ctx context.Context, db *Database, statement string, key string, identifiers []string, mapper func(in *T) R) (out []R, err error) {
	models, err := getAll[T](ctx, db, statement, key, identifiers)
	if err == nil {
		return slices.Map(models, mapper), nil
	}

	return out, err
}

func unique[T comparable](items []T) []T {
	unique := make(map[T]struct{}, len(items))
	for _, item := range items {
		unique[item] = struct{}{}
	}

	return maps.Keys(unique)
}

func executePotentialRollback(tx *sqlx.Tx, err error) error {
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return multierr.Combine(err, rollbackErr)
		}

		return err
	}

	return nil
}
