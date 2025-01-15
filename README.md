# SGF Meetup API

The SGF Meetup API lists Meetup event details for local tech groups in the Springfield, MO area.

## Table of Contents
- [How to Use This API](#how-to-use-this-api)
	- [Requesting an API Key](#requesting-an-api-key)
	- [Authentication](#authentication)
- [For Contributors](#for-contributors)
	- [Prerequisites](#prerequisites)
	- [First Time Setup](#first-time-setup)
	- [Running the Project](#running-the-project)
	- [Shutting Down](#shutting-down)
	- [Troubleshooting](#troubleshooting)

## How to Use This API

### Requesting an API Key

Request an API key by opening a GitHub Issue in the [SGF Meetup API repo](https://github.com/Open-SGF/sgf-meetup-api/issues/).

In the GitHub Issue, submit a username for the API and contact information.  We will assign a password to the username and send it to the contact information listed.

### Authentication



## For Contributors

### Prerequisites
- [Node 18.x](https://nodejs.org) (Ideally using [nvm](https://github.com/nvm-sh/nvm))
- [Docker Desktop](https://www.docker.com/products/docker-desktop/)
- [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html)
- [AWS SAM CLI](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/install-sam-cli.html)
- An Open SGF AWS Account
  - Message a project organizer to get one set up 

### First Time Setup

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
```

#### Create `.env`
```bash
cp lambdas/.env.example lambdas/.env
```

#### Populate additional env variables
- `MEETUP_GROUP_NAMES` should be a comma seperated list of meetup group names to import events from
  - This value can be pulled from the url of a Meetup groups page e.g. with meetup.com/sgfdevs, sgfdevs is the group name
- `API_KEYS` should be a comma seperated list of strings you'll use to locally call the API

#### Setup initial database
- `docker compose up -d`
  - If you are on linux or WSL there will likely be permission issues with a folder used by docker
  - You can fix those with `sudo chown -R 1000:1000 docker/dynamodb`
- `npm run dev:sync-dynamodb`

### Running the project
- `docker compose up -d` (if not already running)
- `nvm use` (if using nvm)
- Run importer script
  - `npm run dev:importer`
- Run API
  - `npm run dev:api`
  - Use an API client (like [Postman](https://www.postman.com/)) to send requests to `http://localhost/events`
    - You'll need to make sure the `Authorization` header is set to one of your `API_KEYS` from your `.env`

### Shutting down
- `docker compose down`

### Troubleshooting

#### `SSOTokenProviderFailure` When Starting Project
If it's been awhile since you've last run the project, your SSO session in the AWS CLI has expired.
To fix it:
- `aws sso login`
- Open the link in a browser and follow the prompts

#### `npm run dev:sync-dynamodb` Hangs in Linux Environments (Including WSL)
This can be caused by permissions errors with the `./docker` folder that docker compose creates.
To fix it change the permissions of that folder to your local user
```bash
sudo chown $USER ./docker -R
```

### `npm run dev:sync-dynamodb` Causes "UnrecognizedClientException: The security token included in the request is invalid."
Specify a DYNAMODB_ENDPOINT environment variable pointing to localhost.

Example: `DYNAMODB_ENDPOINT=http://localhost:8000 npm run dev:sync-dynamodb`
