package crime

import (
	"fmt"

	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CrimeTable struct {
	mgm.DefaultModel `json:"-"`
	Title            string `json:"title"`
	Year             int    `json:"year"`
	DatasetID        uint64 `json:"dataset_id"`
	Data             Data   `json:"data,omitempty"`
}

type Data struct {
	Columns []string
	Entries [][]string
}

func AllTables(nodata bool) ([]*CrimeTable, error) {
	coll := mgm.Coll(&CrimeTable{})
	ctx := mgm.Ctx()

	opts := options.Find()

	if nodata {
		opts.SetProjection(bson.M{"data": 0})
	}

	cur, err := coll.Find(ctx, bson.M{}, opts)
	if err != nil {
		return []*CrimeTable{}, fmt.Errorf("error fetching from db: %s", err)
	}
	tables := make([]*CrimeTable, 0)
	err = cur.All(ctx, &tables)
	if err != nil {
		return []*CrimeTable{}, fmt.Errorf("error fetching from db: %s", err)
	}
	return tables, nil
}

func TablesByYear(year int, nodata bool) ([]*CrimeTable, error) {
	coll := mgm.Coll(&CrimeTable{})
	ctx := mgm.Ctx()

	opts := options.Find()

	if nodata {
		opts.SetProjection(bson.M{"data": 0})
	}

	cur, err := coll.Find(ctx, bson.M{"year": year}, opts)
	if err != nil {
		return []*CrimeTable{}, fmt.Errorf("error fetching from db: %s", err)
	}
	tables := make([]*CrimeTable, 0)
	err = cur.All(ctx, &tables)
	if err != nil {
		return []*CrimeTable{}, fmt.Errorf("error fetching from db: %s", err)
	}
	return tables, nil
}

func TableByID(id uint64) (*CrimeTable, error) {
	entry := &CrimeTable{}
	err := mgm.Coll(entry).First(bson.M{"datasetid": id}, entry)
	if err != nil {
		return nil, fmt.Errorf("error fetching from db: %s", err)
	}
	return entry, nil
}
