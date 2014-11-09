package GAEImageServer

import (
	"html/template"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/tomihiltunen/gae-go-image-optimizer"

	"appengine"
	"appengine/blobstore"
	"appengine/urlfetch"
)

func serveError(c appengine.Context, w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Header().Set("Content-Type", "text/plain")
	io.WriteString(w, "Internal Server Error")
	c.Errorf("%v", err)
}

var formTemplate = template.Must(template.New("root").Parse(formTemplateHTML))

const formTemplateHTML = `
<form action="{{.}}" method="POST" enctype="multipart/form-data">
Image: <input type="file" name="file"><br>
<input type="text" name="callbackurl" value="http://0.0.0.0:8080/callbacktest"> [callbackurl] Url to callback once the file is store in the blobstore<br>
<input type="text" name="entityId" value="myId"> [entityId] Id of the entity that is associated with the image. To be reused in future retrieve query<br>
<input type="text" name="extraparam1" value="val1"/>
<input type="text" name="extraparam2" value="val2"/> Any other post parameter will be passed to the [callbackurl] <br>
<input type="submit" name="submit" value="Submit">
</form>`

const callbackUrl = "callbackurl"
const entityId = "entityId"

func handleFormAction(w http.ResponseWriter, r *http.Request) {

	c := appengine.NewContext(r)
	uploadURL, err := blobstore.UploadURL(c, "/uploaded", nil)
	if err != nil {
		serveError(c, w, err)
		return
	}

	w.Write([]byte(uploadURL.String()))
}

func handleForm(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	c.Infof("In handleForm!! Before getting blobstoreURL", r.PostForm)

	uploadURL, err := blobstore.UploadURL(c, "/uploaded", nil)
	if err != nil {
		serveError(c, w, err)
		return
	}

	c.Infof("In handleForm!! Before template processing", r.PostForm)

	w.Header().Set("Content-Type", "text/html")
	err = formTemplate.Execute(w, uploadURL)
	if err != nil {
		c.Errorf("%v", err)
	}
}

func handleCallbackTest(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()
	c := appengine.NewContext(r)
	c.Infof("callbackdefault received form: %v", r.PostForm)
	w.WriteHeader(200)
}

func handleUploadComplete(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
}

func handleUploadedInBlobStore(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	// Create options to resize image
	o := optimg.NewCompressionOptions(r)

	originalProfile := ImgProfile{Name: _OriginalProfileName}
	if originalProfile.retrieve(&c) != nil {
		// Set max size
		o.Size = 300
		// Set quality
		o.Quality = 75

		c.Warningf("No image profile named '%s'", _OriginalProfileName)
	} else {
		// Set max size from profile
		o.Size = originalProfile.MaxSize
		// Set Quality
		o.Quality = originalProfile.Quality
	}

	// Get the resized blobs and other values
	blobs, vals, err := optimg.ParseBlobs(o)
	if err != nil {
		serveError(c, w, err)
		return
	}

	file := blobs["file"]
	if len(file) == 0 {
		c.Errorf("no file uploaded")
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	blobstoreKey := string(file[0].BlobKey)
	vals.Add("blobKey", blobstoreKey)
	c.Infof("Vals: %v", vals)

	// Store the Image Object
	img := Image{EntityId: vals.Get(entityId), ProfileName: _OriginalProfileName, BlobstoreKey: blobstoreKey}
	img.store(&c)

	client := urlfetch.Client(c)
	_, post_err := client.PostForm(vals.Get(callbackUrl), vals)

	if post_err != nil {
		serveError(c, w, post_err)
		return
	}
	http.Redirect(w, r, "/uploaded/complete", http.StatusFound)
}

func init() {

	r := mux.NewRouter()

	// uploading img
	r.HandleFunc("/action", handleFormAction)
	r.HandleFunc("/exampleForm", handleForm)
	r.HandleFunc("/callbacktest", handleCallbackTest)
	r.HandleFunc("/uploaded", handleUploadedInBlobStore)
	r.HandleFunc("/uploaded/complete", handleUploadComplete)

	// ImgProfile
	r.HandleFunc("/imgProfiles", handleGetAllProfiles).Methods("GET")
	r.HandleFunc("/imgProfile", handleImgProfileStore).Methods("POST")
	r.HandleFunc("/imgProfile/{name}", handleImgProfileDelete).Methods("DELETE")
	r.HandleFunc("/imgProfile/{name}", handleImgProfileGet).Methods("GET")

	// Image
	r.HandleFunc("/image/{entityId}/{profileName}", handleGetImage).Methods("GET")
	r.HandleFunc("/image/{entityId}/{profileName}", handleDeleteImage).Methods("DELETE")
	r.HandleFunc("/image/{entityId}", handleDeleteImageAllProfile).Methods("DELETE")
	// ImageTask
	r.HandleFunc("/image/taskDelete", handleDeleteImageTask).Methods("POST")
	http.Handle("/", r)
}
