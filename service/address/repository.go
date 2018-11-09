package address

import (
	"github.com/NavExplorer/navexplorer-api-go/db"
	"github.com/globalsign/mgo/bson"
)

type Repository struct{}

func (r *Repository) FindOneAddressByHash(hash string) (address Address, err error) {
	c := db.NewConnection().Use("address")
	err = c.Find(bson.M{"hash": hash}).One(&address)

	return address, err
}

func (r *Repository) FindTopAddressesOrderByBalanceDesc(count int) (addresses []Address, err error){
	dbConnection := db.NewConnection()
	c := dbConnection.Use( "address")
	defer dbConnection.Close()

	err = c.Find(bson.M{"balance": bson.M{"$gt": 0}}).Sort("-balance").Limit(count).All(&addresses)

	return addresses, err
}

func (r *Repository) GetRichListPosition(address Address) (count int) {
	dbConnection := db.NewConnection()
	c := dbConnection.Use( "address")
	defer dbConnection.Close()

	count, _ = c.Find(bson.M{"balance": bson.M{"$gte": address.Balance}}).Count()

	return count
}

func (r *Repository) FindTransactionsByAddress(address string, dir string, size int, offset string, types []string) (txs []Transaction, total int, err error) {
	dbConnection := db.NewConnection()
	c := dbConnection.Use( "addressTransaction")
	defer dbConnection.Close()

	conditions := make(bson.M, 0)
	conditions["address"] = address

	if len(types) > 0 {
		conditions["type"] = bson.M{"$in": types}
	}

	if offset != "" && bson.IsObjectIdHex(offset) {
		if dir == "ASC" {
			conditions["_id"] = bson.M{"$gt": bson.ObjectIdHex(offset)}
		} else {
			conditions["_id"] = bson.M{"$lt": bson.ObjectIdHex(offset)}
		}
	}

	q := c.Find(conditions)
	total, _ = q.Count()

	if dir == "ASC" {
		q.Sort("_id")
	} else {
		q.Sort("-_id")
	}

	q.Limit(size)

	err = q.All(&txs)

	return txs, total, err
}
