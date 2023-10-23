import 'dotenv/config';
import { ScanCommand } from '@aws-sdk/client-dynamodb';
import { dynamoDBClient } from './dynamodb-client';

async function getAll() {
	const tableName = 'items';
	const command = new ScanCommand({ TableName: tableName });

	const response = await dynamoDBClient.send(command);
	return response.Items;
}

export const handler = async () => {
	const result = await getAll();
	return {
		statusCode: 200,
		body: JSON.stringify({ hello: 'VERLD', result }),
	};
};
