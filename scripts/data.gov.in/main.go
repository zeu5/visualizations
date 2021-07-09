package datagovin

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"sync"

	"github.com/gosuri/uilive"
	datagovin "github.com/zeu5/visualizations/models/data.gov.in"
	"github.com/zeu5/visualizations/util"
)

var (
	dbURL    string
	dumpPath string
)

type datasetColl struct {
	datasets []*datagovin.Dataset
	lock     *sync.Mutex
}

func newDatasetColl() *datasetColl {
	return &datasetColl{
		datasets: make([]*datagovin.Dataset, 0),
		lock:     new(sync.Mutex),
	}
}

func (d *datasetColl) Insert(dSet *datagovin.Dataset) {
	d.lock.Lock()
	d.datasets = append(d.datasets, dSet)
	d.lock.Unlock()
}

func (d *datasetColl) Append(dColl ...*datagovin.Dataset) {
	d.lock.Lock()
	d.datasets = append(d.datasets, dColl...)
	d.lock.Unlock()
}

func (d *datasetColl) Iter() []*datagovin.Dataset {
	d.lock.Lock()
	i := d.datasets
	d.lock.Unlock()
	return i
}

func (d *datasetColl) Size() int {
	d.lock.Lock()
	defer d.lock.Unlock()
	return len(d.datasets)
}

type progress struct {
	totCatalogs int
	failedCat   int
	totDat      int
	failedDat   int

	mtx *sync.Mutex
}

func (p *progress) AddFailedCat() {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	p.failedCat = p.failedCat + 1
}

func (p *progress) AddFailedDat() {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	p.failedDat = p.failedDat + 1
}

func Fetch() {
	fmt.Println("Initializing...")
	writer := uilive.New()
	requests := newRequests()

	prog := &progress{
		mtx: new(sync.Mutex),
	}
	writer.Start()
	requests.Start()

	err := InitializeDB(dbURL)
	if err != nil {
		log.Fatalf("Could not initialize to database: %s\n", err)
	}
	existingCatalogs, err := GetAllCatalog()
	if err != nil {
		log.Fatalf("Failed to fetch data from database: %s", err)
	}

	catalogs, err := requests.FetchCatalogs()
	if err != nil {
		log.Fatalf("Failed to request Catalog info from data.gov.in: %s\n", err.Error())
	}
	// Fetching only crime in india for now. Too many otherwise
	catalogs = CrimeInIndia(catalogs)

	newCatalgos := CompareCatalogs(catalogs, existingCatalogs)
	newCatLen := len(newCatalgos)
	newDatasets := newDatasetColl()
	fmt.Printf("Found %d new catalogs\n", newCatLen)

	prog.totCatalogs = newCatLen

	wg := util.NewCountableWaitGroup()
	wg.Add(newCatLen)
	for _, c := range newCatalgos {
		go func(cat *datagovin.Catalog) {
			err := SaveCatalog(cat)
			if err == nil {
				existingDatasets, err := GetCatalogInfo(cat)
				if err == nil {
					datasets, err := requests.FetchCatalogInfo(cat)
					if err == nil {
						newDatasets.Append(CompareDataSets(datasets, existingDatasets)...)
					} else {
						prog.AddFailedCat()
					}
				}
			}
			wg.Done()
			fmt.Fprintf(writer, "Pending downloads: %d/%d catalogs\n", wg.Count(), newCatLen)
		}(c)
	}
	wg.Wait()
	writer.Stop()

	if prog.failedCat != 0 {
		fmt.Printf("Failed to fetch %d catalogs\n", prog.failedCat)
	}

	newDatasetsSize := newDatasets.Size()
	fmt.Printf("Fetching data of %d datasets\n", newDatasetsSize)

	prog.totDat = newDatasetsSize

	writer = uilive.New()
	writer.Start()
	wg.Add(newDatasetsSize)
	for _, d := range newDatasets.Iter() {
		go func(dat *datagovin.Dataset) {
			data, err := requests.FetchData(dat)
			if err == nil {
				dat.Data = *data
				SaveDataset(dat)
			} else {
				prog.AddFailedDat()
			}
			wg.Done()
			fmt.Fprintf(writer, "Pending downloads: %d/%d datasets\n", wg.Count(), newDatasetsSize)
		}(d)
	}
	wg.Wait()
	writer.Stop()
	requests.Stop()

	if prog.failedDat != 0 {
		fmt.Printf("Failed to fetch %d datasets\n", prog.failedDat)
	}

	fmt.Println("Completed!")
}

func createFiles() (*os.File, error) {
	dir, err := os.Stat(dumpPath)
	if err != nil {
		if os.IsNotExist(err) {
			err := os.Mkdir(dumpPath, os.ModePerm)
			if err != nil {
				return nil, fmt.Errorf("could not create dump dir: %s", err)
			}
			dir, _ = os.Stat(dumpPath)
		} else {
			return nil, fmt.Errorf("could not read path: %s", err)
		}
	}
	if !dir.IsDir() {
		return nil, errors.New("path is not a directory")
	}
	file, err := os.Create(path.Join(dumpPath, "data.json"))
	if err != nil {
		return nil, fmt.Errorf("could not create dump file: %s", err)
	}
	return file, nil
}

func Dump() {
	fmt.Println("Initializing...")
	dumpFile, err := createFiles()
	if err != nil {
		log.Fatalln(err.Error())
	}
	writer := bufio.NewWriter(dumpFile)

	err = InitializeDB(dbURL)
	if err != nil {
		log.Fatalln(err.Error())
	}

	fmt.Println("Fetching data...")
	data := make(map[string]interface{})
	catalogs, err := GetAllCatalog()
	if err != nil {
		log.Fatalln(err.Error())
	}
	data["catalogs"] = catalogs
	datasets := make([]*datagovin.Dataset, 0)

	for _, c := range catalogs {
		ds, err := GetCatalogInfo(c)
		if err != nil {
			continue
		}
		datasets = append(datasets, ds...)
	}
	data["datasets"] = datasets

	fmt.Println("Writing data...")
	dataS, err := json.Marshal(&data)
	if err != nil {
		log.Fatalln("Failed to marshall data")
	}
	_, err = writer.Write(dataS)
	if err != nil {
		log.Fatalln("Failed to write data to file")
	}
	writer.Flush()
	dumpFile.Close()
	fmt.Println("Completed!")
}
