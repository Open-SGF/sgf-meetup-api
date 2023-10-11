import fs from 'fs';
import dynamodb from '@aws-sdk/client-dynamodb';

const DYNAMODB_ENDPOINT = 'http://localhost:8000'; // TODO: make configurable
const ACCESS_KEY_ID = 'anything';
const SECRET_ACCESS_KEY = 'anything';

const client = new dynamodb.DynamoDBClient({
  endpoint: DYNAMODB_ENDPOINT,
  credentials: {
    accessKeyId: ACCESS_KEY_ID,
    secretAccessKey: SECRET_ACCESS_KEY,
  },
});

async function syncDb() {
  // Load and parse the template from cdk.out
  const templateFileContents = fs
    .readFileSync('cdk.out/springfieldMeetupApi.template.json')
    .toString();
  const template = JSON.parse(templateFileContents);

  await syncTables(template);
  await populateTestData();
}

/**
 * For each DynamoDB table resource defined in the given CloudFormation
 * template, run a CreateTableCommand to create the table in the local DynamoDB
 */
async function syncTables(template: any) {
  let tableResource: any;

  // Look at each resource in the template
  for (const [resourceKey, resourceValue] of Object.entries(
    template.Resources,
  )) {
    if (resourceValue.Type !== 'AWS::DynamoDB::Table') {
      continue; // If resource is not a Table, skip it
    }

    // Make a CreateTableCommand
    const createTableParams = resourceValue.Properties;
    const createTableCommand = new dynamodb.CreateTableCommand(
      createTableParams,
    );
    console.log({ resourceKey, createTableParams });

    try {
      // Send the CreateTableCommand
      const createTableResult = await client.send(createTableCommand);
      console.log({ createTableResult });
    } catch (err) {
      console.error(`Error when creating table resource ${resourceKey}:`);
      console.error(err);
    }
  }
}

/**
 * Add a simple item with a randomly generated ID to the items table
 */
async function populateTestData() {
  const putParams = {
    TableName: 'items',
    Item: {
      itemId: { S: Math.floor(Math.random() * 100000000) },
    },
  };

  const putCommand = new dynamodb.PutItemCommand(putParams);
  const putResult = await client.send(putCommand);
  console.log({ putResult });
}

syncDb();
