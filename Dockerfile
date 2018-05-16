FROM golang:latest

# Install packages
RUN curl -sL https://deb.nodesource.com/setup_10.x | bash -
RUN apt-get install -y git nodejs
RUN go get -u github.com/golang/dep/cmd/dep

# Copy in source and install deps
RUN mkdir -p /go/src/github.com/silinternational/speed-snitch-admin-api
COPY ./package.json /go/src/github.com/silinternational/speed-snitch-admin-api/
WORKDIR /go/src/github.com/silinternational/speed-snitch-admin-api
RUN npm install -g serverless && npm install
COPY ./ /go/src/github.com/silinternational/speed-snitch-admin-api/
RUN dep ensure