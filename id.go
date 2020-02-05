package span

import (
	"encoding/base64"
	"fmt"
)

// GenFincID returns a finc.id string consisting of an arbitraty prefix (e.g.
// "ai"), source id and URL safe record id. No additional checks, sid and rid
// should not be empty.
func GenFincID(sid, rid string) string {
	return fmt.Sprintf("ai-%s-%s", sid,
		base64.RawURLEncoding.EncodeToString([]byte(rid)))
}
