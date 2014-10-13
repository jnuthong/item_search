
package mongo

import (// "fmt"
	"errors"
	// "encoding/json"
	// "reflect"

	"gopkg.in/mgo.v2"
	"github.com/jnuthong/item_search"
)

type Object map[string] interface{}
type ListMap []map[string] interface{}

type MongoIndex struct {
	Key []string
	Unique bool
	DropDups bool
	Background bool
	Sparse bool	
}

func CreateMongoIndex (value interface{}) (mgo.Index, error){
	// TODO should setup the default value
	// x := []string{"key", "Unqiue", "DropDups", "Backgroud", "Sparse"}
	x, err := value.(Object)["Key"]
	if !err {
		indexOpts := mgo.Index{
			Key: []string{item_search.DefaultMongIndexKey}, 	// Index key fields; prefix name with dash (-) for descending order
			Unique: false, 		// Prevent two documents from having the same index key
			DropDups: false,	// Drop documents with the same index key as a previously indexed one
			Background: true,	// Build index in background and return immediately
			Sparse: true,		// Only index documents containing the Key fields
		}
		return indexOpts, nil
	}
	indexOpts := mgo.Index{
			Key: []string{x.(string)}, 	// Index key fields; prefix name with dash (-) for descending order
			Unique: false, 			// Prevent two documents from having the same index key
			DropDups: false,		// Drop documents with the same index key as a previously indexed one
			Background: true,		// Build index in background and return immediately
			Sparse: true,			// Only index documents containing the Key fields
		}
	return indexOpts, nil
}

func CreateMongoDBCollection (addr string, option interface{}) error{
	conn, err:= mgo.Dial(addr)
	if err != nil{
		return err
	}
	conn.SetSafe(&mgo.Safe{})

	dbName := item_search.DefaultDBName		// setting the default DBName
	if x, ok := option.(Object)["database_name"]; ok {
		dbName = x.(string)
	}
	db := conn.DB(dbName)

	dbCollection := item_search.DefaultCollection 	// setting the default Collection Name 
	if x, ok := option.(Object)["collection_name"]; ok{
		dbCollection = x.(string)
	}

	if tmp, ok := option.(Object)["index"]; ok{
		// xy := reflect.ValueOf(&tmp).Elem()
		for index:= range tmp.(ListMap){
			index_entity, err := CreateMongoIndex(tmp.(ListMap)[index])
			if err != nil{	
				db.C(dbCollection).EnsureIndex(index_entity)
			}	
		}
	}
	return errors.New("[error]")
}
