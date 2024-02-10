import 'dotenv/config';
import fetch from 'node-fetch';
import { PutItemCommand, PutItemCommandInput } from '@aws-sdk/client-dynamodb';

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

async function importEventsToDynamoDb(meetupAccessToken: string) {
	const meetupGraphQlEndpoint = 'https://api.meetup.com/gql';
	const batchSize = 10; // Number of events to fetch in each batch

	const GROUP_URLNAMES = (
		process.env.MEETUP_GROUP_NAMES?.split(',').map(
			(userpass) => userpass.split(':')[0],
		) ?? []
	).map((group) => group.trim());

	if (GROUP_URLNAMES.length === 0) {
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
			const events =
				res.data.events?.unifiedEvents.edges.map((edge) => edge.node) ??
				[];

			if (res.data.events?.unifiedEvents.pageInfo.hasNextPage) {
				const nextCursor =
					res.data.events.unifiedEvents.pageInfo.endCursor;
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

	async function saveAllFutureEvents() {
		for (const urlname of GROUP_URLNAMES) {
			// eslint-disable-next-line no-console
			console.log('fetching for', urlname);
			const futureEvents = await fetchAllFutureEvents(urlname);
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
				console.log({ putResult }); // eslint-disable-line no-console
			}
		}
	}

	await saveAllFutureEvents();
}

export async function handler() {
	const token = await getMeetupToken();
	await importEventsToDynamoDb(token);
}
