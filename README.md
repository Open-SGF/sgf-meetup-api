# SGF Meetup API

[![codecov](https://codecov.io/gh/Open-SGF/sgf-meetup-api/graph/badge.svg?token=1OICD52I00)](https://codecov.io/gh/Open-SGF/sgf-meetup-api)

The SGF Meetup API lists Meetup event details for local tech groups in the Springfield, MO area.

## Table of Contents
- [How to Use This API](#how-to-use-this-api)
    - [URLs](#urls)
	- [Documentation](#documentation)
	- [Requesting Credentials](#requesting-credentials)
- [Architecture](#architecture)
- [Contributing](#contributing)
	- [First Time Setup](#first-time-setup)
	- [Running the Project](#running-the-project)
    - [Running Tests](#running-testsc)
	- [Shutting Down](#shutting-down)
	- [Troubleshooting](#troubleshooting)
    - [Project Structure](#project-structure)

## How to Use This API

### URLs

- Production: https://sgf-meetup-api.opensgf.org
- Staging: https://staging-sgf-meetup-api.opensgf.org

### Documentation

- Swagger playground: [/swagger/index.html](https://staging-sgf-meetup-api.opensgf.org/swagger/index.html)
- OpenAPI document: [/swagger/doc.json](https://staging-sgf-meetup-api.opensgf.org/swagger/doc.json)

### Requesting Credentials

Request API credentials by opening a GitHub Issue in the [SGF Meetup API repo](https://github.com/Open-SGF/sgf-meetup-api/issues/).

In the GitHub Issue, submit a username for the API and contact information.  We will assign a password to the username and send it to the contact information listed.

## Architecture

See [docs/architecture.md](./docs/architecture.md)

## Contributing

### First Time Setup

#### Required Tools
- [Go 1.24](https://go.dev/dl/)
- [Node 22.x](https://nodejs.org) (Ideally using [nvm](https://github.com/nvm-sh/nvm))
- [Docker Desktop](https://www.docker.com/products/docker-desktop/)
- [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html)
- [AWS SAM CLI](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/install-sam-cli.html)
- An Open SGF AWS Account
	- Message a project organizer to get one set up

#### Sign in to the AWS CLI
- `aws configure sso`
  - SSO session name: `<your name>-opensgf`
  - SSO start URL: `https://opensgf.awsapps.com/start`
  - SSO region: `us-east-2`
  - SSO registration scopes: use the default (push enter)
- Open the link in a browser
  - Verify/enter the code from the terminal
  - Allow access
- Continue configuration in the terminal
  - CLI default client Region: `us-east-2`
  - CLI default output format: use the default (push enter)
  - CLI profile name: `<your name>-opensgf`

#### Set Default AWS Profile (Optional)
Without this you'll need to manually specify your profile for any aws commands.
Including the commands in npm scripts.
- Set the `AWS_DEFAULT_PROFILE` environment variable to the profile name you just created
  - `export AWS_DEFAULT_PROFILE=<your name>-opensgf`
- Don't forget to reload your shell after setting this!

#### Install dependencies
```bash
nvm install # if using nvm
npm install
go mod download
```

#### Create `.env` and `.lambda-env.json`

`.lambda-env.json` is used when running the application with the SAM cli.

`.env` is used for other instances where programs run directly on a developers machine. 

```bash
cp .env.example .env
cp .lambda-env.json.example .lambda-env.json
```

#### Populate additional env variables
- `MEETUP_GROUP_NAMES` should be a comma seperated list of meetup group names to import events from
  - This value can be pulled from the url of a Meetup groups page e.g. with meetup.com/sgfdevs, sgfdevs is the group name

#### Database/User Setup
- `docker compose up -d`
- `go run ./cmd/syncdynamodb`
- `go run ./cmd/upsertuser -clientId <ID> -clientSecret <SECRET>`

### Running the project
- `docker compose up -d` (if not already running)
- Run importer script
  - `go run ./cmd/localsamrunner importer`
- Run API
  - `go run ./cmd/localsamrunner api`
- Open Swagger docs
  - [http://localhost:3000/swagger/index.html](http://localhost:3000/swagger/index.html)

> **Note:** Valid AWS creds must be present to run either of the above commands.
The easiest way to handle this would be to have a valid aws profile and add the `--profile <profile name>` to the above commands

### Running Tests
- Ensure docker is running
- `go test ./cmd/... ./pkg/...`

### Shutting down
- CTRL + C to shut down API
- `docker compose down`

### Troubleshooting

#### `SSOTokenProviderFailure` When Starting Project
If it's been awhile since you've last run the project, your SSO session in the AWS CLI has expired.
To fix it:
- `aws sso login`
- Open the link in a browser and follow the prompts
