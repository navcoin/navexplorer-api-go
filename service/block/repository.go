package block

import (
	"github.com/NavExplorer/navexplorer-api-go/db"
	"github.com/globalsign/mgo/bson"
)

type Repository struct{}

func (r *Repository) FindBlocks(dir string, size int, offset string) (blocks []Block, err error){
	dbConnection := db.NewConnection()
	c := dbConnection.Use( "block")
	defer dbConnection.Close()

	conditions := make(bson.M, 0)

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

	err = q.All(&blocks)

	return blocks, err
}

func (r *Repository) FindOneBlockByHash(hash string) (block Block, err error) {
	dbConnection := db.NewConnection()
	c := dbConnection.Use( "block")
	defer dbConnection.Close()

	err = c.Find(bson.M{"hash": hash}).One(&block)

	return block, err
}

func (r *Repository) FindOneBlockByHeight(height int) (block Block, err error) {
	dbConnection := db.NewConnection()
	c := dbConnection.Use( "block")
	defer dbConnection.Close()

	err = c.Find(bson.M{"height": height}).One(&block)

	return block, err
}

func (r *Repository) FindTransactions(dir string, size int, offset string, types []string) (transactions []Transaction, err error){
	dbConnection := db.NewConnection()
	c := dbConnection.Use( "blockTransaction")
	defer dbConnection.Close()

	conditions := make(bson.M, 0)
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

	err = q.All(&transactions)

	return transactions, err
}

func (r *Repository) FindAllTransactionsByBlockHash(hash string) (transactions []Transaction, err error) {
	dbConnection := db.NewConnection()
	c := dbConnection.Use( "blockTransaction")
	defer dbConnection.Close()

	err = c.Find(bson.M{"blockHash": hash}).Sort("id").All(&transactions)

	return transactions, err
}

func (r *Repository) FindOneTransactionByHash(hash string) (transaction Transaction, err error) {
	dbConnection := db.NewConnection()
	c := dbConnection.Use( "blockTransaction")
	defer dbConnection.Close()

	err = c.Find(bson.M{"hash": hash}).One(&transaction)

	return transaction, err
}
