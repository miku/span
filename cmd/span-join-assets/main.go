// span-join-assets combines a directory of json configurations into a single file.
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/olebedev/config"
)

// Given a path, remove the extension and turn slashes to dots.
func keyFromPath(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) == 0 {
		return ""
	}
	nameext := strings.Split(parts[len(parts)-1], ".")
	parts[len(parts)-1] = nameext[0]

	return strings.Join(parts[:len(parts)], ".")
}

func main() {
	flag.Parse()
	if flag.NArg() == 0 {
		log.Fatal("directory required")
	}
	conf := config.Config{Root: make(map[string]interface{})}
	filepath.Walk(flag.Arg(0), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Mode().IsDir() {
			return nil
		}
		switch {
		// Arbitrary nested json.
		case strings.HasSuffix(path, ".json"):
			doc := make(map[string]interface{})
			file, err := os.Open(path)
			if err != nil {
				log.Fatal(err)
			}
			if err := json.NewDecoder(file).Decode(&doc); err != nil {
				log.Fatal(err)
			}
			key := keyFromPath(path)
			if key == "" {
				return nil
			}
			if err := conf.Set(key, doc); err != nil {
				log.Fatal(err)
			}
		// Single column only for now.
		case strings.HasSuffix(path, ".tsv"):
			var list []string
			file, err := os.Open(path)
			if err != nil {
				log.Fatal(err)
			}
			br := bufio.NewReader(file)
			for {
				line, err := br.ReadString('\n')
				if err == io.EOF {
					break
				}
				if err != nil {
					log.Fatal(err)
				}
				line = strings.TrimSpace(line)
				if len(line) == 0 {
					continue
				}
				if strings.HasPrefix(line, "#") {
					continue
				}
				list = append(list, line)
			}
			key := keyFromPath(path)
			if key == "" {
				return nil
			}
			if err := conf.Set(key, list); err != nil {
				log.Fatal(err)
			}
		}
		return nil
	})
	if err := json.NewEncoder(os.Stdout).Encode(conf.Root); err != nil {
		log.Fatal(err)
	}
}
