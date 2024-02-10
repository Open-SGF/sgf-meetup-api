import 'dotenv/config';
import type {
	APIGatewayEvent,
	APIGatewayProxyEventQueryStringParameters,
	Handler,
} from 'aws-lambda';
import { QueryCommand, AttributeValue } from '@aws-sdk/client-dynamodb';

import { dynamoDbClient } from './lib/dynamoDbClient';
import {
	MeetupEvent,
	meetupEventFromDynamoDbItem,
} from './types/MeetupFutureEventsPayload';
import { parseDateString } from './lib/util';

const EVENTS_TABLE_NAME = process.env.EVENTS_TABLE_NAME;
const EVENTS_GROUP_INDEX_NAME = process.env.EVENTS_GROUP_INDEX_NAME;

type GetMeetupEventsOptions = {
	count: number;
	page: number;
	group: string;
	before?: Date;
	after?: Date;
};

/**
 * Query the Events table
 *
 * @returns array of MeetupEvents that match the query
 */
async function getMeetupEvents(
	options: GetMeetupEventsOptions,
): Promise<MeetupEvent[]> {
	const queryCommand: QueryCommand = new QueryCommand({
		TableName: EVENTS_TABLE_NAME,
		IndexName: EVENTS_GROUP_INDEX_NAME,
		KeyConditionExpression: makeKeyConditionExpression(options),
		ExpressionAttributeValues: makeExpressionAttributeValues(options),
		Limit: options.count,
	});
	console.log({ queryCommand }); // eslint-disable-line no-console
	const response = await dynamoDbClient.send(queryCommand);

	const events = response.Items?.map((item) =>
		meetupEventFromDynamoDbItem(item),
	);

	return events ?? [];
}

/**
 * Make a string to be used as `KeyConditionExpression` in a QueryCommand
 */
function makeKeyConditionExpression({
	before,
	after,
}: GetMeetupEventsOptions): string {
	let expr = 'MeetupGroupUrlName = :group';

	if (before && after) {
		expr += ' AND EventDateTime BETWEEN :after AND :before';
	} else if (before !== undefined) {
		expr += ' AND EventDateTime <= :before';
	} else if (after !== undefined) {
		expr += ' AND EventDateTime >= :after';
	}

	return expr;
}

/**
 * Make an object to be used as `ExpressionAttributeValues` in a QueryCommand
 */
function makeExpressionAttributeValues({
	group,
	before,
	after,
}: GetMeetupEventsOptions): Record<string, AttributeValue> {
	const vals: Record<string, AttributeValue> = {
		':group': { S: group },
	};

	if (before !== undefined) {
		vals[':before'] = { S: before.toISOString() };
	}

	if (after !== undefined) {
		vals[':after'] = { S: after.toISOString() };
	}

	return vals;
}

/**
 * Take query string parameters from a request and return a set of parameters to
 * be passed to `getMeetupEvents` to run the query the user is requesting.
 */
function makeGetMeetupEventsOptions(
	queryStringParameters: APIGatewayProxyEventQueryStringParameters,
): GetMeetupEventsOptions {
	// Check for `group` query string parameter
	const groupParam = queryStringParameters?.['group'];
	if (!groupParam) {
		const error = 'The `group` query string parameter is required.';
		throw new Error(error);
	}

	const options: GetMeetupEventsOptions = {
		count: 100,
		page: 0,
		group: groupParam,
	};

	// Check for `next` query string parameter
	const nextParam = queryStringParameters?.['next'];
	if (nextParam !== undefined) {
		// If `next` is present, we want 1 event after the current time
		options.count = 1;
		const now = new Date();
		options.after = now;
		return options; // Don't check other params
	}

	// Check for `before` and `after` query string parameters
	const beforeParam = queryStringParameters?.['before'];
	const afterParam = queryStringParameters?.['after'];

	if (beforeParam !== undefined) {
		options.before = parseDateString(beforeParam);
	}

	if (afterParam !== undefined) {
		options.after = parseDateString(afterParam);
	}

	return options;
}

function validateKey(apiKey: string) {
	const validKeys = process.env.API_KEYS!.split(',');
	return validKeys.includes(apiKey);
}

export const handler: Handler = async (event: APIGatewayEvent) => {
	const { headers, queryStringParameters } = event;

	try {
		const authHeader = headers['Authorization'];

		if (authHeader == null) {
			return {
				statusCode: 401,
				body: JSON.stringify({
					error: 'Authorization header is required',
				}),
			};
		}

		if (!validateKey(authHeader)) {
			// eslint-disable-next-line no-console
			console.error('Bad API key');
			return {
				statusCode: 401,
				body: JSON.stringify({
					error: 'Authorization header is not valid',
				}),
			};
		}

		const getMeetupEventsOptions = makeGetMeetupEventsOptions(
			queryStringParameters!,
		);

		const events = await getMeetupEvents(getMeetupEventsOptions);
		const body = JSON.stringify({ success: true, events });
		return { statusCode: 200, body };
		// eslint-disable-next-line @typescript-eslint/no-explicit-any
	} catch (error: any) {
		console.error(error); // eslint-disable-line no-console
		const body = JSON.stringify({
			success: false,
			error,
			errorName: error?.name,
			errorMessage: error?.message,
		});
		return { statusCode: 500, body };
	}
};
