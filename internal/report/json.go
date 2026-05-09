package report

import (
	"encoding/json"
	"io"

	"querywise/pkg/types"
)

// WriteJSON encodes report as indented JSON.
func WriteJSON(w io.Writer, rep types.Report) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(rep); err != nil {
		return err
	}
	return nil
}
