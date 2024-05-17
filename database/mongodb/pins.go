package mongodb

import (
	"context"
	"encoding/json"
	"log"
	"manindexer/pin"
	"strconv"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (mg *Mongodb) GetMaxHeight() (height int64, err error) {
	findOp := options.FindOne()
	findOp.SetSort(bson.D{{Key: "number", Value: -1}})
	var pinInscription pin.PinInscription
	err = mongoClient.Collection(PinsCollection).FindOne(context.TODO(), bson.D{}, findOp).Decode(&pinInscription)
	if err != nil && err == mongo.ErrNoDocuments {
		err = nil
		return
	}
	if pinInscription.GenesisHeight > 1 {
		height = pinInscription.GenesisHeight
	}
	return
}

func (mg *Mongodb) GetMaxNumber() (number int64) {
	findOp := options.FindOne()
	findOp.SetSort(bson.D{{Key: "number", Value: -1}})
	var pinInscription pin.PinInscription
	err := mongoClient.Collection(PinsCollection).FindOne(context.TODO(), bson.D{}, findOp).Decode(&pinInscription)
	if err != nil && err == mongo.ErrNoDocuments {
		err = nil
		return
	}
	number = pinInscription.Number + 1
	return
}

func (mg *Mongodb) BatchAddPins(pins []interface{}) (err error) {
	ordered := false
	option := options.InsertManyOptions{Ordered: &ordered}
	_, err = mongoClient.Collection(PinsCollection).InsertMany(context.TODO(), pins, &option)
	return
}
func (mg *Mongodb) UpdateTransferPin(addressMap map[string]string) (err error) {
	var models []mongo.WriteModel
	for id, address := range addressMap {
		filter := bson.D{{Key: "id", Value: id}}
		var updateInfo bson.D
		updateInfo = append(updateInfo, bson.E{Key: "istransfered", Value: true})
		updateInfo = append(updateInfo, bson.E{Key: "address", Value: address})
		update := bson.D{{Key: "$set", Value: updateInfo}}
		m := mongo.NewUpdateOneModel()
		m.SetFilter(filter).SetUpdate(update)
		models = append(models, m)
	}
	bulkWriteOptions := options.BulkWrite().SetOrdered(false)
	_, err = mongoClient.Collection(PinsCollection).BulkWrite(context.Background(), models, bulkWriteOptions)

	return
}
func (mg *Mongodb) BatchUpdatePins(pins []*pin.PinInscription) (err error) {
	var models []mongo.WriteModel
	for _, pin := range pins {
		if pin.OriginalId == "" {
			continue
		}
		filter := bson.D{{Key: "id", Value: pin.OriginalId}, {Key: "address", Value: pin.Address}}
		var updateInfo bson.D
		if pin.Status != 0 {
			updateInfo = append(updateInfo, bson.E{Key: "status", Value: pin.Status})
		}
		update := bson.D{{Key: "$set", Value: updateInfo}}
		m := mongo.NewUpdateOneModel()
		m.SetFilter(filter).SetUpdate(update)
		models = append(models, m)
	}
	bulkWriteOptions := options.BulkWrite().SetOrdered(false)
	_, err = mongoClient.Collection(PinsCollection).BulkWrite(context.Background(), models, bulkWriteOptions)

	return
}
func (mg *Mongodb) AddMempoolPin(pin *pin.PinInscription) (err error) {
	_, err = mongoClient.Collection(MempoolPinsCollection).InsertOne(context.TODO(), pin)
	return
}
func (mg *Mongodb) GetPinPageList(page int64, size int64) (pins []*pin.PinInscription, err error) {
	cursor := (page - 1) * size
	opts := options.Find().SetSort(bson.D{{Key: "number", Value: -1}}).SetSkip(cursor).SetLimit(size)
	result, err := mongoClient.Collection(PinsCollection).Find(context.TODO(), bson.M{}, opts)
	if err != nil {
		return
	}
	err = result.All(context.TODO(), &pins)
	return
}
func (mg *Mongodb) GetPinListByIdList(idList []string) (pinList []*pin.PinInscription, err error) {
	filter := bson.M{"id": bson.M{"$in": idList}}
	result, err := mongoClient.Collection(PinsCollection).Find(context.TODO(), filter)
	if err != nil {
		return
	}
	err = result.All(context.TODO(), &pinList)
	return
}
func (mg *Mongodb) GetMempoolPinPageList(page int64, size int64) (pins []*pin.PinInscription, err error) {
	cursor := (page - 1) * size
	opts := options.Find().SetSort(bson.D{{Key: "timestamp", Value: -1}, {Key: "number", Value: -1}}).SetSkip(cursor).SetLimit(size)
	result, err := mongoClient.Collection(MempoolPinsCollection).Find(context.TODO(), bson.M{}, opts)
	if err != nil {
		return
	}
	err = result.All(context.TODO(), &pins)
	return
}
func (mg *Mongodb) DeleteMempoolInscription(txIds []string) (err error) {
	filter := bson.M{"id": bson.M{"$in": txIds}}
	_, err = mongoClient.Collection(MempoolPinsCollection).DeleteMany(context.TODO(), filter)
	if err != nil {
		log.Println("DeleteMempoolInscription err", err)
	}
	return
}
func (mg *Mongodb) GetPinListByAddress(address string, addressType string, cursor int64, size int64) (pins []*pin.PinInscription, err error) {
	opts := options.Find().SetSort(bson.D{{Key: "number", Value: -1}}).SetSkip(cursor).SetLimit(size)
	filter := bson.M{"address": address, "status": 0}
	if addressType == "creator" {
		filter = bson.M{"createaddress": address, "status": 0}
	}
	result, err := mongoClient.Collection(PinsCollection).Find(context.TODO(), filter, opts)
	if err != nil {
		return
	}
	err = result.All(context.TODO(), &pins)
	return
}

func (mg *Mongodb) GetPinRootByAddress(address string) (pin *pin.PinInscription, err error) {
	err = mongoClient.Collection(PinsCollection).FindOne(context.TODO(), bson.M{"address": address, "operation": "init"}).Decode(&pin)
	if err == mongo.ErrNoDocuments {
		err = mongoClient.Collection(MempoolPinsCollection).FindOne(context.TODO(), bson.M{"address": address, "operation": "init"}).Decode(&pin)
	}
	return
}

func (mg *Mongodb) GetPinByNumberOrId(numberOrId string) (pinInscription *pin.PinInscription, err error) {
	number, err1 := strconv.ParseInt(numberOrId, 10, 64)
	if err1 == nil {
		err = mongoClient.Collection(PinsCollection).FindOne(context.TODO(), bson.D{{Key: "number", Value: number}}).Decode(&pinInscription)
	} else {
		err = mongoClient.Collection(PinsCollection).FindOne(context.TODO(), bson.D{{Key: "id", Value: numberOrId}}).Decode(&pinInscription)
	}
	if err == mongo.ErrNoDocuments {
		pinInscription, err = mg.GetMemPoolPinByNumberOrId(numberOrId)
	}
	if err == mongo.ErrNoDocuments {
		err = nil
	}
	return
}

func (mg *Mongodb) GetMemPoolPinByNumberOrId(numberOrId string) (pinInscription *pin.PinInscription, err error) {
	number, err1 := strconv.ParseInt(numberOrId, 10, 64)
	if err1 == nil {
		err = mongoClient.Collection(MempoolPinsCollection).FindOne(context.TODO(), bson.D{{Key: "number", Value: number}}).Decode(&pinInscription)
	} else {
		err = mongoClient.Collection(MempoolPinsCollection).FindOne(context.TODO(), bson.D{{Key: "id", Value: numberOrId}}).Decode(&pinInscription)
	}
	if err == mongo.ErrNoDocuments {
		err = nil
	}
	return
}

func (mg *Mongodb) GetBlockPin(height int64, size int64) (pins []*pin.PinInscription, total int64, err error) {
	filter := bson.D{{Key: "genesisheight", Value: height}}
	opts := options.Find().SetSort(bson.D{{Key: "number", Value: -1}}).SetLimit(size)
	result, err := mongoClient.Collection(PinsCollection).Find(context.TODO(), filter, opts)
	if err != nil {
		return
	}
	err = result.All(context.TODO(), &pins)
	if err != nil {
		return
	}
	total, err = mongoClient.Collection(PinsCollection).CountDocuments(context.TODO(), filter)
	return
}

func (mg *Mongodb) GetMetaIdPin(roottxid string, page int64, size int64) (pins []*pin.PinInscription, total int64, err error) {
	cursor := (page - 1) * size
	filter := bson.D{{Key: "roottxid", Value: roottxid}}
	opts := options.Find().SetSort(bson.D{{Key: "number", Value: -1}}).SetSkip(cursor).SetLimit(size)
	result, err := mongoClient.Collection(PinsCollection).Find(context.TODO(), filter, opts)
	if err != nil {
		return
	}
	err = result.All(context.TODO(), &pins)
	if err != nil {
		return
	}
	total, err = mongoClient.Collection(PinsCollection).CountDocuments(context.TODO(), filter)
	return
}
func (mg *Mongodb) GetChildNodeById(roottxid string) (pins []*pin.PinInscription, err error) {
	var p *pin.PinInscription
	err = mongoClient.Collection(PinsCollection).FindOne(context.TODO(), bson.M{"roottxid": roottxid}).Decode(&p)
	if err != nil {
		return
	}
	filter := bson.D{{Key: "parentpath", Value: p.Path}}
	opts := options.Find().SetSort(bson.D{{Key: "number", Value: -1}})
	result, err := mongoClient.Collection(PinsCollection).Find(context.TODO(), filter, opts)
	if err != nil {
		return
	}
	err = result.All(context.TODO(), &pins)
	return
}
func (mg *Mongodb) GetParentNodeById(pinId string) (pinnode *pin.PinInscription, err error) {
	var p *pin.PinInscription
	err = mongoClient.Collection(PinsCollection).FindOne(context.TODO(), bson.M{"id": pinId}).Decode(&p)
	if err != nil {
		return
	}
	err = mongoClient.Collection(PinsCollection).FindOne(context.TODO(), bson.M{"roottxid": p.RootTxId, "path": p.ParentPath}).Decode(&p)
	if err != nil {
		return
	}
	return
}
func (mg *Mongodb) GetAllPinByPath(page, limit int64, path string) (pins []*pin.PinInscription, total int64, err error) {
	filter := bson.M{"path": path}
	cursor := (page - 1) * limit
	opts := options.Find().SetSort(bson.D{{Key: "timestamp", Value: -1}, {Key: "number", Value: -1}}).SetSkip(cursor).SetLimit(limit)
	mempoolResult, err := mongoClient.Collection(MempoolPinsCollection).Find(context.TODO(), filter, opts)
	if err != nil && err != mongo.ErrNoDocuments {
		return
	}
	var memPins []*pin.PinInscription
	var blockPins []*pin.PinInscription
	if mempoolResult != nil {
		err = mempoolResult.All(context.TODO(), &memPins)
		if err != nil {
			return
		}
	}
	newLimit := limit - int64(len(memPins))
	if newLimit > 0 {
		opts = options.Find().SetSort(bson.D{{Key: "timestamp", Value: -1}, {Key: "number", Value: -1}}).SetSkip(cursor).SetLimit(newLimit)
		result, err1 := mongoClient.Collection(PinsCollection).Find(context.TODO(), filter, opts)
		if err1 != nil {
			return
		}
		err = result.All(context.TODO(), &blockPins)
		if err != nil {
			return
		}
	}
	var blockTotal int64
	var memTotal int64
	blockTotal, err = mongoClient.Collection(PinsCollection).CountDocuments(context.TODO(), filter)
	memTotal, err = mongoClient.Collection(MempoolPinsCollection).CountDocuments(context.TODO(), filter)
	total = blockTotal + memTotal
	pins = append(pins, memPins...)
	pins = append(pins, blockPins...)
	return
}
func (mg *Mongodb) BatchAddProtocolData(pins []*pin.PinInscription) (err error) {
	dataMap := make(map[string][]*pin.PinInscription)
	for _, pinItem := range pins {
		keyArr := strings.Split(pinItem.Path, "/")
		key := keyArr[len(keyArr)-1]
		if list, ok := dataMap[key]; ok {
			dataMap[key] = append(list, pinItem)
		} else {
			dataMap[key] = []*pin.PinInscription{pinItem}
		}
	}
	//ordered := false
	//option := options.InsertManyOptions{Ordered: &ordered}
	for collectionName, pinList := range dataMap {
		data := getDataByContent(pinList)
		if len(data) > 0 {
			upsertProtocolData(data, collectionName)
			//mongoClient.Collection(collectionName).InsertMany(context.TODO(), data, &option)
		}
	}
	return
}
func upsertProtocolData(data []map[string]interface{}, collectionName string) (err error) {
	var models []mongo.WriteModel
	for _, info := range data {
		filter := bson.D{{Key: "pinId", Value: info["pinId"]}}
		var updateInfo bson.D
		for k, v := range info {
			updateInfo = append(updateInfo, bson.E{Key: k, Value: v})
		}
		update := bson.D{{Key: "$set", Value: updateInfo}}
		m := mongo.NewUpdateOneModel()
		m.SetFilter(filter).SetUpdate(update).SetUpsert(true)
		models = append(models, m)
	}
	_, err = mongoClient.Collection(collectionName).BulkWrite(context.Background(), models)
	return
}
func getDataByContent(pinList []*pin.PinInscription) (data []map[string]interface{}) {
	for _, pinItem := range pinList {
		var d map[string]interface{}
		err := json.Unmarshal(pinItem.ContentBody, &d)
		if err == nil {
			d["pinId"] = pinItem.Id
			d["pinNumber"] = pinItem.Number
			d["pinAddress"] = pinItem.Address
			data = append(data, d)
		} else {
			//fmt.Println(err)
		}
	}
	return
}