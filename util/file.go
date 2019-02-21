package util

//	various file system convenience tools
import (
	"io"
	"os"
)

func Move(from, to string) error {
	sf, err := os.Open(from)
	if err != nil {
		return err
	}
	defer sf.Close()

	df, err := os.Create(to)
	if err != nil {
		return err
	}
	defer df.Close()

	_, err = io.Copy(df, sf)
	if err != nil {
		return err
	}

	// clean up
	os.Remove(from)
	return nil
}

//	check if the file exists
func FileExist(src string) bool {
	if _, err := os.Stat(src); os.IsNotExist(err) {
		return false
	}
	return true
}
