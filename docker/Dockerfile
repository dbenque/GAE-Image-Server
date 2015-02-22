#
#    Do no use cache in order to get fresh github sources
#    docker build --no-cache -t "gaeimageserver" .
#
FROM dbenque/goappengine

MAINTAINER dbenque

# Get the project
RUN go get -d github.com/dbenque/GAE-Image-Server

# The GoAppengine project must not be under GOPATH : https://cloud.google.com/appengine/docs/go/#Go_Organizing_Go_apps
RUN mv $GOPATH/src/github.com/dbenque/GAE-Image-Server ~/

# cd first because cannot serve an absolute path
CMD goapp serve --host=0.0.0.0 ~/GAE-Image-Server
