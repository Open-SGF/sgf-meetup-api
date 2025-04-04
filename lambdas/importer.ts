import * as dotenv from 'dotenv';
import fetch from 'node-fetch';
import {
	AttributeValue,
	BatchGetItemCommand,
	BatchWriteItemCommand,
	PutItemCommand,
	PutItemCommandInput,
	ScanCommand,
} from '@aws-sdk/client-dynamodb';
import * as lambda from '@aws-sdk/client-lambda';
import { v4 as uuid } from 'uuid';

dotenv.config({ override: true });

import {
	MeetupEvent,
	MeetupFutureEventsPayload,
	meetupEventFromDynamoDbItem,
	meetupEventToDynamoDbItem,
} from './types/MeetupFutureEventsPayload';
import { dynamoDbClient } from './lib/dynamoDbClient';
import { chunkArray } from './lib/util';

const EVENTS_TABLE_NAME = process.env.EVENTS_TABLE_NAME!;
const OLD_EVENTS_TABLE_NAME = process.env.OLD_EVENTS_TABLE_NAME!;
const GET_MEETUP_TOKEN_FUNCTION_NAME =
	process.env.GET_MEETUP_TOKEN_FUNCTION_NAME!;

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

async function getAllSavedFutureEvents(): Promise<MeetupEvent[]> {
	const allEvents = new Array<MeetupEvent>();
	let lastEvaluatedKey: Record<string, AttributeValue> | undefined;

	/**
	 * Scan the next page of future events and return the LastEvaluatedKey from the response
	 */
	async function scanNextPage(): Promise<
		Record<string, AttributeValue> | undefined
	> {
		const lastCheckedId =
			lastEvaluatedKey === undefined
				? undefined
				: { S: lastEvaluatedKey.Id! };

		const scanCommand: ScanCommand = new ScanCommand({
			TableName: EVENTS_TABLE_NAME,
			ExclusiveStartKey: lastCheckedId,
			FilterExpression:
				'attribute_not_exists(DeletedAtDateTime) AND EventDateTime > :now',
			ExpressionAttributeValues: {
				':now': { S: new Date().toISOString() },
			},
		});

		console.log({ scanCommand }); // eslint-disable-line no-console

		const response = await dynamoDbClient.send(scanCommand);

		const events =
			response.Items?.map((item) => meetupEventFromDynamoDbItem(item)) ??
			[];

		allEvents.push(...events);
		return response.LastEvaluatedKey;
	}

	let done = false;

	while (!done) {
		lastEvaluatedKey = await scanNextPage();
		if (!lastEvaluatedKey) {
			done = true;
		}
	}

	return allEvents;
}

async function deleteEventsById(eventIds: string[]): Promise<void> {
	if (eventIds.length === 0) {
		return;
	}

	const originalEvents = [];
	for (const idsChunk of chunkArray(eventIds, 100)) {
		// BatchGet limit: 100 items
		const batchGet = new BatchGetItemCommand({
			RequestItems: {
				[EVENTS_TABLE_NAME]: {
					Keys: idsChunk.map((id) => ({ Id: { S: id } })),
					ConsistentRead: true,
				},
			},
		});
		const response = await dynamoDbClient.send(batchGet);
		originalEvents.push(...(response.Responses?.[EVENTS_TABLE_NAME] || []));
	}

	for (const eventsChunk of chunkArray(originalEvents, 25)) {
		// BatchWrite limit: 25 items
		const batchWrite = new BatchWriteItemCommand({
			RequestItems: {
				[OLD_EVENTS_TABLE_NAME]: eventsChunk.map((event) => ({
					PutRequest: { Item: event },
				})),
			},
		});
		await dynamoDbClient.send(batchWrite);
	}

	for (const idsChunk of chunkArray(eventIds, 25)) {
		// BatchWrite limit: 25 items
		const batchDelete = new BatchWriteItemCommand({
			RequestItems: {
				[EVENTS_TABLE_NAME]: idsChunk.map((id) => ({
					DeleteRequest: { Key: { Id: { S: id } } },
				})),
			},
		});
		await dynamoDbClient.send(batchDelete);
	}
}

async function importEventsToDynamoDb(
	meetupAccessToken: string,
	deleteMissingFutureEvents: boolean,
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
	async function fetchNext6MonthsOfEvents(
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
			const events =
				unifiedEvents?.edges.map((edge) => {
					edge.node.dateTime = new Date(edge.node.dateTime); // Rewrite string timestamp to Date object
					return edge.node;
				}) ?? [];
			const sixMonthsFromNow = new Date();
			sixMonthsFromNow.setMonth(new Date().getMonth() + 6);
			const hasEventsPast6Months = events.some(
				(event) => event.dateTime >= sixMonthsFromNow,
			);

			if (!hasEventsPast6Months && unifiedEvents?.pageInfo.hasNextPage) {
				const nextCursor = unifiedEvents.pageInfo.endCursor;
				const nextEvents = await fetchNext6MonthsOfEvents(
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

	let preexistingFutureEvents = new Array<MeetupEvent>();

	try {
		console.log('Loading events saved by previous imports...'); // eslint-disable-line no-console
		preexistingFutureEvents = await getAllSavedFutureEvents();
		// eslint-disable-next-line no-console
		console.log(
			`Found ${preexistingFutureEvents.length} preexisting future events`,
		);
	} catch (err) {
		console.error('Unable to get saved events'); // eslint-disable-line no-console
		console.error(err); // eslint-disable-line no-console

		if (err instanceof Error) {
			errors.push({
				errorName: err.name,
				errorMessage: err.message,
				errorStack: err.stack,
			});
		} else {
			errors.push({
				errorName: 'Unknown',
				errorMessage: String(err),
			});
		}
	}

	/**
	 * List of future events which were found by a previous import, but which were not found by this import.
	 * These events should be marked as "deleted".
	 */
	const eventIdsToDelete = new Set(
		preexistingFutureEvents.map((event) => event.id),
	);

	async function saveAllFutureEvents() {
		for (const groupName of GROUP_NAMES) {
			try {
				// eslint-disable-next-line no-console
				console.log('fetching for', groupName);
				const futureEvents = await fetchNext6MonthsOfEvents(groupName);
				// eslint-disable-next-line no-console
				console.log({ futureEvents });

				for (const event of futureEvents) {
					eventIdsToDelete.delete(event.id); // Write down that we don't want to delete this event

					const dynamoDbItem = meetupEventToDynamoDbItem(event);

					const putParams = {
						TableName: EVENTS_TABLE_NAME,
						Item: dynamoDbItem,
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
				} else {
					errors.push({
						errorName: 'Unknown',
						errorMessage: String(err),
					});
				}
			}
		}

		if (deleteMissingFutureEvents) {
			console.log('Cleaning up old events not found by importer...'); // eslint-disable-line no-console

			const ids = [...eventIdsToDelete];
			if (ids.length === 0) {
				console.log('Nothing to clean up'); // eslint-disable-line no-console
			}
			// eslint-disable-next-line no-console
			console.log(
				`Deleting ${ids.length} old events: ${ids.join(', ')}...`,
			);
			await deleteEventsById(ids);
			console.log('Done'); // eslint-disable-line no-console
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
	const invokeGetMeetupTokenCommand = new lambda.InvokeCommand({
		FunctionName: GET_MEETUP_TOKEN_FUNCTION_NAME,
		Payload: JSON.stringify({ clientId: 'importer' }),
		LogType: lambda.LogType.Tail,
	});
	const client = new lambda.LambdaClient();

	const response = await client.send(invokeGetMeetupTokenCommand);
	const payload = Buffer.from(response.Payload!).toString();
	const logs = Buffer.from(response.LogResult!).toString();
	console.log('response from getMeetupToken'); // eslint-disable-line no-console
	console.log({ payload, logs }); // eslint-disable-line no-console

	const token = JSON.parse(payload).token;

	await importEventsToDynamoDb(token, true);
}
