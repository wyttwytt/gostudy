package lookupcache

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/lnshi/json-lookup/tool/json"
)

//LookupCache is an interface your API should implement
type LookupCache interface {
	GetSegmentForOrgAndKey(orgKey string, paramKey string) []SegmentConfig
	GetSegmentForOrgAndKeyAndVal(orgKey string, paramKey string, paramVal string) []SegmentConfig
}

//SegmentConfig is a struct that holds an id for 1 segment
type SegmentConfig struct {
	Id string
}

func (cfg *SegmentConfig) GetId() string {
	return cfg.Id
}

// Starting Leonard's code

// `data` is the raw bytes slice which represents the very original data in memory.
var data []byte

type ParamSeg struct {
	ParamVal string
	SegId    string
}

type Cache map[string]map[string][]*ParamSeg

var Ec Cache

var lock = sync.RWMutex{}

func init() {
	Ec = make(map[string]map[string][]*ParamSeg)

	dataFilePath := filepath.Join(os.Getenv("GOPATH"), "src/github.com/lnshi/json-lookup/data/data.json")

	if res, err := ioutil.ReadFile(dataFilePath); err != nil {
		panic(err)
	} else {
		data = res
	}
}

func (ec Cache) GetSegmentForOrgAndKey(orgKey string, paramKey string) []SegmentConfig {
	return ec.GetSegmentForOrgAndKeyAndVal(orgKey, paramKey, "")
}

func (ec Cache) GetSegmentForOrgAndKeyAndVal(orgKey string, paramKey string, paramVal string) []SegmentConfig {
	if orgKey == "" || paramKey == "" {
		return []SegmentConfig{}
	}

	// The data for this `orgKey` has already been parsed, we just need to search in `Ec`.
	if paramMap, ok := Ec[orgKey]; ok {
		// Found segs with this `paramKey`.
		if segs, ok := paramMap[paramKey]; ok {

			res := make([]SegmentConfig, 0)

			for _, paramSeg := range segs {
				if paramSeg.ParamVal == paramVal {
					res = append(res, SegmentConfig{Id: paramSeg.SegId})
				} else if paramVal != "" && strings.Index(paramSeg.ParamVal, paramVal) != -1 {
					res = append(res, SegmentConfig{Id: paramSeg.SegId})
				}
			}
			return res
		} else {
			// No this `paramKey`.
			return []SegmentConfig{}
		}
	} else {
		// We need to parse data for this `orgKey` from the raw bytes `data`.

		chOrgs := make(chan *json.V)
		go json.IterateArray(chOrgs, data)

		var wg sync.WaitGroup

		// Before there is org object returned we can do nothing.
		for org := range chOrgs {
			if org.Err != nil {
				return []SegmentConfig{}
			}

			chOrgDetails := make(chan *json.Kv)
			go json.IterateObject(chOrgDetails, org.V)

			// Before knowing the `orgKey` we can do nothing.
			for orgDetail := range chOrgDetails {
				if orgDetail.Err != nil {
					return []SegmentConfig{}
				}

				currOrgKey := string(orgDetail.K)

				if _, ok := Ec[currOrgKey]; ok {
					continue
				}

				paramSegMap := make(map[string][]*ParamSeg)

				// Added `orgDetail.V` to `Ec`.
				wg.Add(1)
				go addedToEc(&wg, currOrgKey, paramSegMap, orgDetail.V)

				// This is the org current call is looking.
				if currOrgKey == orgKey {
					// Haven't reached end yet, close the channel deliberately to terminate the orgs iterator goroutine.
					if _, ok := <-chOrgs; ok {
						// Will cause race condition, but it is fine.
						close(chOrgs)
					}
				}
			}
		}

		wg.Wait()

		return Ec.GetSegmentForOrgAndKeyAndVal(orgKey, paramKey, paramVal)
	}
}

func addedToEc(wg *sync.WaitGroup, orgKey string, paramSegMap map[string][]*ParamSeg, orgDetails []byte) {
	defer wg.Done()

	chParam := make(chan *json.V)
	go json.IterateArray(chParam, orgDetails)

	for param := range chParam {
		// Parse param object.
		wg.Add(1)
		go parseParamObj(wg, orgKey, paramSegMap, param.V)
	}
}

func parseParamObj(wg *sync.WaitGroup, orgKey string, paramSegMap map[string][]*ParamSeg, paramObj []byte) {
	defer wg.Done()

	chParamDetails := make(chan *json.Kv)
	go json.IterateObject(chParamDetails, paramObj)

	for paramDetail := range chParamDetails {
		// Parse segs array.
		segs := make([]*ParamSeg, 0)

		paramName := string(paramDetail.K)

		lock.Lock()
		paramSegMap[paramName] = segs
		lock.Unlock()

		wg.Add(1)
		go parseSegArr(wg, orgKey, paramName, paramSegMap, paramDetail.V)
	}
}

func parseSegArr(wg *sync.WaitGroup, orgKey string, paramName string, paramSegMap map[string][]*ParamSeg, segsArr []byte) {
	defer wg.Done()

	chSegs := make(chan *json.V)
	go json.IterateArray(chSegs, segsArr)

	for seg := range chSegs {
		segInStr := string(seg.V)

		firstQuotMark := strings.Index(segInStr, `"`)
		secondQuotMark := strings.Index(segInStr[firstQuotMark+1:], `"`)
		paramVal := segInStr[firstQuotMark+1 : firstQuotMark+secondQuotMark+1]

		lastQuotMark := strings.LastIndex(segInStr, `"`)
		secondLastQuotMark := strings.LastIndex(segInStr[:lastQuotMark], `"`)
		segId := segInStr[secondLastQuotMark+1 : lastQuotMark]

		lock.Lock()
		paramSegMap[paramName] = append(paramSegMap[paramName], &ParamSeg{ParamVal: paramVal, SegId: segId})
		lock.Unlock()
	}

	lock.Lock()
	Ec[orgKey] = paramSegMap
	lock.Unlock()
}
