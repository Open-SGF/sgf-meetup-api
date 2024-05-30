import { config } from 'dotenv';
import { expand } from 'dotenv-expand';
import {
	DynamoDBClient,
	CreateTableCommand,
	PutItemCommand,
	CreateTableCommandInput,
	PutItemCommandInput,
	ListTablesCommand,
} from '@aws-sdk/client-dynamodb';
import * as template from '../cdk.out/sgf-meetup-api.template.json';

const env = config({ path: 'lambdas/.env' });
expand(env);

const client = new DynamoDBClient({
	endpoint: process.env.DYNAMODB_ENDPOINT,
	region: process.env.DYNAMODB_REGION,
	credentials: {
		accessKeyId: 'user',
		secretAccessKey: 'password',
	},
});

async function syncDb(): Promise<void> {
	await syncTables();
	if (false) {
		await populateTestData();
	}
}

/**
 * For each DynamoDB table resource defined in the given CloudFormation
 * template, run a CreateTableCommand to create the table in the local DynamoDB
 */
async function syncTables(): Promise<void> {
	const listTablesCommand = new ListTablesCommand({});

	const listTablesResult = await client.send(listTablesCommand);

	console.log(listTablesResult);

	// Look at each resource in the template
	for (const [resourceKey, resourceValue] of Object.entries(
		template.Resources,
	)) {
		if (resourceValue.Type !== 'AWS::DynamoDB::Table') {
			continue; // If resource is not a Table, skip it
		}

		// Make a CreateTableCommand
		const createTableParams =
			resourceValue.Properties as CreateTableCommandInput;

		const tableName = createTableParams.TableName;

		if (tableName === undefined) {
			continue;
		}

		if (listTablesResult.TableNames?.includes(tableName)) {
			console.log(`Table ${tableName} already exists, skipping`);
			continue;
		}

		const createTableCommand = new CreateTableCommand(createTableParams);
		try {
			// Send the CreateTableCommand
			await client.send(createTableCommand);
			console.log(`Table ${tableName} created`);
		} catch (err) {
			// eslint-disable-next-line no-console
			console.error(`Error when creating table resource ${resourceKey}:`);
			// eslint-disable-next-line no-console
			console.error(err);
		}
	}
}

/**
 * Add an event with a randomly generated ID to the Events table
 */
async function populateTestData() {
	for (let groupIndex = 0; groupIndex < 5; groupIndex += 1) {
		const groupNumber = Math.floor(Math.random() * 1000);

		for (let i = 0; i < 5; i += 1) {
			const itemIndex = groupNumber + i;
			const randomOffset = Math.floor(Math.random() * 100 - 50);

			const startTime = new Date();

			// Add `randomOffset` days to today
			startTime.setDate(startTime.getDate() + randomOffset);

			const putParams = {
				TableName: 'Events',
				Item: {
					Id: { S: itemIndex.toString() },
					MeetupGroupName: { S: 'group' + groupNumber },
					MeetupGroupUrlName: { S: 'group-url' + groupNumber },
					Title: { S: 'title' + itemIndex },
					EventUrl: { S: 'eventUrl' + itemIndex },
					Description: {
						S: `random offset was ${randomOffset} days`,
					},
					EventDateTime: { S: startTime.toISOString() },
					Duration: { S: 'duration' + itemIndex },
					VenueName: { S: 'venue-name' + itemIndex },
					VenueAddress: { S: 'venue-address' + itemIndex },
					VenueCity: { S: 'venue-city' + itemIndex },
					VenueState: { S: 'venue-state' + itemIndex },
					VenuePostalCode: { S: 'venue-postcode' + itemIndex },
					HostName: { S: 'host-name' + itemIndex },
				},
			} satisfies PutItemCommandInput;

			const putCommand = new PutItemCommand(putParams);
			const putResult = await client.send(putCommand);
			console.log({ putResult }); // eslint-disable-line no-console
		}
	}
}

void syncDb();
