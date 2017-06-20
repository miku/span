// Package formeta implements marshaling for formeta (metafacture internal format).
package formeta

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"

	"github.com/fatih/structs"
)

var (
	// ErrValueNotAllowed signal formeta semantic mismatch.
	ErrValueNotAllowed = errors.New("value not allowed")

	escaper = strings.NewReplacer(`\`, `\\`, "\r\n", `\n`, "\n", `\n`, "'", `\'`, "\r", " ")
)

// marshal encodes a value as formeta and writes it to the given writer.
// Toplevel object must be a struct. JSON tags are reused as keys, if defined.
func marshal(w io.Writer, k string, v interface{}) error {
	switch reflect.TypeOf(v).Kind() {
	case reflect.Struct:
		// If v is a time.Time, it will be recognized as struct as well, but we
		// really want the time formatted.
		if t, ok := v.(time.Time); ok {
			s := fmt.Sprintf("%s: '%s', ", k, t.Format(time.RFC3339))
			_, err := io.WriteString(w, s)
			return err
		}

		if _, err := io.WriteString(w, k+" { "); err != nil {
			return err
		}

		for _, f := range structs.New(v).Fields() {
			if !f.IsExported() {
				continue
			}

			var key string
			var tagv = strings.Split(f.Tag("json"), ",")

			if len(tagv) > 0 && tagv[0] != "" {
				key = tagv[0]
			} else {
				key = f.Name()
			}

			if err := marshal(w, key, f.Value()); err != nil {
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
			// No toplevel key present and value is a string.
			return ErrValueNotAllowed
		}
		ss := fmt.Sprintf("%s: '%v', ", k, escaper.Replace(s))
		if _, err := io.WriteString(w, ss); err != nil {
			return err
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if _, err := io.WriteString(w, fmt.Sprintf("%s: %d, ", k, v)); err != nil {
			return err
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
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

// Marshal serializes a value as metafacture formeta. Mostly complete, might missing some edge cases.
// Example formeta snippets:
//
//     person-1 {
//         Name: Grimm,
//         Vorname: Wilhelm,
//         Vorname: Carl
//     }
//     person-2 {
//         Name: Grimm,
//         Vorname: Jacob,
//     }
//     person-3 {
//         surname: 'Jung',
//         forename: 'Carl', forename: 'Gustav',
//         affiliation {
//             institution: 'Basel University',
//             country: 'Switzerland',
//         },
//     }
//     978-3-525-20764-2 {
//         title: Kinder- und Hausmärchen,
//         authoredById: person-1,
//         authoredById: person-2,
//         readById: person-3,
//         publicationYear: 1986,
//         publisher: Vandenhoeck und Ruprecht\: Göttingen,
//         isbn: 978-3-525-20764-2
//     }
//
func Marshal(v interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := marshal(buf, "", v); err != nil {
		return []byte{}, err
	}
	return bytes.TrimSpace(buf.Bytes()), nil
}
