import 'dotenv/config';
import fetch from 'node-fetch';
import { getMeetupToken } from './lib/getMeetupToken';
import { MeetupFutureEventsPayload } from './types/MeetupFutureEventsPayload';

async function fetchEvents(meetupAccessToken: string) {
	const meetupGraphQlEndpoint = 'https://api.meetup.com/gql';
	const batchSize = 10; // Number of events to fetch in each batch

	const GROUP_URLNAMES = process.env['MEETUP_GROUP_URLNAMES']?.split(',');

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
			const events = res.data.events?.unifiedEvents.edges ?? [];

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

	async function fetchAndPrintAllFutureEvents() {
		if (GROUP_URLNAMES === undefined) {
			throw new Error('Group Urlnames is undefined');
		}

		for (const urlname of GROUP_URLNAMES) {
			const futureEvents = await fetchAllFutureEvents(urlname);

			futureEvents.forEach((event) => {
				// eslint-disable-next-line no-console
				console.log('--------------');
				// eslint-disable-next-line no-console
				console.log('Event:', event.node.title);
				// eslint-disable-next-line no-console
				console.log('Event Link:', event.node.eventUrl);
				// eslint-disable-next-line no-console
				console.log('Description:', event.node.description);
				// eslint-disable-next-line no-console
				console.log('Time:', new Date(event.node.dateTime));
				// eslint-disable-next-line no-console
				console.log('Duration:', event.node.duration);
				// eslint-disable-next-line no-console
				console.log(
					'Location:',
					event.node.venue.name +
						'\n' +
						event.node.venue.address +
						'\n' +
						event.node.venue.city +
						', ' +
						event.node.venue.state +
						'\n' +
						event.node.venue.postalCode,
				);
				// eslint-disable-next-line no-console
				console.log('Group:', event.node.group.name);
				// eslint-disable-next-line no-console
				console.log('--------------');
			});
		}
	}

	await fetchAndPrintAllFutureEvents();
}

export async function handler() {
	const token = await getMeetupToken();
	await fetchEvents(token);
}
