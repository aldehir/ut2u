package query

import (
	"encoding/json"
	"fmt"
	"os"
)

type JSONFormatter struct {
	Reports []Server
}

func (f *JSONFormatter) Report(rpt Server) error {
	f.Reports = append(f.Reports, rpt)
	return nil
}

func (f *JSONFormatter) Flush() error {
	var doc struct {
		Reports []Server `json:"servers"`
	}

	doc.Reports = f.Reports

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")

	err := encoder.Encode(doc)
	if err != nil {
		return err
	}

	fmt.Fprintln(os.Stdout)
	return nil
}
