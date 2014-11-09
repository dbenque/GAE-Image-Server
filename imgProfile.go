package GAEImageServer

import (
	"encoding/json"
	"net/http"
	"sort"

	"github.com/gorilla/mux"

	"datastoreEntity"

	"appengine"
	"appengine/datastore"
)

const _OriginalProfileName = "original"

type ImgProfile struct {
	Name    string `json:"name"`
	Quality int    `json:"quality" datastore:",noindex"`
	MaxSize int    `json:"maxSize" datastore:",noindex"`
}

type ImgProfileNameSet struct {
	Index sort.StringSlice `json:"index"` // Index of profiles

	//Index []string `json:"index"` // Index of profiles
}

// retrieve the index of profiles (initialize the key if nil)
func GetImgProfileNameSet(context *appengine.Context) ImgProfileNameSet {

	key := datastore.NewKey(*context, "ImgProfileNameSet", "ImgProfileNameSet", 0, nil)

	allProfiles := ImgProfileNameSet{}

	if err := datastore.Get(*context, key, &allProfiles); err != nil {
		(*context).Infof("No profile could be retrieved")
		return ImgProfileNameSet{}
	}
	return allProfiles

}

func (p *ImgProfile) GetKey() string {
	return p.Name
}
func (p *ImgProfile) GetKind() string {
	return "ImgProfile"
}

func (p *ImgProfile) store(context *appengine.Context) error {

	if err := datastoreEntity.Store(context, p); err != nil {
		return err
	}

	// Update profile index
	keyIndex := datastore.NewKey(*context, "ImgProfileNameSet", "ImgProfileNameSet", 0, nil)
	allProfilesStruct := GetImgProfileNameSet(context)

	// avoid dupe insertion
	var found bool = false
	for _, value := range allProfilesStruct.Index {
		found = value == p.Name
		if found {
			break
		}
	}

	if !found {
		allProfilesStruct.Index = sort.StringSlice(append(allProfilesStruct.Index, p.Name)[0:])
		allProfilesStruct.Index.Sort()
		if _, err := datastore.Put(*context, keyIndex, &allProfilesStruct); err != nil {
			(*context).Errorf("Can't store index, %v", err)
		}
	}

	(*context).Infof("Profile stored: %v", *p)

	return nil
}

func (p *ImgProfile) remove(context *appengine.Context) error {

	if err := datastoreEntity.Delete(context, p); err != nil {
		return err
	}

	// Update profile index
	keyIndex := datastore.NewKey(*context, "ImgProfileNameSet", "ImgProfileNameSet", 0, nil)
	allProfilesStruct := GetImgProfileNameSet(context)
	i := allProfilesStruct.Index.Search(p.Name)
	if i != len(allProfilesStruct.Index) {
		allProfilesStruct.Index = sort.StringSlice(append(allProfilesStruct.Index[:i], allProfilesStruct.Index[i+1:]...)[0:]) // remove element at i and create a StringSlice with remaining items
		allProfilesStruct.Index.Sort()
		if _, err := datastore.Put(*context, keyIndex, &allProfilesStruct); err != nil {
			(*context).Errorf("Can't store index, %v", err)
		}

	}
	(*context).Infof("Profile removed: %v", p.Name)

	return nil
}

func (p *ImgProfile) retrieve(context *appengine.Context) error {

	return datastoreEntity.Retrieve(context, p)
}

func handleImgProfileGet(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	c := appengine.NewContext(r)

	profile := ImgProfile{Name: vars["name"]}
	if err := profile.retrieve(&c); err != nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	profileJson, _ := json.Marshal(profile)
	w.Header().Set("Content-Type", "application/json")
	w.Write(profileJson)

}

func handleImgProfileStore(w http.ResponseWriter, r *http.Request) {

	c := appengine.NewContext(r)

	decoder := json.NewDecoder(r.Body)
	var profile ImgProfile
	err := decoder.Decode(&profile)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	profile.store(&c)
	w.WriteHeader(http.StatusOK)

}

func handleImgProfileDelete(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	c := appengine.NewContext(r)

	profile := ImgProfile{Name: vars["name"]}

	if profile.Name == _OriginalProfileName {
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}

	profile.remove(&c)

	// Clean all associated image
	q := datastore.NewQuery(imageKind).Filter("ProfileName =", vars["name"])
	if err := CreateTasksToDeleteImages(&c, q); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
func handleGetAllProfiles(w http.ResponseWriter, r *http.Request) {

	c := appengine.NewContext(r)

	profilesJson, _ := json.Marshal(GetImgProfileNameSet(&c).Index)
	w.Header().Set("Content-Type", "application/json")
	w.Write(profilesJson)
}
