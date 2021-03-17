FROM golang:latest

# Install packages
RUN curl -sL https://deb.nodesource.com/setup_10.x | bash -
RUN apt-get install -y git nodejs netcat

WORKDIR /src

# Copy in source and install deps
COPY ./package.json .
RUN npm install -g serverless && npm install
COPY ./ .
RUN go get ./...

# Get whenavail for tests
RUN curl -o /usr/local/bin/whenavail https://bitbucket.org/silintl/docker-whenavail/raw/1.0.0/whenavail \
    && chmod a+x /usr/local/bin/whenavail