package table

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"

	"github.com/360EntSecGroup-Skylar/excelize"
)

// WriteCSV writes the table as a CSV.
func WriteCSV(path string, t *TableData) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)

	if len(t.Columns) > 0 {
		err = writer.Write(t.Columns)
		if err != nil {
			return err
		}
	}
	err = writer.WriteAll(t.Records)
	if err != nil {
		return err
	}
	writer.Flush()
	return writer.Error()
}

// WriteCSVSimple writes a file with cols and records data.
func WriteCSVSimple(cols []string, records [][]string, filename string) error {
	tbl := NewTableData()
	tbl.Columns = cols
	tbl.Records = records
	return tbl.WriteCSV(filename)
}

/*
// WriteXLSX writes a table as an Excel XLSX file.
func WriteXLSX(path, sheetname string, t *TableData) error {
	t.Name = sheetname
	return WriteXLSXMore(path, []*TableData{t})
}*/

// WriteXLSX writes a table as an Excel XLSX file.
func WriteXLSX(path string, tbls []*TableData) error {
	fmt.Println(len(tbls))
	//panic("A")
	f := excelize.NewFile()
	// Create a new sheet.
	for i, t := range tbls {
		fmt.Printf("[%v][%v]\n", i, t.Name)
		sheetname := strings.TrimSpace(t.Name)
		if len(sheetname) == 0 {
			sheetname = fmt.Sprintf("Sheet %d", i)
		}
		index := f.NewSheet(sheetname)
		// Set value of a cell.
		rowBase := 0
		if len(t.Columns) > 0 {
			rowBase++
			for i, cellValue := range t.Columns {
				cellLocation := CoordinatesToSheetLocation(uint32(i), 0)
				f.SetCellValue(sheetname, cellLocation, cellValue)
			}
		}
		for y, row := range t.Records {
			for x, cellValue := range row {
				cellLocation := CoordinatesToSheetLocation(uint32(x), uint32(y+rowBase))
				f.SetCellValue(sheetname, cellLocation, cellValue)
			}
		}
		// Set active sheet of the workbook.
		if i == 0 {
			f.SetActiveSheet(index)
		}
	}
	// Save xlsx file by the given path.
	return f.SaveAs(path)
}
