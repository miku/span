// Command span converts, licenses and exports library metadata.
//
// This is the only binary. The tools span used to install as separate
// executables (span-import, span-tag, span-export, ...) are symlinks to it; it
// recognises the name it was invoked under and runs the matching command. See
// "span help".
package main

import "github.com/miku/span/internal/cli"

func main() { cli.Main() }
