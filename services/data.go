package services

import (
	"fmt"
	"github.com/nutsdb/nutsdb"
)

type DataManagement struct {
	DB *nutsdb.DB
}

func (dm *DataManagement) StartDB(dir string) error {
	fmt.Println("creating database at: ", dir)
	// if err := os.MkdirAll(dir, os.ModePerm); err != nil {
	//   return err
	// }
	db, err := nutsdb.Open(
		nutsdb.DefaultOptions,
		nutsdb.WithDir(dir),
	)
	if err != nil {
		return err
	}
	dm.DB = db
	fmt.Println("database is created an online")
	return nil
}
