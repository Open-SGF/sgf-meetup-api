import 'dotenv/config';
import * as jwt from 'jsonwebtoken';
import fetch from 'node-fetch';
import { promises as fs } from 'fs';
import { PutItemCommand, DynamoDBClient } from '@aws-sdk/client-dynamodb';

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

// await Promise.all([fetchEvents, fetchGroups]);

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

            const client = new DynamoDBClient();
            const futureEvents = await fetchAllFutureEvents(urlname);

            const promises = futureEvents.map(async (event) => {
              const input = {
                title: event.node.title,
                eventUrl: event.node.eventUrl,
                description: event.node.description,
                time: new Date(event.node.dateTime),
                duration: event.node.duration,
                venue: {
                  name: event.node.venue.name,
                  address: event.node.venue.address,
                  city: event.node.venue.city,
                  state: event.node.venue.state,
                  postalCode: event.node.venue.postalCode
                },
                group: {
                  name: event.node.group.name
                }
              };
              const command = new PutItemCommand(input);
              await client.send(command);
            })
            
            await Promise.all(promises);
        }
    }

    await fetchAndPrintAllFutureEvents();
}

export async function handler() {
    const token = await getMeetupAccessToken();
    await fetchEvents(token);
}