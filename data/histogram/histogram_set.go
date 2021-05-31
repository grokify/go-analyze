package histogram

import (
	"sort"
	"strings"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/grokify/gocharts/data/excelizeutil"
	"github.com/grokify/gocharts/data/statictimeseries"
	"github.com/grokify/gocharts/data/table"
	"github.com/grokify/simplego/time/timeutil"
)

type HistogramSet struct {
	Name         string
	HistogramMap map[string]*Histogram
	KeyIsTime    bool
}

func NewHistogramSet(name string) *HistogramSet {
	return &HistogramSet{
		Name:         name,
		HistogramMap: map[string]*Histogram{}}
}

func NewHistogramSetWithData(name string, data map[string]map[string]int) *HistogramSet {
	hset := &HistogramSet{
		Name:         name,
		HistogramMap: map[string]*Histogram{}}
	for statsName, statsData := range data {
		for statsItemName, statsItemValue := range statsData {
			hset.Add(statsName, statsItemName, statsItemValue)
		}
	}
	return hset
}

func (hset *HistogramSet) AddDateUidCount(dt time.Time, uid string, count int) {
	fName := dt.Format(time.RFC3339)
	hset.Add(fName, uid, count)
	if !hset.KeyIsTime {
		hset.KeyIsTime = true
	}
}

func (hset *HistogramSet) Add(setName, binName string, count int) {
	fstats, ok := hset.HistogramMap[setName]
	if !ok {
		fstats = NewHistogram(setName)
	}
	fstats.Add(binName, count)
	hset.HistogramMap[setName] = fstats
}

func (hset *HistogramSet) ItemCount() uint {
	return uint(len(hset.HistogramMap))
}

/*
func (hset *HistogramSet) AddString(frequencyName, itemName string) {
	fstats, ok := hset.HistogramMap[frequencyName]
	if !ok {
		fstats = NewHistogram(frequencyName)
	}
	fstats.Add(itemName, 1)
	hset.HistogramMap[frequencyName] = fstats
}*/

func (hset *HistogramSet) HistogramNames() []string {
	names := []string{}
	for name := range hset.HistogramMap {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (hset *HistogramSet) HistogramNameExists(histName string) bool {
	if _, ok := hset.HistogramMap[histName]; ok {
		return true
	}
	return false
}

func (hset *HistogramSet) TotalCount() uint64 {
	totalCount := uint64(0)
	for _, fstats := range hset.HistogramMap {
		totalCount += fstats.TotalCount()
	}
	return totalCount
}

func (hset *HistogramSet) HistogramBinNames(setName string) []string {
	if hist, ok := hset.HistogramMap[setName]; ok {
		return hist.BinNames()
	}
	return []string{}
}

func (hset *HistogramSet) LeafStats(name string) *Histogram {
	if len(name) == 0 {
		name = "leaf stats"
	}
	setLeafStats := NewHistogram(name)
	for _, hist := range hset.HistogramMap {
		for binName, binCount := range hist.Bins {
			setLeafStats.Add(binName, binCount)
		}
	}
	return setLeafStats
}

func (hset *HistogramSet) ToDataSeriesDistinct() (statictimeseries.DataSeries, error) {
	ds := statictimeseries.NewDataSeries()
	ds.SeriesName = hset.Name
	for rfc3339, hist := range hset.HistogramMap {
		dt, err := time.Parse(time.RFC3339, rfc3339)
		if err != nil {
			return ds, err
		}
		ds.AddItem(statictimeseries.DataItem{
			SeriesName: hset.Name,
			Time:       dt,
			Value:      int64(len(hist.Bins))})
	}
	return ds, nil
}

func (hset *HistogramSet) WriteXLSX(path, colName1, colName2, colNameCount string) error {
	// WriteXLSX writes a table as an Excel XLSX file with
	// row formatter option.
	f := excelize.NewFile()
	// Create a new sheet.

	sheetName := strings.TrimSpace(hset.Name)
	if len(sheetName) == 0 {
		sheetName = "Sheet0"
	}
	index := f.NewSheet(sheetName)

	colName1 = strings.TrimSpace(colName1)
	if len(colName1) == 0 {
		colName1 = hset.Name
	}
	if len(colName1) == 0 {
		colName1 = "Column1"
	}
	colName2 = strings.TrimSpace(colName2)
	if len(colName1) == 0 {
		for _, fstats := range hset.HistogramMap {
			fstats.Name = strings.TrimSpace(fstats.Name)
			if len(fstats.Name) > 0 {
				colName2 = fstats.Name
				break
			}
		}
	}
	colNameCount = strings.TrimSpace(colNameCount)
	if len(colNameCount) == 0 {
		colNameCount = "Count"
	}
	header := []interface{}{colName1, colName2, colNameCount}

	excelizeutil.SetRowValues(f, sheetName, 0, header)
	var err error
	rowIdx := uint(1)
	for fstatsName, fstats := range hset.HistogramMap {
		fstatsNameDt := time.Now()
		if hset.KeyIsTime {
			fstatsNameDt, err = time.Parse(time.RFC3339, fstatsName)
			if err != nil {
				return err
			}
		}
		for binName, binCount := range fstats.Bins {
			var rowVals []interface{}
			if hset.KeyIsTime {
				rowVals = []interface{}{fstatsNameDt, binName, binCount}
			} else {
				rowVals = []interface{}{fstatsName, binName, binCount}
			}
			excelizeutil.SetRowValues(f, sheetName, rowIdx, rowVals)
			rowIdx++
		}
	}
	f.SetActiveSheet(index)
	// Delete Original Sheet
	f.DeleteSheet(f.GetSheetName(0))
	// Save xlsx file by the given path.
	return f.SaveAs(path)

}

// HistogramSetDatetimeToQuarter converts a HistogramSet
// by date to one by quarter.s.
func HistogramSetDatetimeToQuarter(name string, fsetIn *HistogramSet) (*HistogramSet, error) {
	fsetQtr := NewHistogramSet(name)
	for rfc3339, hist := range fsetIn.HistogramMap {
		dt, err := time.Parse(time.RFC3339, rfc3339)
		if err != nil {
			return fsetQtr, err
		}
		dt = timeutil.QuarterStart(dt)
		rfc3339Qtr := dt.Format(time.RFC3339)
		for binName, binCount := range hist.Bins {
			fsetQtr.Add(rfc3339Qtr, binName, binCount)
		}
	}
	return fsetQtr, nil
}

// HistogramSetTimeKeyCount returns a DataSeries when
// the first key is a RFC3339 time and a sum of items
// is desired per time.
func HistogramSetTimeKeyCount(hset HistogramSet) (statictimeseries.DataSeries, error) {
	ds := statictimeseries.NewDataSeries()
	ds.SeriesName = hset.Name
	for rfc3339, hist := range hset.HistogramMap {
		dt, err := time.Parse(time.RFC3339, rfc3339)
		if err != nil {
			return ds, err
		}
		ds.AddItem(statictimeseries.DataItem{
			SeriesName: hset.Name,
			Time:       dt,
			Value:      int64(len(hist.Bins))})
	}
	return ds, nil
}

func HistogramSetTimeKeyCountTable(hset HistogramSet, interval timeutil.Interval, countColName string) (table.Table, error) {
	ds, err := HistogramSetTimeKeyCount(hset)
	if err != nil {
		return table.NewTable(), err
	}
	ds.Interval = interval
	countColName = strings.TrimSpace(countColName)
	if len(countColName) == 0 {
		countColName = "Count"
	}
	return statictimeseries.DataSeriesToTable(ds, countColName, statictimeseries.TimeFormatRFC3339), nil
}

func HistogramSetTimeKeyCountWriteXLSX(filename string, hset HistogramSet, interval timeutil.Interval, countColName string) error {
	tbl, err := HistogramSetTimeKeyCountTable(hset, interval, countColName)
	if err != nil {
		return err
	}
	tbl.FormatFunc = table.FormatTimeAndInts
	return table.WriteXLSX(filename, &tbl)
}
