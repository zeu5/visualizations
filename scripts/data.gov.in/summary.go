package datagovin

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/gosuri/uilive"
	"github.com/kamva/mgm/v3"
	"github.com/zeu5/visualizations/models/crime"
	datagovin "github.com/zeu5/visualizations/models/data.gov.in"
	"github.com/zeu5/visualizations/util"
)

func Summarise() {
	fmt.Println("Initializing...")

	err := InitializeDB(dbURL)
	if err != nil {
		log.Fatalln(err)
	}

	catalogs, err := GetAllCatalog()
	if err != nil {
		log.Fatalf("could not fetch catalogs: %s", err)
	}

	wg := util.NewCountableWaitGroup()
	writer := uilive.New()
	writer.Start()
	totCatalogs := len(catalogs)
	wg.Add(totCatalogs)
	for _, c := range catalogs {
		go func(cat *datagovin.Catalog) {
			year := getCatYear(cat.Title)
			tableEntries := make([]interface{}, 0)
			datasets, err := GetCatalogInfo(cat)
			if err != nil {
				return
			}
			for _, d := range datasets {
				tableEntries = append(tableEntries, &crime.CrimeTable{
					Title:     d.Title,
					Year:      year,
					DatasetID: d.DID,
					Data:      mapData(d),
				})
			}

			coll := mgm.Coll(&crime.CrimeTable{})
			coll.InsertMany(mgm.Ctx(), tableEntries)
			wg.Done()
			fmt.Fprintf(writer, "Pending: %d/%d\n", wg.Count(), totCatalogs)
		}(c)
	}
	wg.Wait()
	writer.Stop()
	fmt.Println("Completed!")
}

func mapData(d *datagovin.Dataset) crime.Data {
	columns := make([]string, len(d.Data.Fields))
	entries := make([][]string, len(d.Data.Entries))
	for i, field := range d.Data.Fields {
		label, ok := field["label"]
		if ok {
			columns[i] = label
		}
	}

	for i, entry := range d.Data.Entries {
		values := make([]string, len(entry))
		for j, v := range entry {
			vS, ok := v.(string)
			if !ok {
				continue
			}
			values[j] = vS
		}
		entries[i] = values
	}
	return crime.Data{
		Columns: columns,
		Entries: entries,
	}
}

func getCatYear(title string) int {
	parts := strings.Split(title, "-")
	if len(parts) != 2 {
		return -1
	}
	year, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return 0
	}
	return year
}
