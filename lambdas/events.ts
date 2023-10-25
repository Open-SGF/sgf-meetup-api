import 'dotenv/config';
import { type Handler } from 'aws-lambda';
import { QueryCommand, AttributeValue } from '@aws-sdk/client-dynamodb';

import { dynamoDbClient } from './lib/dynamoDbClient';
import { Group, Venue, Node as MeetupEvent } from './types/MeetupFutureEventsPayload';
import { parseDateString } from './lib/util';

const EVENTS_TABLE_NAME = process.env['EVENTS_TABLE_NAME'];
const EVENTS_GROUP_INDEX_NAME = process.env['EVENTS_GROUP_INDEX_NAME'];

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
	console.log({ queryCommand });
	const response = await dynamoDbClient.send(queryCommand);

	const events =
		response.Items?.map((item) => {
			const group = {
				name: item.MeetupGroupName.S!,
				urlname: item.MeetupGroupUrl.S!
			} satisfies Group;

			const title = item.Title.S!;
			const eventUrl = item.EventUrl.S!;
			const description = item.Description.S!;
			const dateTime = item.EventDateTime.S!;
			const duration = item.Duration.S!;
			const venue = {
				name: item.VenueName.S!,
				address: item.VenueAddress.S!,
				city: item.VenueCity.S!,
				state: item.VenueState.S!,
				postalCode: item.VenuePostalCode.S!,
			};

			const meetupEvent: MeetupEvent = {
				group,
				title,
				eventUrl,
				description,
				dateTime,
				duration,
				venue,
				host: {
					name: item.HostName.S!
				},
				images: [] // TODO
			};

			return meetupEvent;
		});

	return events ?? [];
}

/**
 * Make a string to be used as `KeyConditionExpression` in a QueryCommand
 */
function makeKeyConditionExpression({
	before,
	after,
}: GetMeetupEventsOptions): string {
	let expr = 'MeetupGroupName = :group';

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
	queryStringParameters: Record<string, string>,
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

export const handler: Handler = async (event, context) => {
	const { queryStringParameters } = event;

	try {
		const getMeetupEventsOptions = makeGetMeetupEventsOptions(
			queryStringParameters,
		);
		const events = await getMeetupEvents(getMeetupEventsOptions);
		const body = JSON.stringify({ success: true, events });
		return { statusCode: 200, body };
	} catch (error: any) {
		console.error(error);
		const body = JSON.stringify({
			success: false,
			error,
			errorName: error?.name,
			errorMessage: error?.message,
		});
		return { statusCode: 500, body };
	}
};