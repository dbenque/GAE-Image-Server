package GAEImageServer

import (
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"math"
	"net/http"

	"./resize"

	"github.com/gorilla/mux"

	"appengine"
	"appengine/blobstore"

	"github.com/dbenque/goAppengineToolkit/datastoreEntity"

	"appengine/datastore"
	"appengine/taskqueue"
)

const imageKind = "image"

type ImageInGAE struct {
	EntityId     string `json:"entityId"`
	ProfileName  string `json:"profileName"`
	BlobstoreKey string `json:"blobstoreKey" datastore:",noindex"`
}

func (p *ImageInGAE) GetKey() string {
	return p.EntityId + ":" + p.ProfileName
}
func (p *ImageInGAE) GetKind() string {
	return imageKind
}
func (p *ImageInGAE) store(context *appengine.Context) error {

	return datastoreEntity.Store(context, p)
}
func (p *ImageInGAE) remove(context *appengine.Context) error {

	if err := p.retrieve(context); err == nil {
		if err := blobstore.Delete(*context, appengine.BlobKey(p.BlobstoreKey)); err != nil {
			(*context).Infof("Error while removing in blobstore: %v", *p)
		}

		// Due to BlobStore Appengine Bug: https://code.google.com/p/googleappengine/issues/detail?id=6849
		query := datastore.NewQuery("__BlobFileIndex__").Filter("blob_key =", appengine.BlobKey(p.BlobstoreKey)).KeysOnly()
		var values []interface{}
		if keys, err := query.GetAll(*context, &values); err == nil && len(keys) == 1 {
			datastore.Delete(*context, keys[0])
		}
	}

	return datastoreEntity.Delete(context, p)
}
func (p *ImageInGAE) retrieve(context *appengine.Context) error {

	if p.ProfileName == "" {
		p.ProfileName = _OriginalProfileName
	}

	return datastoreEntity.Retrieve(context, p)
}
func handleDeleteImageAllProfile(w http.ResponseWriter, r *http.Request) {

	c := appengine.NewContext(r)
	vars := mux.Vars(r)

	q := datastore.NewQuery(imageKind).Filter("EntityId =", vars["entityId"])
	if err := CreateTasksToDeleteImages(&c, q); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)

}
func CreateTasksToDeleteImages(context *appengine.Context, q *datastore.Query) error {

	t := q.Run(*context)
	for {
		var p ImageInGAE
		_, err := t.Next(&p)
		if err == datastore.Done {
			break // No further entities match the query.
		}
		if err != nil {
			(*context).Errorf("fetching next Image: %v", err)
			break
		}

		t := taskqueue.NewPOSTTask("/image/taskDelete", map[string][]string{"entityId": {p.EntityId}, "profileName": {p.ProfileName}})
		if _, err := taskqueue.Add((*context), t, ""); err != nil {
			return err
		}
	}
	return nil
}
func handleDeleteImageTask(w http.ResponseWriter, r *http.Request) {

	c := appengine.NewContext(r)
	r.ParseForm()
	img := ImageInGAE{EntityId: r.PostFormValue("entityId"), ProfileName: r.PostFormValue("profileName")}

	img.remove(&c)

	w.WriteHeader(http.StatusOK)

}
func handleDeleteImage(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	c := appengine.NewContext(r)
	img := ImageInGAE{EntityId: vars["entityId"], ProfileName: vars["profileName"]}

	img.remove(&c)

	w.WriteHeader(http.StatusOK)

}

func handleGetImage(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	c := appengine.NewContext(r)
	img := ImageInGAE{EntityId: vars["entityId"], ProfileName: vars["profileName"]}

	if err := img.retrieve(&c); err != nil {

		// check if the image exist with the "original" profile
		img.ProfileName = _OriginalProfileName

		if err := img.retrieve(&c); err != nil {
			c.Infof("Error: TODO serve a default image")
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// need to resize and serve
		img.createResizedImage(&c, vars["profileName"])
	}

	// Serve Image
	blobstore.Send(w, appengine.BlobKey(img.BlobstoreKey))
}

func (p *ImageInGAE) createResizedImage(context *appengine.Context, targetProfileName string) error {

	targetProfile := ImgProfile{Name: targetProfileName}
	if err := targetProfile.retrieve(context); err != nil {
		return err
	}

	reader := blobstore.NewReader(*context, appengine.BlobKey(p.BlobstoreKey))

	img, _, err := image.Decode(reader)

	if err != nil {
		return err
	}

	if targetProfile.MaxSize > 0 && (img.Bounds().Max.X > targetProfile.MaxSize || img.Bounds().Max.Y > targetProfile.MaxSize) {
		size_x := img.Bounds().Max.X
		size_y := img.Bounds().Max.Y
		if size_x > targetProfile.MaxSize {
			size_x_before := size_x
			size_x = targetProfile.MaxSize
			size_y = int(math.Floor(float64(size_y) * float64(float64(size_x)/float64(size_x_before))))
		}
		if size_y > targetProfile.MaxSize {
			size_y_before := size_y
			size_y = targetProfile.MaxSize
			size_x = int(math.Floor(float64(size_x) * float64(float64(size_y)/float64(size_y_before))))
		}
		img = resizeImage.Resize(img, img.Bounds(), size_x, size_y)
	}
	// JPEG options
	o := &jpeg.Options{Quality: targetProfile.Quality}

	// get writerurn e
	writer, err := blobstore.Create(*context, "image/jpeg")
	if err != nil {
		return err
	}

	// Write to Blobstore
	if err := jpeg.Encode(writer, img, o); err != nil {
		_ = writer.Close()
		return err
	}

	// close writer
	if err := writer.Close(); err != nil {
		return err
	}

	newKey, err := writer.Key()
	if err != nil {
		return err
	}

	// All good, need to update the image object and store it
	p.BlobstoreKey = string(newKey)
	p.ProfileName = targetProfileName

	p.store(context)

	return nil

}
