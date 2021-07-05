package datagovin

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/zeu5/visualizations/util"
)

const (
	BaseURL       = "https://data.gov.in"
	CatalogURL    = BaseURL + "/cust-api/v1"
	DatasetURL    = CatalogURL + "/resource"
	TokenURL      = BaseURL + "/save_reasons"
	DataURL       = BaseURL + "/node"
	DataURLSuffix = "/datastore/export/json"
)

type requests struct {
	network *util.ThrottledClient
}

func noTimeout(c *http.Client) {
	c.Timeout = time.Duration(0)
}

func newRequests() *requests {
	return &requests{
		network: util.NewThrottledClient(200*time.Millisecond, 10, noTimeout),
	}
}

func (r *requests) Start() {
	r.network.Start()
}

func (r *requests) Stop() {
	r.network.Stop()
}

type response struct {
	Status  string                   `json:"status"`
	Records []map[string]interface{} `json:"records"`
	Total   int                      `json:"total"`
	Count   int                      `json:"count"`
}

func (r *requests) FetchCatalogs() ([]*Catalog, error) {

	query := make(url.Values)
	query.Add("format", "json")
	query.Add("offset", "0")
	query.Add("limit", "7000")
	query.Add("sort[_score]", "desc")
	query.Add("query", "Crime in India")
	request, err := http.NewRequest("GET", CatalogURL+"/?"+query.Encode(), nil)
	if err != nil {
		return []*Catalog{}, err
	}

	reqRes, err := r.network.Do(request)
	if err != nil {
		return nil, err
	}
	select {
	case resp := <-reqRes.Response:
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return []*Catalog{}, fmt.Errorf("could not read response: %s", err.Error())
		}
		var respS response
		err = json.Unmarshal(body, &respS)
		if err != nil {
			return []*Catalog{}, fmt.Errorf("could not parse response: %s", err.Error())
		}
		if respS.Status != "ok" {
			return []*Catalog{}, errors.New("response status not ok")
		}
		return parseCatalogs(respS.Records)
	case err := <-reqRes.Error:
		return []*Catalog{}, err
	}
}

type catalogRecord struct {
	ID          uint64                 `mapstructure:"id"`
	Title       string                 `mapstructure:"title"`
	Departments []string               `mapstructure:"field_ministry_department:name"`
	Created     string                 `mapstructure:"created"`
	Changed     string                 `mapstructure:"changed"`
	Rest        map[string]interface{} `mapstructure:",remain"`
}

func parseCatalogs(records []map[string]interface{}) ([]*Catalog, error) {
	catalogs := make([]*Catalog, 0)
	for _, record := range records {
		var r catalogRecord
		err := mapstructure.Decode(record, &r)
		if err != nil {
			continue
		}
		createdI, _ := strconv.ParseInt(r.Created, 10, 64)
		lastModifiedI, _ := strconv.ParseInt(r.Changed, 10, 64)

		catalogs = append(catalogs, &Catalog{
			Title:        r.Title,
			Created:      time.Unix(createdI, 0),
			LastModified: time.Unix(lastModifiedI, 0),
			CatID:        r.ID,
			Departments:  r.Departments,
			Other:        r.Rest,
		})

	}
	return catalogs, nil
}

func (r *requests) FetchCatalogInfo(c *Catalog) ([]*Dataset, error) {
	query := make(url.Values)
	query.Set("filters[field_catalog_reference]", strconv.FormatUint(c.CatID, 10))
	query.Set("format", "json")
	query.Set("sort[created]", "desc")

	totLen := -1
	count := 0

	responses := make([]*response, 0)
	for {
		if totLen == count {
			break
		}

		request, err := http.NewRequest("GET", DatasetURL+"/?"+query.Encode(), nil)
		if err != nil {
			return []*Dataset{}, err
		}
		reqRes, err := r.network.Do(request)
		if err != nil {
			return []*Dataset{}, err
		}
		select {
		case resp := <-reqRes.Response:
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				resp.Body.Close()
				return []*Dataset{}, fmt.Errorf("failed to read response body: %s", err.Error())
			}
			rS := new(response)
			err = json.Unmarshal(body, rS)
			if err != nil {
				resp.Body.Close()
				return []*Dataset{}, fmt.Errorf("failed to parse response body: %s", err.Error())
			}
			responses = append(responses, rS)
			totLen = rS.Total
			count = count + rS.Count

			query.Set("offset", strconv.Itoa(count))
			query.Set("limit", strconv.Itoa(totLen))
			resp.Body.Close()
		case err := <-reqRes.Error:
			return []*Dataset{}, err
		}
	}
	datasetRecords := make([]map[string]interface{}, 0)
	for _, resp := range responses {
		datasetRecords = append(datasetRecords, resp.Records...)
	}
	return parseDatasets(c, datasetRecords)
}

type datasetRecord struct {
	ID      uint64                 `mapstructure:"id"`
	Changed string                 `mapstructure:"changed"`
	Created string                 `mapstructure:"created"`
	Title   string                 `mapstructure:"title"`
	Rest    map[string]interface{} `mapstructure:",remain"`
}

func parseDatasets(c *Catalog, records []map[string]interface{}) ([]*Dataset, error) {
	datasets := make([]*Dataset, 0)
	for _, record := range records {
		var r datasetRecord
		err := mapstructure.Decode(record, &r)
		if err != nil {
			continue
		}
		createdI, _ := strconv.ParseInt(r.Created, 10, 64)
		lastModifiedI, _ := strconv.ParseInt(r.Changed, 10, 64)

		datasets = append(datasets, &Dataset{
			DID:          r.ID,
			CatID:        c.CatID,
			Created:      time.Unix(createdI, 0),
			LastModified: time.Unix(lastModifiedI, 0),
			Title:        r.Title,
			Other:        r.Rest,
		})
	}
	return datasets, nil
}

type dataRequestGroup struct {
	token   string
	data    []byte
	dataset *Dataset
}

type dataResponse struct {
	Fields []map[string]string `json:"fields"`
	Data   [][]interface{}     `json:"data"`
}

func (d *dataRequestGroup) Next(resp *http.Response) *http.Request {
	if resp == nil {
		data := url.Values{}
		data.Set("reasons", `{"download_reasons":"2"}`)
		data.Set("reasons1", `{"reasons_d[3]":"3","reasons_d[4]":"4"}`)
		data.Set("reasons2", `{"node_id":"`+strconv.FormatUint(d.dataset.DID, 10)+`","file_for_mat":"xls","redirect_url":"/catalog/crime-india-2016","name_d":"","mail_d":"","form_id":"download_confirmation_resources_form"}`)

		tokenReq, err := http.NewRequest("POST", TokenURL, strings.NewReader(data.Encode()))
		tokenReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		if err != nil {
			return nil
		}
		return tokenReq
	}
	if d.token == "" {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			resp.Body.Close()
			return nil
		}
		d.token = string(body)

		dataReq, err := http.NewRequest(
			"GET",
			DataURL+"/"+strconv.FormatUint(d.dataset.DID, 10)+DataURLSuffix+"/?token="+d.token,
			nil,
		)
		resp.Body.Close()
		if err != nil {
			return nil
		}
		return dataReq
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err == nil {
		d.data = body
	}
	return nil
}

func (r *requests) FetchData(d *Dataset) (*Data, error) {
	reqGroup := &dataRequestGroup{
		token:   "",
		data:    []byte{},
		dataset: d,
	}
	reqRes, err := r.network.DoGroup(reqGroup)
	if err != nil {
		return nil, fmt.Errorf("failed to submit data req: %s", err)
	}
	select {
	case <-reqRes.ResponseGroup:
		if bytes.Equal(reqGroup.data, []byte{}) {
			return nil, errors.New("no data fetched")
		}
		var dataResp dataResponse
		err := json.Unmarshal(reqGroup.data, &dataResp)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshall data: %s", err)
		}
		return &Data{
			Fields:  dataResp.Fields,
			Entries: dataResp.Data,
		}, nil
	case err := <-reqRes.Error:
		return nil, fmt.Errorf("failed to fetch data: %s", err)
	}
}
