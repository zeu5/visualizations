package datagovin

import "strings"

func CrimeInIndia(cats []*Catalog) []*Catalog {
	result := make([]*Catalog, 0)
	for _, c := range cats {
		if strings.Contains(c.Title, "Crime in India") {
			result = append(result, c)
		}
	}
	return result
}

func CompareCatalogs(list1 []*Catalog, list2 []*Catalog) []*Catalog {
	result := make([]*Catalog, 0)
	catMap := make(map[uint64]*Catalog)
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

func CompareDataSets(list1 []*Dataset, list2 []*Dataset) []*Dataset {
	result := make([]*Dataset, 0)
	datasetMap := make(map[uint64]*Dataset)
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
