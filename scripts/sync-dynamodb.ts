import 'dotenv/config';
import {
  DynamoDBClient,
  CreateTableCommand,
  PutItemCommand,
  CreateTableCommandInput,
  PutItemCommandInput
} from '@aws-sdk/client-dynamodb'
import * as template from '../cdk.out/sgf-meetup-api.template.json';

const client = new DynamoDBClient({
  endpoint: process.env.DYNAMODB_ENDPOINT,
  credentials: {
    accessKeyId: 'user',
    secretAccessKey: 'password',
  },
});

async function syncDb(): Promise<void> {
  await syncTables();
  await populateTestData();
}

/**
 * For each DynamoDB table resource defined in the given CloudFormation
 * template, run a CreateTableCommand to create the table in the local DynamoDB
 */
async function syncTables(): Promise<void> {
  // Look at each resource in the template
  for (const [resourceKey, resourceValue] of Object.entries(template.Resources)) {
    if (resourceValue.Type !== "AWS::DynamoDB::Table") {
      continue; // If resource is not a Table, skip it
    }

    // Make a CreateTableCommand
    const createTableParams = resourceValue.Properties as CreateTableCommandInput;

    const createTableCommand = new CreateTableCommand(createTableParams);
    try {
      // Send the CreateTableCommand
      await client.send(createTableCommand);
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
    TableName: "items",
    Item: {
      itemId: { S: Math.floor(Math.random() * 100000000).toString() },
    },
  } satisfies PutItemCommandInput;

  const putCommand = new PutItemCommand(putParams);
  await client.send(putCommand);
}

void syncDb();
