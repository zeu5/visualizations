package datagovin

import (
	"fmt"
	"time"

	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Catalog struct {
	mgm.DefaultModel
	Title        string                 `json:"title" bson:"title"`
	Created      time.Time              `json:"created" bson:"created"`
	LastModified time.Time              `json:"last_modified" bson:"last_modified"`
	CatID        uint64                 `json:"cat_id" bson:"cat_id"`
	Departments  []string               `json:"department" bson:"department"`
	Other        map[string]interface{} `json:"other" bson:"other"`
}

func (c *Catalog) CollectionName() string {
	return "data_gov_in_catalogs"
}

type Dataset struct {
	mgm.DefaultModel
	DID          uint64                 `json:"d_id" bson:"d_id"`
	Title        string                 `json:"title" bson:"title"`
	CatID        uint64                 `json:"cat_id" bson:"cat_id"`
	Created      time.Time              `json:"created" bson:"created"`
	LastModified time.Time              `json:"last_modified" bson:"last_modified"`
	Other        map[string]interface{} `json:"other" bson:"other"`
	Data         Data                   `json:"data" bson:"data"`
}

func (d *Dataset) CollectionName() string {
	return "data_gov_in_datasets"
}

type Data struct {
	Fields  []map[string]string `json:"fields" bson:"fields"`
	Entries [][]interface{}     `json:"entries" bson:"entries"`
}

func InitializeDB(url string) error {
	return mgm.SetDefaultConfig(nil, "vis", options.Client().ApplyURI(url))
}

func GetAllCatalog() ([]*Catalog, error) {
	coll := mgm.Coll(&Catalog{})
	ctx := mgm.Ctx()
	var catalogs []*Catalog

	cur, err := coll.Find(ctx, bson.M{})
	if err != nil {
		return []*Catalog{}, fmt.Errorf("could not fetch from db: %s", err)
	}
	err = cur.All(ctx, &catalogs)
	if err != nil {
		return []*Catalog{}, fmt.Errorf("could not decode records: %s", err)
	}

	return catalogs, nil
}

func GetCatalogInfo(c *Catalog) ([]*Dataset, error) {
	coll := mgm.Coll(&Dataset{})
	ctx := mgm.Ctx()
	cur, err := coll.Find(ctx, bson.M{"cat_id": c.CatID})
	if err != nil {
		return []*Dataset{}, fmt.Errorf("could not fetch datasets: %s", err)
	}
	var datasets []*Dataset
	err = cur.All(ctx, &datasets)
	if err != nil {
		return []*Dataset{}, fmt.Errorf("could not decode datasets: %s", err)
	}
	return datasets, nil
}

func SaveCatalog(c *Catalog) error {
	if c.ID.IsZero() {
		return mgm.Coll(c).Create(c)
	}
	return mgm.Coll(c).Update(c)
}

func SaveDataset(d *Dataset) error {
	if d.ID.IsZero() {
		return mgm.Coll(d).Create(d)
	}
	return mgm.Coll(d).Update(d)
}
