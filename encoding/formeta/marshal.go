package formeta

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"

	"github.com/miku/structs"
)

var ErrValueNotAllowed = errors.New("value not allowed")

func escapeSingleQuote(s string) string {
	return strings.Replace(s, "'", `\'`, -1)
}

func marshal(w io.Writer, k string, v interface{}) error {
	kind := reflect.TypeOf(v).Kind()
	switch kind {
	case reflect.Struct:
		// we want time.Time formatted, not the struct
		switch t := v.(type) {
		case time.Time:
			if _, err := io.WriteString(w, fmt.Sprintf("%s: '%s', ", k, t.Format(time.RFC3339))); err != nil {
				return err
			}
			return nil
		}

		if _, err := io.WriteString(w, k+" { "); err != nil {
			return err
		}

		for _, f := range structs.New(v).Fields() {
			if !f.IsExported() {
				continue
			}
			if err := marshal(w, f.Name(), f.Value()); err != nil {
				return err
			}
		}

		if _, err := io.WriteString(w, " } "); err != nil {
			return err
		}
	case reflect.Slice:
		vv := reflect.ValueOf(v)
		for i := 0; i < vv.Len(); i++ {
			if err := marshal(w, k, vv.Index(i).Interface()); err != nil {
				return err
			}
		}
	case reflect.String:
		s := v.(string)
		if s == "" {
			return nil
		}
		if k == "" {
			return ErrValueNotAllowed
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:

		if _, err := io.WriteString(w, fmt.Sprintf("%s: %d, ", k, v)); err != nil {
			return err
		}
	case reflect.Float32, reflect.Float64:
		if _, err := io.WriteString(w, fmt.Sprintf("%s: %f, ", k, v)); err != nil {
			return err
		}
	default:
		if _, err := io.WriteString(w, fmt.Sprintf("%s: '%v', ", k, v)); err != nil {
			return err
		}
	}
	return nil
}

// Marshal serializes a value as metafacture formeta. Partial implementation.
func Marshal(v interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := marshal(buf, "", v); err != nil {
		return []byte{}, err
	}
	return buf.Bytes(), nil
}
