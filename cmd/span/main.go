package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
)

// Author is given by family and given name
type Author struct {
	Family string `json:"family"`
	Given  string `json:"given"`
}

// String pretty print the author
func (author *Author) String() string {
	if author.Given != "" {
		return fmt.Sprintf("%s, %s", author.Family, author.Given)
	} else {
		return author.Family
	}
}

// DatePart consists of up to three int, representing year, month, day
type DatePart []int

// Document is a example API response
type Document struct {
	Prefix         string     `json:"prefix"`
	Type           string     `json:"type"`
	Volume         string     `json:"volume"`
	Deposited      []DatePart `json:"deposited"`
	Source         string     `json:"source"`
	Authors        []Author   `json:"author"`
	Score          int        `json:"score"`
	Page           string     `json:"page"`
	Subject        []string   `json:"subject"`
	Title          []string   `json:"title"`
	Publisher      string     `json:"publisher"`
	ISSN           []string   `json:"ISSN"`
	Indexed        []DatePart `json:"indexed"`
	Issued         []DatePart `json:"issued"`
	Subtitle       []string   `json:"subtitle"`
	URL            string     `json:"URL"`
	Issue          string     `json:"issue"`
	ContainerTitle []string   `json:"container-title"`
	ReferenceCount int        `json:"reference-count"`
	DOI            string     `json:"DOI"`
}

// CombinedTitle returns a longish title
func (d *Document) CombinedTitle() string {
	if len(d.Title) > 0 {
		if len(d.Subtitle) > 0 {
			return fmt.Sprintf("%s : %s", d.Title[0], d.Subtitle[0])
		} else {
			return d.Title[0]
		}
	} else {
		if len(d.Subtitle) > 0 {
			return d.Subtitle[0]
		} else {
			return ""
		}
	}
}

// FullTitle returns everything title
func (d *Document) FullTitle() string {
	return strings.Join(append(d.Title, d.Subtitle...), " ")
}

// ShortTitle returns the first main title only
func (d *Document) ShortTitle() string {
	if len(d.Title) > 0 {
		return d.Title[0]
	} else {
		return ""
	}
}

// Transform converts a document to a schematized version
func Transform(doc Document) (map[string]interface{}, error) {
	output := make(map[string]interface{})

	if doc.URL == "" {
		return nil, errors.New("input document has no URL")
	}

	encoded := base64.StdEncoding.EncodeToString([]byte(doc.URL))
	output["id"] = fmt.Sprintf("ai049%s", encoded)
	output["issn"] = doc.ISSN
	output["publisher"] = doc.Publisher
	output["source_id"] = "49"
	output["title"] = doc.CombinedTitle()
	output["title_full"] = doc.FullTitle()
	output["title_short"] = doc.ShortTitle()
	output["topic"] = doc.Subject
	output["url"] = doc.URL

	if len(doc.ContainerTitle) > 0 {
		output["hierarchy_parent_title"] = doc.ContainerTitle[0]
	}

	if doc.Type == "journal-article" {
		output["format"] = "ElectronicArticle"
	}

	var authors []string
	for _, author := range doc.Authors {
		authors = append(authors, author.String())
	}
	output["author2"] = authors

	if len(doc.Issued) > 0 {
		if len(doc.Issued[0]) > 0 {
			output["publishDateSort"] = doc.Issued[0][0]
		}
	}

	allfields := [][]string{authors, doc.Subject, doc.ISSN, doc.Title,
		doc.Subtitle, doc.ContainerTitle, []string{doc.Publisher, doc.URL}}

	var buf bytes.Buffer
	for _, f := range allfields {
		_, err := buf.WriteString(fmt.Sprintf("%s ", strings.Join(f, " ")))
		if err != nil {
			log.Fatal(err)
		}
	}

	output["allfields"] = buf.String()
	return output, nil
}

// Worker receives batches of strings, parses, transforms and serializes them
func Worker(batches chan []string, out chan []byte, wg *sync.WaitGroup) {
	defer wg.Done()
	var doc Document
	for batch := range batches {
		for _, line := range batch {
			json.Unmarshal([]byte(line), &doc)
			output, err := Transform(doc)
			if err != nil {
				log.Fatal(err)
			}
			b, err := json.Marshal(output)
			if err != nil {
				log.Fatal(err)
			}
			out <- b
		}
	}
}

// Collector collects docs and writes them out to stdout
func Collector(docs chan []byte, done chan bool) {
	f := bufio.NewWriter(os.Stdout)
	defer f.Flush()
	for b := range docs {
		f.Write(b)
		f.Write([]byte("\n"))
	}
	done <- true
}

func main() {

	numWorkers := flag.Int("w", runtime.NumCPU(), "workers")
	batchSize := flag.Int("b", 25000, "batch size")

	flag.Parse()
	runtime.GOMAXPROCS(*numWorkers)

	if flag.NArg() == 0 {
		log.Fatal("input file required")
	}

	ff, err := os.Open(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}
	defer ff.Close()
	reader := bufio.NewReader(ff)

	batches := make(chan []string)
	docs := make(chan []byte)
	done := make(chan bool)

	go Collector(docs, done)

	var wg sync.WaitGroup
	for i := 0; i < runtime.NumCPU(); i++ {
		wg.Add(1)
		go Worker(batches, docs, &wg)
	}

	counter := 0
	batch := make([]string, *batchSize)
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		batch = append(batch, line)
		if counter == *batchSize-1 {
			batches <- batch
			batch = batch[:0]
			counter = 0
		}
		counter++
	}
	batches <- batch
	close(batches)
	wg.Wait()
	close(docs)
	<-done
}
