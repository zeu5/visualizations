package datagovin

import (
	"fmt"

	"github.com/kamva/mgm/v3"
	datagovin "github.com/zeu5/visualizations/models/data.gov.in"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func InitializeDB(url string) error {
	return mgm.SetDefaultConfig(nil, "vis", options.Client().ApplyURI(url))
}

func GetAllCatalog() ([]*datagovin.Catalog, error) {
	coll := mgm.Coll(&datagovin.Catalog{})
	ctx := mgm.Ctx()
	var catalogs []*datagovin.Catalog

	cur, err := coll.Find(ctx, bson.M{})
	if err != nil {
		return []*datagovin.Catalog{}, fmt.Errorf("could not fetch from db: %s", err)
	}
	err = cur.All(ctx, &catalogs)
	if err != nil {
		return []*datagovin.Catalog{}, fmt.Errorf("could not decode records: %s", err)
	}

	return catalogs, nil
}

func GetCatalogInfo(c *datagovin.Catalog) ([]*datagovin.Dataset, error) {
	coll := mgm.Coll(&datagovin.Dataset{})
	ctx := mgm.Ctx()
	cur, err := coll.Find(ctx, bson.M{"cat_id": c.CatID})
	if err != nil {
		return []*datagovin.Dataset{}, fmt.Errorf("could not fetch datasets: %s", err)
	}
	var datasets []*datagovin.Dataset
	err = cur.All(ctx, &datasets)
	if err != nil {
		return []*datagovin.Dataset{}, fmt.Errorf("could not decode datasets: %s", err)
	}
	return datasets, nil
}

func SaveCatalog(c *datagovin.Catalog) error {
	if c.ID.IsZero() {
		return mgm.Coll(c).Create(c)
	}
	return mgm.Coll(c).Update(c)
}

func SaveDataset(d *datagovin.Dataset) error {
	if d.ID.IsZero() {
		return mgm.Coll(d).Create(d)
	}
	return mgm.Coll(d).Update(d)
}
