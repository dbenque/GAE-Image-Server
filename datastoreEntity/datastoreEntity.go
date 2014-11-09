package datastoreEntity

import (
	"encoding/json"

	"appengine"
	"appengine/datastore"
	"appengine/memcache"
)

type DatastoreEntity interface {
	GetKey() string
	GetKind() string
}

func Delete(context *appengine.Context, entity DatastoreEntity) error {
	key := datastore.NewKey(*context, (entity).GetKind(), (entity).GetKey(), 0, nil)

	if err := datastore.Delete(*context, key); err != nil {
		(*context).Infof("Error while removing entity  %v", entity)
		return err
	}

	if err := memcache.Delete(*context, key.String()); err != nil {
		(*context).Infof("Error while removing entity from memcache %v", entity)
		return err
	}
	return nil

}

func Retrieve(context *appengine.Context, entity DatastoreEntity) error {

	key := datastore.NewKey(*context, (entity).GetKind(), (entity).GetKey(), 0, nil)

	//Try using memcache
	if item, err := memcache.Get(*context, key.String()); err == nil {
		// Nice getting from memchache
		if err := json.Unmarshal(item.Value, entity); err == nil {
			return nil
		}
		(*context).Infof("Error while unserializing json from memcache")
	}

	if err := datastore.Get(*context, key, entity); err != nil {
		(*context).Infof("No entity %v", entity)
		return err
	}
	return nil
}

func Store(context *appengine.Context, entity DatastoreEntity) error {

	key := datastore.NewKey(*context, (entity).GetKind(), (entity).GetKey(), 0, nil)

	if _, err := datastore.Put(*context, key, entity); err != nil {
		(*context).Infof("Error while storing entity %v", entity)
		return err
	}

	//Update memcache
	b, errJson := json.Marshal(entity)
	if errJson == nil {
		item := &memcache.Item{
			Key:   key.String(),
			Value: b,
		}

		if err := memcache.Set(*context, item); err != nil {
			(*context).Infof("Error while storing in memcache %v", key)
		}
	}

	return nil
}
