#
#    Do no use cache in order to get fresh github sources
#    docker build --no-cache -t "gaeimageserver" .
#
FROM dbenque/goappengine

MAINTAINER dbenque

# Get the project
RUN go get -d github.com/dbenque/GAE-Image-Server

# cd first because cannot serve an absolute path
CMD cd $GOPATH/src/github.com/dbenque/ && goapp serve --host=0.0.0.0 GAE-Image-Server
