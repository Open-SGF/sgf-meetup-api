# SGF Meetup API

## Prerequisites
- [Node 18.x](https://nodejs.org) (Ideally using [nvm](https://github.com/nvm-sh/nvm))
- [Docker Desktop](https://www.docker.com/products/docker-desktop/)
- [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html)
- [AWS SAM CLI](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/install-sam-cli.html)
- An Open SGF AWS Account
  - Message a project organizer to get one set up 

## First Time Setup

### Sign in to the AWS CLI
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

### Set Default AWS Profile (Optional)
Without this you'll need to manually specify your profile for any aws commands.
Including the commands in npm scripts.
- Set the `AWS_DEFAULT_PROFILE` environment variable to the profile name you just created
  - `export AWS_DEFAULT_PROFILE=<your name>-opensgf`
- Don't forget to reload your shell after setting this!

### Install dependencies
```bash
nvm install # if using nvm
npm install
```

### Create `.env`
```bash
cp .env.example .env
```

## Running the project
- `nvm use` (if using nvm)
- `docker compose up -d`
- `npm run dev:sync-dynamodb`
- Run importer script
  - `npm run dev:importer`
- Run API
  - `npm run dev:api`

## Shutting down
- `docker compose down`

## Troubleshooting

### `SSOTokenProviderFailure` When Starting Project
If it's been awhile since you've last run the project, your SSO session in the AWS CLI has expired.
To fix it:
- `aws sso login`
- Open the link in a browser and follow the prompts

### `npm run dev:sync-dynamodb` Hangs in Linux Environments (Including WSL)
This can be caused by permissions errors with the `./docker` folder that docker compose creates.
To fix it change the permissions of that folder to your local user
```bash
sudo chown $USER ./docker -R
```