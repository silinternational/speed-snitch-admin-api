# Admin API for Speed Snitch agent

[![Go Report Card](https://goreportcard.com/badge/github.com/silinternational/speed-snitch-admin-api)](https://goreportcard.com/report/github.com/silinternational/speed-snitch-admin-api)

## Dev setup
1. Run `make dep` to have `dep` install all Go dependencies
2. Create a local `aws.credentials` file with format (See `aws.credentials.example`)
3. Run `make deploy` to build and deploy lambda service

Note: You may also want to run `dep ensure` locally to get all Go packages installed for IDE intelligence.