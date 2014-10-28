
package mongo

import ("fmt"
	// "errors"
	// "encoding/json"
	// "reflect"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"github.com/jnuthong/item_search/utils/log"
)

type Object map[string] interface{}
type ListMap []map[string] interface{}

type MongoObj struct {
	collection *mgo.Collection
	session *mgo.Session
	db 	*mgo.Database
}

func (opt *MongoObj) GetSessioni() (*mgo.Session) {
	return opt.session
}

func (opt *MongoObj) GetDb() (*mgo.Database){
	return opt.db
}

func (opt *MongoObj) GetCollection () (*mgo.Collection){
	return opt.collection
}

type MongoIndex struct {
	Key []string
	Unique bool
	DropDups bool
	Background bool
	Sparse bool	
}

func CreateMongoIndex (value interface{}, DefaultMongoIndexKey string) (mgo.Index, error){
	// TODO should setup the default value
	// x := []string{"key", "Unqiue", "DropDups", "Backgroud", "Sparse"}
	x, err := value.(map[string]interface{})["Key"]
	if !err || x == nil {
		indexOpts := mgo.Index{
			Key: []string{DefaultMongoIndexKey}, 	// Index key fields; prefix name with dash (-) for descending order
			Unique: false, 		// Prevent two documents from having the same index key
			DropDups: false,	// Drop documents with the same index key as a previously indexed one
			Background: true,	// Build index in background and return immediately
			Sparse: true,		// Only index documents containing the Key fields
		}
		return indexOpts, nil
	}
	fmt.Println("[Info] Create the Mongo.Index for key", x)
	indexOpts := mgo.Index{
			Key: []string{x.(string)}, 	// Index key fields; prefix name with dash (-) for descending order
			Unique: false, 			// Prevent two documents from having the same index key
			DropDups: false,		// Drop documents with the same index key as a previously indexed one
			Background: true,		// Build index in background and return immediately
			Sparse: true,			// Only index documents containing the Key fields
		}
	return indexOpts, nil
}

// REF: http://godoc.org/gopkg.in/mgo.v2#Collection.Insert
func CreateMongoDBCollection (addr string, option interface{}) (*MongoObj, error){
	conn, err:= mgo.Dial(addr)
	if err != nil{
		return nil, err
	}
	conn.SetSafe(&mgo.Safe{})

	dbName := option.(map[string]interface{})["DefaultDBName"]		// setting the default DBName
	db := conn.DB(dbName.(string))

	dbCollection := option.(map[string]interface{})["DefaultCollection"] 	// setting the default Collection Name 
	dbIndex := option.(map[string]interface{})["DefaultMongIndexKey"]

	if tmp, ok := option.(map[string]interface{})["Index"]; ok{
		// xy := reflect.ValueOf(&tmp).Elem()
		for i := range tmp.([]map[string]interface{}){
			index_entity, err := CreateMongoIndex(tmp.([]map[string]interface{})[i], dbIndex.(string))
			if err != nil{
				db.C(dbCollection.(string)).EnsureIndex(index_entity)
			}	
		}
	}

	var instance MongoObj
	instance.db = conn.DB(dbName.(string))
	instance.session = conn
	instance.collection = db.C(dbCollection.(string))
	return &instance, nil
}

// http://docs.mongodb.org/manual/tutorial/insert-documents/

// +++++++++++++++++++++++++++++++++++++++++++++++++++
// ++++++++++ operation on the collection ++++++++++++
// +++++++++++++++++++++++++++++++++++++++++++++++++++

// CREATE ---------- set up the writting process for collection -----------

// writing single/multiple documents for the collection
func InsertDoc(c *mgo.Collection, value interface{}) error{
	return c.Insert(value)
}

// set up queue in inseration & insert data back up
// create a insert bulk
func CreateBulk(c *mgo.Collection) *mgo.Bulk {
	return c.Bulk()
}

// insert a document to a bulk
func InsertBulk(bulk *mgo.Bulk, doc interface{}) bool{
	bulk.Insert(doc)
	return true
}

// start a bulk for a collection
func RunBulk(bulk *mgo.Bulk) (*mgo.BulkResult, error){
	result, err := bulk.Run()
	return result, err
}

// UPDATE ----------- update mongo data --------------

func UpdateOrInsert_DocByDocID(c *mgo.Collection, id string, value map[string]interface{}) (*mgo.ChangeInfo, error){
	out := c.Find(bson.M{"id": id})
	doc := bson.M{}
	for sub_key, sub_value := range value{
		doc[sub_key] = sub_value			
	}

	// couldn't find the document, so insert the new document
	if count, err := out.Count(); count == 0 && err == nil{
		c.Insert(doc)
		return nil, nil
	}

	change := mgo.Change{
			Update : bson.M{"$set": doc},
			ReturnNew : true,
			Upsert : true,
	}
	var result interface{}
	info, err := out.Apply(change, &result)
	if err != nil{
		log.Log("error", fmt.Sprintf("%s", err))
		return info, err
	}
	return info, nil
}

// update the current doc or insert new one depend on whether corresponding doc path exist or not
// Ref : http://godoc.org/gopkg.in/mgo.v2#Query.Apply
// Ref : http://docs.mongodb.org/manual/tutorial/modify-documents/ 
func UpdateOrInsert_FieldByDocID(c *mgo.Collection, id string, field string, value interface{})(*mgo.ChangeInfo, error){
	out := c.Find(bson.M{"id": id})
	if count, err := out.Count(); count == 0 && err == nil{
		c.Insert(bson.M{"id": id, field: value})
		return nil, nil
	}

	change := mgo.Change{
			Update : bson.M{"$set" : bson.M{field : value}}, 
			ReturnNew : true,
		}

	var result interface{}
	info, err := out.Apply(change, &result)
	if err != nil{
		log.Log("error", fmt.Sprintf("%s", err))
	}
	return info, nil
}

/*
func UpdateByID(c *mgo.Collection, id string, value interface{}) error {
	
}

// QUERY ------------ set up PIPE query line ---------------

func Find(db *mgo.Database, c_name string, query interface{}) {
	c := db.C(c_name)
	r := c.Find(query)
}

// REF: http://docs.mongodb.org/manual/core/aggregation-pipeline/
func Pipe(c *mgo.Collection) *mgo.Pipe {

}
*/
