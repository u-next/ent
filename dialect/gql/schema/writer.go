package schema

import (
	"context"

	"entgo.io/ent/dialect"
)

type nopDriver struct {
	dialect.Driver
	dialect string
}

func (d nopDriver) Dialect() string { return d.dialect }

func (nopDriver) Query(context.Context, string, any, any) error {
	return nil
}
