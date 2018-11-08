package softFork

import (
	"github.com/NavExplorer/navexplorer-api-go/db"
	"github.com/globalsign/mgo/bson"
)

type Repository struct{}

func (r *Repository) FindAll() (softForks []SoftFork, err error) {
	dbConnection := db.NewConnection()

	c := dbConnection.Use("softFork")
	err = c.Find(bson.M{}).Sort("signalBit").All(&softForks)

	defer dbConnection.Close()

	return softForks, err
}