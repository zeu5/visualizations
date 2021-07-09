package datagovin

import (
	"strings"

	datagovin "github.com/zeu5/visualizations/models/data.gov.in"
)

func CrimeInIndia(cats []*datagovin.Catalog) []*datagovin.Catalog {
	result := make([]*datagovin.Catalog, 0)
	for _, c := range cats {
		if strings.Contains(c.Title, "Crime in India") {
			result = append(result, c)
		}
	}
	return result
}

func CompareCatalogs(list1 []*datagovin.Catalog, list2 []*datagovin.Catalog) []*datagovin.Catalog {
	result := make([]*datagovin.Catalog, 0)
	catMap := make(map[uint64]*datagovin.Catalog)
	for _, c := range list2 {
		catMap[c.CatID] = c
	}
	for _, c := range list1 {
		existing, ok := catMap[c.CatID]
		if ok && c.LastModified.After(existing.LastModified) {
			existing.LastModified = c.LastModified
			existing.Title = c.Title
			existing.Departments = c.Departments
			existing.Other = c.Other
			result = append(result, existing)
		} else if !ok {
			result = append(result, c)
		}
	}
	return result
}

func CompareDataSets(list1 []*datagovin.Dataset, list2 []*datagovin.Dataset) []*datagovin.Dataset {
	result := make([]*datagovin.Dataset, 0)
	datasetMap := make(map[uint64]*datagovin.Dataset)
	for _, d := range list2 {
		datasetMap[d.DID] = d
	}
	for _, d := range list1 {
		existing, ok := datasetMap[d.DID]
		if ok && d.LastModified.After(existing.LastModified) {
			existing.Title = d.Title
			existing.LastModified = d.LastModified
			existing.Other = d.Other
			result = append(result, existing)
		} else if !ok {
			result = append(result, d)
		}
	}
	return result
}
