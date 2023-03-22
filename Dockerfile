FROM golang:1.20

# Install packages
RUN curl -sL https://deb.nodesource.com/setup_16.x | bash -
RUN apt-get install -y nodejs netcat

# Get whenavail for tests
RUN curl -o /usr/local/bin/whenavail https://bitbucket.org/silintl/docker-whenavail/raw/1.0.0/whenavail \
    && chmod a+x /usr/local/bin/whenavail

# Install Serverless
RUN npm --no-fund install -g serverless@3.7

# Copy in source and install deps
WORKDIR /src
COPY ./package.json .
RUN npm --no-fund install
COPY ./ .
RUN go get ./...
