package ngword

import (
	"github.com/go-gota/gota/dataframe"
	"os"
)

func ReadDataframeFromCSV(fname string) dataframe.DataFrame {
	fp, err := os.Open(fname)
	if err != nil {
		panic(err)
	}
	defer fp.Close()
	return dataframe.ReadCSV(fp)
}
