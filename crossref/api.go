package crossref

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
)

// Message covers a generic API response.
type Message struct {
	Status  string          `json:"status"`
	Version string          `json:"message-version"`
	Message json.RawMessage `json:"message"`
	Type    string          `json:"message-type"`
}

// Member covers a member type message.
type Member struct {
	ID          int      `json:"id"`
	PrimaryName string   `json:"primary-name"`
	Names       []string `json:"names"`
	Location    string   `json:"location"`
	Prefixes    []string `json:"prefixes"`
	Tokens      []string `json:"tokens"`
}

// Cache for members, thread-safe.
type Cache struct {
	lock    *sync.RWMutex
	Entries map[int]Member
}

// NewMemberCache creates a new in memory members cache.
func NewCache() Cache {
	return Cache{Entries: make(map[int]Member), lock: new(sync.RWMutex)}
}

var cache Cache = NewCache()

// LookupMember retrieves member information via crossref API.
// Subsequent requests are served by an in-memory cache.
// Example: http://api.crossref.org/members/56
func LookupMember(id int) (Member, error) {
	member, ok := cache.Entries[id]
	if !ok {
		link := fmt.Sprintf("http://api.crossref.org/members/%d", id)
		log.Printf("Fetching: %s", link)
		resp, err := http.Get(link)
		if err != nil {
			return member, err
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return member, err
		}
		var message Message
		err = json.Unmarshal(body, &message)
		if err != nil {
			return member, err
		}
		if message.Type == "member" {
			err = json.Unmarshal(message.Message, &member)
			if err != nil {
				return member, err
			}
			cache.lock.Lock()
			if _, ok := cache.Entries[id]; !ok {
				cache.Entries[id] = member
			}
			cache.lock.Unlock()
		} else {
			return member, fmt.Errorf("unsupported message type: %s", message.Type)
		}
	}
	return member, nil
}
