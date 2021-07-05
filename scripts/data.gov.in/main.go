package datagovin

import (
	"fmt"
	"log"
	"sync"

	"github.com/gosuri/uilive"
	"github.com/spf13/cobra"
	"github.com/zeu5/visualizations/util"
)

var (
	dbURL string
)

type datasetColl struct {
	datasets []*Dataset
	lock     *sync.Mutex
}

func newDatasetColl() *datasetColl {
	return &datasetColl{
		datasets: make([]*Dataset, 0),
		lock:     new(sync.Mutex),
	}
}

func (d *datasetColl) Insert(dSet *Dataset) {
	d.lock.Lock()
	d.datasets = append(d.datasets, dSet)
	d.lock.Unlock()
}

func (d *datasetColl) Append(dColl ...*Dataset) {
	d.lock.Lock()
	d.datasets = append(d.datasets, dColl...)
	d.lock.Unlock()
}

func (d *datasetColl) Iter() []*Dataset {
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

func Run() {
	fmt.Println("Initializing...")
	writer := uilive.New()
	requests := newRequests()

	prog := &progress{
		mtx: new(sync.Mutex),
	}
	writer.Start()
	requests.Start()

	err := InitializeDB()
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
		go func(cat *Catalog) {
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
		go func(dat *Dataset) {
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

func CrimeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "crime",
		Short: "Fetch Crime records from data.gov.in",
		Run: func(cmd *cobra.Command, args []string) {
			Run()
		},
	}
	cmd.PersistentFlags().StringVar(&dbURL, "Mongo db url", "mondogdb://localhost:27017", "MongoDB URI")
	return cmd
}
