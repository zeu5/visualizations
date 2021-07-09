package datagovin

import (
	"time"

	"github.com/kamva/mgm/v3"
)

type Catalog struct {
	mgm.DefaultModel `json:"-"`
	Title            string                 `json:"title" bson:"title"`
	Created          time.Time              `json:"created" bson:"created"`
	LastModified     time.Time              `json:"last_modified" bson:"last_modified"`
	CatID            uint64                 `json:"cat_id" bson:"cat_id"`
	Departments      []string               `json:"department" bson:"department"`
	Other            map[string]interface{} `json:"other" bson:"other"`
}

func (c *Catalog) CollectionName() string {
	return "data_gov_in_catalogs"
}

type Dataset struct {
	mgm.DefaultModel `json:"-"`
	DID              uint64                 `json:"d_id" bson:"d_id"`
	Title            string                 `json:"title" bson:"title"`
	CatID            uint64                 `json:"cat_id" bson:"cat_id"`
	Created          time.Time              `json:"created" bson:"created"`
	LastModified     time.Time              `json:"last_modified" bson:"last_modified"`
	Other            map[string]interface{} `json:"other" bson:"other"`
	Data             Data                   `json:"data" bson:"data"`
}

func (d *Dataset) CollectionName() string {
	return "data_gov_in_datasets"
}

type Data struct {
	Fields  []map[string]string `json:"fields" bson:"fields"`
	Entries [][]interface{}     `json:"entries" bson:"entries"`
}
