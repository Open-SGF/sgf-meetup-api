import 'dotenv/config';
import fetch from 'node-fetch';
import {
	AttributeValue,
	PutItemCommand,
	PutItemCommandInput,
} from '@aws-sdk/client-dynamodb';
import { v4 as uuid } from 'uuid';

import { getMeetupToken } from './lib/getMeetupToken';
import {
	MeetupFutureEventsPayload,
	meetupEventToDynamoDbItem,
} from './types/MeetupFutureEventsPayload';
import { dynamoDbClient } from './lib/dynamoDbClient';

const GET_FUTURE_EVENTS = `
  query ($urlname: String!, $itemsNum: Int!, $cursor: String) {
	events: groupByUrlname(urlname: $urlname) {
	  unifiedEvents(input: { first: $itemsNum, after: $cursor }) {
		count
		pageInfo {
		  endCursor
		  hasNextPage
		}
		edges {
		  node {
			id
			title
			eventUrl
			description
			dateTime
			duration
			venue {
			  name
			  address
			  city
			  state
			  postalCode
			}
			group {
			  name
			  urlname
			}
			host {
			  name
			}
			images {
			  baseUrl
			  preview
			}
		  }
		}
	  }
	}
  }
`;

interface ImportErrorRecord {
	errorName: string;
	errorMessage: string;
	errorStack?: string;
	groupName?: string;
}

export interface ImportLogRecord {
	successGroupNames: string[];
	failedGroupNames: string[];
	start: Date;
	end: Date;
	eventCount: number;
	errors: ImportErrorRecord[];
}

async function writeImportLog({
	successGroupNames,
	failedGroupNames,
	start,
	end,
	eventCount,
	errors,
}: ImportLogRecord): Promise<void> {
	const id = uuid();

	const item: Record<string, AttributeValue> = {
		Id: { S: id },
		SuccessGroupNames: {
			L: successGroupNames.map((name) => ({ S: name })),
		},
		FailedGroupNames: { L: failedGroupNames.map((name) => ({ S: name })) },
		StartedAt: { S: start.toISOString() },
		FinishedAt: { S: end.toISOString() },
		TotalEventsSaved: { N: eventCount.toString() },
		Errors: {
			L: errors.map((error) => {
				return {
					M: {
						Name: { S: error.errorName },
						Message: { S: error.errorMessage },
						Stack: { S: error.errorStack ?? '' },
						GroupName: { S: error.groupName ?? '' },
					},
				};
			}),
		},
	};

	const putParams = {
		TableName: process.env.IMPORTER_LOG_TABLE_NAME,
		Item: item,
	} satisfies PutItemCommandInput;

	const putCommand = new PutItemCommand(putParams);
	const putResult = await dynamoDbClient.send(putCommand);

	console.log({ writeImportLogResult: putResult }); // eslint-disable-line no-console
}

async function importEventsToDynamoDb(
	meetupAccessToken: string,
): Promise<void> {
	const meetupGraphQlEndpoint = 'https://api.meetup.com/gql';
	const batchSize = 10; // Number of events to fetch in each batch

	const GROUP_NAMES = (
		process.env.MEETUP_GROUP_NAMES?.split(',').map(
			(userpass) => userpass.split(':')[0],
		) ?? []
	).map((group) => group.trim());

	if (GROUP_NAMES.length === 0) {
		throw new Error('No groups specified in environment variable');
	}

	async function fetchAllFutureEvents(
		urlname: string,
		cursor: string | null = null,
	) {
		const requestBody = JSON.stringify({
			query: GET_FUTURE_EVENTS,
			variables: {
				urlname,
				itemsNum: batchSize,
				cursor,
			},
		});

		const requestOptions = {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json',
				Authorization: `Bearer ${meetupAccessToken}`,
			},
			body: requestBody,
		};

		try {
			const response = await fetch(meetupGraphQlEndpoint, requestOptions);
			const res = (await response.json()) as MeetupFutureEventsPayload;
			const unifiedEvents = res.data.events?.unifiedEvents;
			const events = unifiedEvents?.edges.map((edge) => edge.node) ?? [];

			if (unifiedEvents?.pageInfo.hasNextPage) {
				const nextCursor = unifiedEvents.pageInfo.endCursor;
				const nextEvents = await fetchAllFutureEvents(
					urlname,
					nextCursor,
				);
				events.push(...nextEvents);
			}

			return events;
		} catch (error) {
			// eslint-disable-next-line no-console
			console.error('Error fetching future events:', error);
			return [];
		}
	}

	const successGroupNames = new Array<string>();
	const failedGroupNames = new Array<string>();
	const errors = new Array<ImportErrorRecord>();
	let eventCount = 0;

	async function saveAllFutureEvents() {
		for (const groupName of GROUP_NAMES) {
			try {
				// eslint-disable-next-line no-console
				console.log('fetching for', groupName);
				const futureEvents = await fetchAllFutureEvents(groupName);
				// eslint-disable-next-line no-console
				console.log({ futureEvents });

				const eventsAsDynamoDbItems = futureEvents.map((event) =>
					meetupEventToDynamoDbItem(event),
				);

				for (const item of eventsAsDynamoDbItems) {
					const putParams = {
						TableName: process.env.EVENTS_TABLE_NAME,
						Item: item,
					} satisfies PutItemCommandInput;
					const putCommand = new PutItemCommand(putParams);
					const putResult = await dynamoDbClient.send(putCommand);
					eventCount += 1;
					console.log({ putResult }); // eslint-disable-line no-console
				}

				successGroupNames.push(groupName);
			} catch (err) {
				failedGroupNames.push(groupName);
				// eslint-disable-next-line no-console
				console.error(`Failed to save events for ${groupName}`);
				console.error(err); // eslint-disable-line no-console

				if (err instanceof Error) {
					errors.push({
						errorName: err.name,
						errorMessage: err.message,
						errorStack: err.stack,
					});
				}
			}
		}
	}

	const start = new Date();
	await saveAllFutureEvents();
	const end = new Date();
	await writeImportLog({
		successGroupNames,
		failedGroupNames,
		start,
		end,
		eventCount,
		errors,
	});
}

export async function handler() {
	const token = await getMeetupToken();
	await importEventsToDynamoDb(token);
}
