#
#    Do no use cache in order to get fresh github sources
#    docker build --no-cache -t "gaeimageserver" .
#
FROM dbenque/goappengine

MAINTAINER dbenque

# Get the project
RUN go get github.com/dbenque/GAE-Image-Server

CMD goapp serve --host=0.0.0.0 ~/GAE-Image-Server
