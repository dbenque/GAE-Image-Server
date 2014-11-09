GAE-Image-Server
================

Use GAE Blobstore to host your images. Upload image and profiles (Quality/Size). Serve Image for a given profile (resize on the fly).

## Upload
To get the "url action" to put in the upload form:

```http://{img20sur20_baseURL}/action | GET```

Example of form to be used in the upload page:

```http://{img20sur20_baseURL}/exampleForm | GET```

Of course the form will have to post the file to be uploaded, but note also the 2 important fields  that help to index the image and continue the flow:
```
<input type="text" name="entityId" value="myId">
<input type="text" name="callbackurl" value="http://0.0.0.0:8080/callbacktest">
```

Once the form is posted to the blobstore action, the file will be stored, and the *callbackurl* will be invoked with the updated form. All the input parameters of the initial form will be transfered.

## Image Profiles
**Retrieve the list** of existing profiles:

```http://{img20sur20_baseURL}/imgProfiles | GET```

This returns a JSON flat list containing the list of profile names:

```
[
"original",      // This profil is created by default, and  is associated by default to any uploaded image (+resize if needed)
"small"
]
```

**Retrieve** a specific profile:

```http://{img20sur20_baseURL}/imgProfile/{profileName} | GET```

**Delete** a specific profile:

```http://{img20sur20_baseURL}/imgProfile/{profileName} | DELETE```

This removes all the files associated to the profile.

**Create** or **update** a specific profile:
```
http://{img20sur20_baseURL}/imgProfile | POST | JSON {
        "name" : "small",
        "quality" : 50,
        "maxSize" : 100
}
```

Note that to update a profile you need to post a JSON with the name associated to the profile to update. So far a modification of the profile does dot trig a resizing/recomputation of the images associated to the profile. To force the resizing/recomputation, you need to delete the profile. This will remove (asynchronously) all the images associated to the profile. The each image will be recomputed for the new profile when being queried for the first time.

## Images
**Retrieve an image** associated to an *entityId* for a given *profile*:

```http://{img20sur20_baseURL}/image/{entityId}/{profileName} | GET```

If the image does not exist for the requested profile, it will be computed on the fly, provided an *original* image is associated to the *entityId*

**Delete an image** for a given profile:

```http://{img20sur20_baseURL}/image/{entityId}/{profileName} | DELETE```

**Delete all the images** (all profiles) associated to an *entityId*:

```http://{img20sur20_baseURL}/image/{entityId} | DELETE```

Note that this kind of deletion is done asynchronously using task queues. This means that the image can still be accessible for a short period.
