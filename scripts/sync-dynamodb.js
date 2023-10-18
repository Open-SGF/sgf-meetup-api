const fs = require("fs");
const dynamodb = require("@aws-sdk/client-dynamodb");

const DYNAMODB_ENDPOINT = "http://localhost:8000"; // TODO: make configurable
const ACCESS_KEY_ID = "anything";
const SECRET_ACCESS_KEY = "anything";

const client = new dynamodb.DynamoDBClient({
  endpoint: DYNAMODB_ENDPOINT,
  credentials: {
    accessKeyId: ACCESS_KEY_ID,
    secretAccessKey: SECRET_ACCESS_KEY,
  },
});

async function syncDb(_populateTestData = false) {
  // Load and parse the template from cdk.out
  const templateFileContents = fs.readFileSync("cdk.out/springfieldMeetupApi.template.json").toString();
  const template = JSON.parse(templateFileContents);

  await syncTables(template);

  if (_populateTestData) {
    await populateTestEvents();
  }
}

/**
 * For each DynamoDB table resource defined in the given CloudFormation
 * template, run a CreateTableCommand to create the table in the local DynamoDB
 */
async function syncTables(template) {
  let tableResource;

  // Look at each resource in the template
  for (const [resourceKey, resourceValue] of Object.entries(template.Resources)) {
    if (resourceValue.Type !== "AWS::DynamoDB::Table") {
      continue; // If resource is not a Table, skip it
    }

    // Make a CreateTableCommand
    const createTableParams = resourceValue.Properties;
    const createTableCommand = new dynamodb.CreateTableCommand(createTableParams);
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

async function populateTestEvents() {
  const groupNumber = Math.floor(Math.random() * 1000);

  for (let i = 0; i < 5; i += 1) {
    const itemIndex = groupNumber + i;
    const randomOffset = Math.floor(Math.random() * 100 - 50); 

    const startTime = new Date();

    // Add `randomOffset` days to today
    startTime.setDate(startTime.getDate() + randomOffset);

    const putParams = {
      TableName: "Events",
      Item: {
        Id: { S: itemIndex },
        MeetupGroup: { S: "group" + groupNumber },
        Title: { S: "title" + itemIndex },
        EventUrl: { S: "eventUrl" + itemIndex },
        Description: { S: `random offset was ${randomOffset} days` },
        StartTime: { S: startTime.toISOString() },
        Duration: { S: "duration" + itemIndex },
      },
    };

    const putCommand = new dynamodb.PutItemCommand(putParams);
    const putResult = await client.send(putCommand);
    console.log({ putResult });
  }
}



syncDb(true);
