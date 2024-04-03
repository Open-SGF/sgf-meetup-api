import {
	SecretsManagerClient,
	GetSecretValueCommand,
} from '@aws-sdk/client-secrets-manager';
import { atob } from 'buffer';
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

const AWS_MEETUP_SECRET_NAME = 'prod/sgf-meetup-api/meetup';
const MEETUP_AUTH_URL = 'https://secure.meetup.com/oauth2/access';

export async function getMeetupToken(): Promise<string> {
	const client = new SecretsManagerClient({
		region: 'us-east-2',
	});

	const response = await client.send(
		new GetSecretValueCommand({
			SecretId: AWS_MEETUP_SECRET_NAME,
		}),
	);

	const secret = parseSecret(response.SecretString);

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

function parseSecret(secretString: string | undefined): MeetupSecret {
	if (!secretString) {
		throw new Error('Invalid secret json from AWS');
	}

	const secret = JSON.parse(secretString) as unknown;

	if (!isRecord(secret)) {
		throw new Error('Invalid secret json from AWS');
	}

	const {
		meetupPrivateKeyBase64,
		meetupUserId,
		meetupClientKey,
		meetupSigningKeyId,
	} = secret;

	if (
		typeof meetupPrivateKeyBase64 !== 'string' ||
		typeof meetupUserId !== 'string' ||
		typeof meetupClientKey !== 'string' ||
		typeof meetupSigningKeyId !== 'string'
	) {
		throw new Error('Missing or invalid keys in AWS secret');
	}

	const privateKey = atob(meetupPrivateKeyBase64);

	return {
		privateKey,
		userId: meetupUserId,
		clientKey: meetupClientKey,
		signingKeyId: meetupSigningKeyId,
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
