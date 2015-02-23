package crossref

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
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

// FetchMember makes an API request for a member given by its ID.
func FetchMember(id int) (Member, error) {
	var member Member
	link := fmt.Sprintf("http://api.crossref.org/members/%d", id)
	log.Printf("Fetching crossref member: %s", link)

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

	if message.Status != "ok" {
		return member, fmt.Errorf("message status: %s", message.Status)
	}

	if message.Type != "member" {
		return member, fmt.Errorf("invalid message type: %s", message.Type)
	}

	err = json.Unmarshal(message.Message, &member)
	if err != nil {
		return member, err
	}

	return member, nil
}

// cache holds member ids and their primary names
var cache = NewIntStringCache()

// LookupMemberName returns the primary name for a member given by its ID.
// Example URL: http://api.crossref.org/members/56
func LookupMemberName(id int) (name string, err error) {
	name, ok := cache.Entries[id]
	if !ok {
		member, err := FetchMember(id)
		if err != nil {
			return name, err
		}
		name = member.PrimaryName
		cache.Set(id, name)
	}
	return name, nil
}

// PopulateMemberNameCache takes an LDJ filename with one member document per line and populates the cache.
func PopulateMemberNameCache(filename string) error {
	handle, err := os.Open(filename)
	defer handle.Close()
	if err != nil {
		return err
	}
	var member Member
	reader := bufio.NewReader(handle)
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		err = json.Unmarshal([]byte(line), &member)
		if err != nil {
			return err
		}
		cache.Set(member.ID, member.PrimaryName)
	}
	return nil
}
