package xfile_test

import (
	"github.com/develop-top/due/v2/utils/xfile"
	"testing"
)

func TestStat(t *testing.T) {
	fi, err := xfile.Stat("a.txt")
	if err != nil {
		t.Fatal(err)
	}

	t.Log(fi.CreateTime())
}
