import 'dotenv/config';
import jwt from 'jsonwebtoken';
import fetch from 'node-fetch';
import { promises as fs } from 'fs';

async function getMeetupAccessToken() {
    const privateKey = await fs.readFile('./meetup-private-key');
    const url = 'https://secure.meetup.com/oauth2/access';

    const signedJWT = jwt.sign(
        {},
        privateKey,
        {
            algorithm: 'RS256',
            issuer: process.env.MEETUP_CLIENT_KEY,
            subject: process.env.MEETUP_USER_ID,
            audience: 'api.meetup.com',
            keyid: process.env.MEETUP_SIGNING_KEY_ID,
            expiresIn: 120
        }
    );

    const requestBody = new URLSearchParams({
        grant_type: 'urn:ietf:params:oauth:grant-type:jwt-bearer',
        assertion: signedJWT
    });


    const res = await fetch(url, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/x-www-form-urlencoded'
        },
        body: requestBody
    }).then(res => res.json());

    return res.access_token
}

async function fetchEvents(meetupAccessToken) {
    const meetupGraphQlEndpoint = 'https://api.meetup.com/gql';
    const batchSize = 10; // Number of events to fetch in each batch

    const GROUP_URLNAMES = ['open-sgf', 'springfield-women-in-tech', 'sgfdevs'];

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

    async function fetchAllFutureEvents(urlname, cursor = null) {
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
            const data = await response.json();
            const events = data.data.events.unifiedEvents.edges;

            if (data.data.events.unifiedEvents.pageInfo.hasNextPage) {
                const nextCursor = data.data.events.unifiedEvents.pageInfo.endCursor;
                const nextEvents = await fetchAllFutureEvents(urlname, nextCursor);
                events.push(...nextEvents);
            }

            return events;
        } catch (error) {
            console.error('Error fetching future events:', error);
            return [];
        }
    }

    async function fetchAndPrintAllFutureEvents() {
        for (const urlname of GROUP_URLNAMES) {
            const futureEvents = await fetchAllFutureEvents(urlname);

            futureEvents.forEach(event => {
                console.log('--------------');
                console.log('Event:', event.node.title);
                console.log('Event Link:', event.node.eventUrl);
                console.log('Description:', event.node.description);
                console.log('Time:', new Date(event.node.dateTime));
                console.log('Duration:', event.node.duration);
                console.log('Location:', event.node.venue.name + '\n'
                    + event.node.venue.address + '\n'
                    + event.node.venue.city + ', '
                    + event.node.venue.state + '\n'
                    + event.node.venue.postalCode);
                console.log('Group:', event.node.group.name);
                console.log('--------------');
            });
        }
    }

    await fetchAndPrintAllFutureEvents();
}

export async function handler() {
    const token = await getMeetupAccessToken();
    await fetchEvents(token);
}