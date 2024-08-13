package oracle

import (
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
)

type Oracle struct {
	db     *dbShard
	events chan *v1.Model
}

func New(path string) (*Oracle, error) {
	db, err := newDBShard(path)
	if err != nil {
		return nil, err
	}
	return &Oracle{db: db, events: make(chan *v1.Model, 2<<10)}, nil
}

func (o *Oracle) Close() error {
	close(o.events)
	return o.db.Close()
}

func (o *Oracle) Save(e *v1.Model) {
	o.db.Write(e)
}
