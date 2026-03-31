package output

import (
	"encoding/json"
	"fmt"
	"io"
)

func PrintJSON(w io.Writer, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal json: %w", err)
	}

	_, err = fmt.Fprintln(w, string(data))
	return err
}
