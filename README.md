GAE-Image-Server
================

Use GAE Blobstore to host your images. Upload image and profiles (Quality/Size). Serve Image for a given profile (resize on the fly). Note that Quality indicator does not apply in case of png image. In that case the png format is used for storage to keep transparency. In other cases, file will be saved as jpeg.

The application is written in GO and should be deployed to GAE. As the blostore requires a *callback url*, to ease local testing, you will find a Dockerfile that will allow you to run locally an AppEngine instance configured for the project. It can also be used to deploy the application.

Prerequisit, is to build the dbenque/goappengine base image. To do so:

``` go get github.com/dbenque/goAppengineToolkit ```

Enter the *docker* folder and:

``` dbenque/goappengine ```

Then build your image, go to the *docker* folder and type

``` docker build --no-cache -t "gaeimageserver" .```

Once the image is built, you can test it locally:

```docker run -p 127.0.0.1:8080:8080 -p 127.0.0.1:8000:8000 -p 127.0.0.1:9000:9000 gaeimageserver```

If you want to deploy the project on your AppEngine instance, create an application app_id='YOURPROJECTNAME' under the appengine console ( https://appengine.google.com/ ) and then:

```docker run -t -i gaeimageserver /bin/bash```

Modify the name of the project in app.yaml:

```sed -i 's/gae-image-server/YOURPROJECTNAME/' /home/GAE-Image-Server/app.yaml```

Deploy to AppEngine:

```goapp deploy /home/GAE-Image-Server/```

The dependencies are (included in the built docker image):

https://github.com/gorilla/mux

https://github.com/gorilla/context


## Upload
To get the "url action" to put in the upload form:

```http://{img_baseURL}/action | GET```

Example of form to be used in the upload page:

```http://{baseURL}/exampleForm | GET```

Of course the form will have to post the file to be uploaded, but note also the 2 important fields  that help to index the image and continue the flow:
```
<input type="text" name="entityId" value="myId">
<input type="text" name="callbackurl" value="http://0.0.0.0:8080/callbacktest">
```

Once the form is posted to the blobstore action, the file will be stored, and the *callbackurl* will be invoked with the updated form. All the input parameters of the initial form will be transfered.

## Image Profiles
**Retrieve the list** of existing profiles:

```http://{baseURL}/imgProfiles | GET```

This returns a JSON flat list containing the list of profile names:

```
[
"original",      // This profil is created by default, and  is associated by default to any uploaded image (+resize if needed)
"small"
]
```

**Retrieve** a specific profile:

```http://{baseURL}/imgProfile/{profileName} | GET```

**Delete** a specific profile:

```http://{baseURL}/imgProfile/{profileName} | DELETE```

This removes all the files associated to the profile.

**Create** or **update** a specific profile:
```
http://{baseURL}/imgProfile | POST | JSON {
        "name" : "small",
        "quality" : 50,
        "maxSize" : 100
}
```

Note that to update a profile you need to post a JSON with the name associated to the profile to update. So far a modification of the profile does dot trig a resizing/recomputation of the images associated to the profile. To force the resizing/recomputation, you need to delete the profile. This will remove (asynchronously) all the images associated to the profile. The each image will be recomputed for the new profile when being queried for the first time.

## Images
**Retrieve an image** associated to an *entityId* for a given *profile*:

```http://{baseURL}/image/{entityId}/{profileName} | GET```

If the image does not exist for the requested profile, it will be computed on the fly, provided an *original* image is associated to the *entityId*

**Delete an image** for a given profile:

```http://{baseURL}/image/{entityId}/{profileName} | DELETE```

**Delete all the images** (all profiles) associated to an *entityId*:

```http://{baseURL}/image/{entityId} | DELETE```

Note that this kind of deletion is done asynchronously using task queues. This means that the image can still be accessible for a short period.
