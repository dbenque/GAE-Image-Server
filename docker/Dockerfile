#
#    Do no use cache in order to get fresh github sources
#    docker build --no-cache -t "gaeimageserver" .
#
FROM dbenque/goappengine
MAINTAINER dbenque

# Setting the MODULES for "go get" in this image and run from CMD in dbenque/goappengine
# modules must be separeted by space
ENV MODULES="github.com/dbenque/GAE-Image-Server"

# Get the project
RUN /home/goGetModules.sh
