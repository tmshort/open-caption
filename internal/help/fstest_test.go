package help

import (
	"io/fs"
	"testing/fstest"
)

func fstest_(name, body string) fs.FS { return fstest.MapFS{name: {Data: []byte(body)}} }
