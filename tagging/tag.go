package tagging

import (
	"io/ioutil"

	"github.com/miku/span/atomic"
	"github.com/sethgrid/pester"
)

// AtomicDownload retrieves a link and saves its content atomically in
// filename. TODO(martin): should live in an io related package.
func AtomicDownload(link, filename string) error {
	resp, err := pester.Get(link)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return atomic.WriteFile(filename, b, 0644)
}
