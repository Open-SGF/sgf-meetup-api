import { sign } from 'jsonwebtoken';
import fetch from 'node-fetch';
import { isRecord } from './lib/isRecord';
import type { Handler } from 'aws-lambda';

interface MeetupSecret {
	privateKey: string;
	userId: string;
	clientKey: string;
	signingKeyId: string;
}

const MEETUP_AUTH_URL = 'https://secure.meetup.com/oauth2/access';

const MEETUP_PRIVATE_KEY = process.env.MEETUP_PRIVATE_KEY;
const MEETUP_USER_ID = process.env.MEETUP_USER_ID;
const MEETUP_CLIENT_KEY = process.env.MEETUP_CLIENT_KEY;
const MEETUP_SIGNING_KEY_ID = process.env.MEETUP_SIGNING_KEY_ID;

export async function getMeetupToken(): Promise<string> {
	const secret = parseSecret();

	const signedJWT = sign({}, secret.privateKey, {
		algorithm: 'RS256',
		issuer: secret.clientKey,
		subject: secret.userId,
		audience: 'api.meetup.com',
		keyid: secret.signingKeyId,
		expiresIn: 120,
	});

	const requestBody = new URLSearchParams({
		grant_type: 'urn:ietf:params:oauth:grant-type:jwt-bearer',
		assertion: signedJWT,
	});

	const res = await fetch(MEETUP_AUTH_URL, {
		method: 'POST',
		headers: {
			'Content-Type': 'application/x-www-form-urlencoded',
		},
		body: requestBody,
	}).then((res) => res.json());

	if (
		!isRecord(res) ||
		!('access_token' in res) ||
		typeof res.access_token !== 'string'
	) {
		throw new Error('Unexpected response from meetup');
	}

	return res.access_token;
}

function parseSecret(): MeetupSecret {
	if (
		typeof MEETUP_PRIVATE_KEY !== 'string' ||
		typeof MEETUP_USER_ID !== 'string' ||
		typeof MEETUP_CLIENT_KEY !== 'string' ||
		typeof MEETUP_SIGNING_KEY_ID !== 'string'
	) {
		throw new Error('Missing or invalid keys in AWS secret');
	}

	return {
		privateKey: MEETUP_PRIVATE_KEY,
		userId: MEETUP_USER_ID,
		clientKey: MEETUP_CLIENT_KEY,
		signingKeyId: MEETUP_SIGNING_KEY_ID,
	};
}

export interface MeetupTokenRequestEvent {
	clientId: string;
}

export const handler: Handler = async (event: MeetupTokenRequestEvent) => {
	try {
		const { clientId } = event;
		console.info(`Getting Meetup.com API key for client ${clientId}`); // eslint-disable-line no-console

		const token = await getMeetupToken();
		return { token };
		// eslint-disable-next-line @typescript-eslint/no-explicit-any
	} catch (error: any) {
		if (error instanceof Error) {
			return {
				errorName: error.name,
				errorMessage: error.message,
			};
		}

		return {
			errorName: error?.name,
			errorMessage: String(error),
		};
	}
};
