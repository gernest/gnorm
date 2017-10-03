package postgres

import (
	"io/ioutil"
	"path/filepath"
)

const schemaTpl = `
package gnorm

{{ range (makeSlice "int" "string" "sql.NullString" "int64" "sql.NullInt64" "float64" "sql.NullFloat64" "bool" "sql.NullBool" "time.Time" "mysql.NullTime" "uint32" ) }}
{{ $fieldName := title (replace . "." "" 1) }}
// {{$fieldName}}Field is a component that returns a WhereClause that contains a
// comparison based on its field and a strongly typed value.
type {{$fieldName}}Field string

// Equals returns a WhereClause for this field.
func (f {{$fieldName}}Field) Equals(v {{.}}) WhereClause {
	return whereClause{
		field: string(f),
		comp:  compEqual,
		value: v,
	}
}

// GreaterThan returns a WhereClause for this field.
func (f {{$fieldName}}Field) GreaterThan(v {{.}}) WhereClause {
	return whereClause{
		field: string(f),
		comp:  compGreater,
		value: v,
	}
}

// LessThan returns a WhereClause for this field.
func (f {{$fieldName}}Field) LessThan(v {{.}}) WhereClause {
	return whereClause{
		field: string(f),
		comp:  compEqual,
		value: v,
	}
}

// GreaterOrEqual returns a WhereClause for this field.
func (f {{$fieldName}}Field) GreaterOrEqual(v {{.}}) WhereClause {
	return whereClause{
		field: string(f),
		comp:  compGTE,
		value: v,
	}
}

// LessOrEqual returns a WhereClause for this field.
func (f {{$fieldName}}Field) LessOrEqual(v {{.}}) WhereClause {
	return whereClause{
		field: string(f),
		comp:  compLTE,
		value: v,
	}
}

// NotEqual returns a WhereClause for this field.
func (f {{$fieldName}}Field) NotEqual(v {{.}}) WhereClause {
	return whereClause{
		field: string(f),
		comp:  compNE,
		value: v,
	}
}

// In returns a WhereClause for this field.
func (f {{$fieldName}}Field) In(vals []{{.}}) WhereClause {
	values := make([]interface{}, len(vals))
	for x := range vals {
		values[x] = vals[x]
	}
	return inClause{
		field: string(f),
		values: values,
	}
}

{{end}}
`

const tableTpl = `
{{with .Table}}
package {{toLower .Name}}

import "gnorm.org/gnorm/database/drivers/postgres/gnorm"

{{$table := .DBName -}}
{{$schema := .DBSchema -}}
// Row represents a row from '{{ $table }}'.
type Row struct {
{{- range .Columns }}
	{{ .Name }} {{ .Type }}  // {{ .DBName }}
{{- end }}
}


// Field values for every column in {{.Name}}.
var (
{{- range .Columns }}
	{{.Name}}Col gnorm.{{ title (replace .Type "." "" 1) }}Field = "{{ .DBName }}"
{{- end -}}
)

// Query retrieves rows from '{{ $table }}' as a slice of Row.
func Query(db gnorm.DB, where gnorm.WhereClause) ([]*Row, error) {
	 origsqlstr := "SELECT "+
		"{{ join .Columns.DBNames ", " }}"+
		"FROM {{$schema}}.{{ $table }} WHERE ("

	idx := 1
	sqlstr := origsqlstr + where.String(&idx) + ") "

	var vals []*Row
	q, err := db.Query(sqlstr, where.Values()...)
	if err != nil {
		return nil, err
	}
	for q.Next() {
		r := Row{}

		err = q.Scan(
			{{- $lastCol := dec (len .Columns)  }}
			{{- range $x, $c :=  .Columns -}}
				&r.{{$c.Name}}{{ if ne $x $lastCol}}, {{end -}}
			{{end -}}
		)			
		if err != nil {
			return nil, err
		}

		vals = append(vals, &r)
	}
	return vals, nil
}

{{end}}`

const configTpl = `# This is the gnorm.toml that generates the DB access code for mysql databases.
# Its output is contained in the database/drivers/postgres/gnorm driectory.  It
# assumes you have a postgres instance running locally.

ConnStr = "dbname=gnorm-db host=127.0.0.1 sslmode=disable user=gnorm-user"

DBType = "postgres"

Schemas = ["information_schema"]

NameConversion = "{{. | pascal}}"

IncludeTables = ["tables", "columns"]

PostRun = ["goimports", "-w", "$GNORMFILE"]

[TablePaths]
"gnorm/{{toLower .Table}}/{{toLower .Table}}.go" = "templates/table.gotmpl"

[SchemaPaths]
"gnorm/fields.go" = "templates/schema.gotmpl"

[TypeMap]
"timestamp with time zone" = "time.Time"
"text" = "string"
"boolean" = "bool"
"uuid" = "uuid.UUID"
"character varying" = "string"
"integer" = "int"
"numeric" = "float64"

[NullableTypeMap]
"timestamp with time zone" = "pq.NullTime"
"text" = "sql.NullString"
"boolean" = "sql.NullBool"
"uuid" = "uuid.NullUUID"
"character varying" = "sql.NullString"
"integer" = "sql.NullInt64"
"numeric" = "sql.NullFloat64"
`

func InitTemplates(root string) error {
	s := []struct {
		path string
		tpl  string
	}{
		{"schema.gotmpl", schemaTpl},
		{"table.gotmpl", tableTpl},
		{"gnorm.toml", configTpl},
	}

	for _, v := range s {
		err := ioutil.WriteFile(filepath.Join(root, v.path), []byte(v.tpl), 0600)
		if err != nil {
			return err
		}
	}
	return nil
}
