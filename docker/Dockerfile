FROM google/golang

MAINTAINER dbenque

RUN apt-get update
RUN apt-get -y install procps

# Install cmake
RUN apt-get -y install build-essential cmake

# Install Python 2.x
RUN apt-get -y install python-dev

# Env
ENV HOME /home
WORKDIR /home

RUN curl -G https://storage.googleapis.com/appengine-sdks/featured/go_appengine_sdk_linux_amd64-1.9.14.zip > go_appengine_sdk_linux_amd64-1.9.14.zip 
RUN apt-get install -y unzip
RUN unzip go_appengine_sdk_linux_amd64-1.9.14.zip
RUN rm -rf go_appengine_sdk_linux_amd64-1.9.14.zip

ENV PATH $PATH:/home/go_appengine
ENV GOROOT /home/go_appengine/goroot
ENV GOPATH /home/go_appengine/gopath

RUN ln -s /home/go_appengine/goroot/bin/goapp /home/go_appengine/go

#### dont use localhost but 0.0.0.0
RUN sed -i "s/'localhost'/'0.0.0.0'/g" go_appengine/google/appengine/tools/devappserver2/devappserver2.py
RUN sed -i "s/'--api_port', type=PortParser(), default=0,/'--api_port', type=PortParser(), default=9000,/g" go_appengine/google/appengine/tools/devappserver2/devappserver2.py

EXPOSE 8080
EXPOSE 8000
EXPOSE 9000

# Get Project Dependencies
RUN mkdir -p $GOPATH/src/github.com/gorilla && cd $GOPATH/src/github.com/gorilla && git clone https://github.com/gorilla/mux.git && git clone https://github.com/gorilla/context.git

RUN git clone https://github.com/TomiHiltunen/GAE-Go-image-optimizer.git && cd $GOPATH/src/github.com && ln -s ~/GAE-Go-image-optimizer/github.com/tomihiltunen

# Get the project
RUN git clone https://github.com/dbenque/GAE-Image-Server.git

CMD goapp serve --host=0.0.0.0 ~/GAE-Image-Server
