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
	c := db.NewConnection().Use( "address")
	err = c.Find(bson.M{"balance": bson.M{"$gt": 0}}).Sort("-balance").Limit(count).All(&addresses)

	return addresses, err
}

func (r *Repository) GetRichListPosition(address Address) (count int) {
	c := db.NewConnection().Use( "address")

	count, _ = c.Find(bson.M{"balance": bson.M{"$gte": address.Balance}}).Count()

	return count
}

func (r *Repository) FindTransactionsByAddress(address string, types []string, dir string, size int, offset string) (txs []Transaction, err error) {
	c := db.NewConnection().Use("addressTransaction")

	conditions := make(bson.M, 0)
	conditions["address"] = address
	conditions["type"] = bson.M{"$in": types}

	if offset != "" && bson.IsObjectIdHex(offset) {
		if dir == "ASC" {
			conditions["_id"] = bson.M{"$gt": bson.ObjectIdHex(offset)}
		} else {
			conditions["_id"] = bson.M{"$lt": bson.ObjectIdHex(offset)}
		}
	}

	q := c.Find(conditions)
	if dir == "ASC" {
		q.Sort("_id")
	} else {
		q.Sort("-_id")
	}

	q.Limit(size)

	err = q.All(&txs)

	return txs, err
}
