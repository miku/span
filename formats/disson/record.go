package disson

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/miku/span"
	"github.com/miku/span/formats/finc"
	"github.com/miku/span/formats/marc"
)

type Record struct {
	marc.Record
}

func (r Record) Title() string {
	result := r.MustGetFirstDataField("245.a")
	subtitle := strings.TrimSpace(r.MustGetFirstDataField("245.b"))
	if subtitle != "" {
		result = result + ": " + subtitle
	}
	return result
}

func (r Record) FindYear() string {
	for _, v := range r.MustGetDataFields("502.a") {
		if _, err := strconv.Atoi(v); err == nil {
			return v
		}
	}
	for _, v := range r.MustGetDataFields("264.c") {
		if _, err := strconv.Atoi(v); err == nil {
			return v
		}
	}
	p := regexp.MustCompile(`[12][6789012][0-9]{2,2}`)
	for _, v := range r.MustGetDataFields("502.a") {
		if w := p.FindString(v); w != "" {
			return w
		}
	}
	return ""
}

func (r Record) ToIntermediateSchema() (*finc.IntermediateSchema, error) {
	output := finc.NewIntermediateSchema()
	output.SourceID = "13"
	output.RecordID = r.MustGetControlField("001")
	output.ID = fmt.Sprintf("ai-%s-%s", output.SourceID, output.RecordID)
	output.Format = "ElectronicThesis"
	output.Genre = "book"

	output.Languages = r.MustGetDataFields("041.a")
	output.URL = r.MustGetDataFields("856.u")

	for _, author := range r.MustGetDataFields("100.a") {
		output.Authors = append(output.Authors, finc.Author{Name: author})
	}
	for _, author := range r.MustGetDataFields("700.a") {
		output.Authors = append(output.Authors, finc.Author{Name: author})
	}

	output.Abstract = r.MustGetFirstDataField("520.a")
	output.ArticleTitle = r.Title()

	for _, p := range r.MustGetDataFields("264.a") {
		output.Places = append(output.Places, p)
	}
	for _, p := range r.MustGetDataFields("264.b") {
		output.Publishers = append(output.Publishers, p)
	}

	year := r.FindYear()
	if year == "" {
		return output, span.Skip{Reason: fmt.Sprintf("no year found in %s", output.RecordID)}
	}
	output.RawDate = fmt.Sprintf("%s-01-01", year)
	t, err := time.Parse("2006-01-02", output.RawDate)
	if err != nil {
		return output, err
	}
	output.Date = t

	for _, v := range r.MustGetDataFields("650.a") {
		for _, w := range strings.Split(v, ",") {
			w = strings.TrimSpace(w)
			output.Subjects = append(output.Subjects, w)
		}
	}
	for _, v := range r.MustGetDataFields("653.a") {
		for _, w := range strings.Split(v, ",") {
			w = strings.TrimSpace(w)
			output.Subjects = append(output.Subjects, w)
		}
	}
	for _, v := range r.MustGetDataFields("689.a") {
		for _, w := range strings.Split(v, ",") {
			w = strings.TrimSpace(w)
			output.Subjects = append(output.Subjects, w)
		}
	}

	return output, nil
}
