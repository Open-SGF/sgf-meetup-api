import { AttributeValue } from '@aws-sdk/client-dynamodb';

/**
 * Container for all the stuff associated with a Meetup event
 */
export interface MeetupEvent {
	id: string;
	title: string;
	eventUrl: string;
	description: string;
	dateTime: string;
	duration: string;
	venue: Venue;
	group: Group;
	host: Host;
	images: Image[];
}

export function meetupEventFromDynamoDbItem(
	item: Record<string, AttributeValue>,
): MeetupEvent {
	const group = {
		name: item.MeetupGroupName.S!,
		urlname: item.MeetupGroupUrlName.S!,
	} satisfies Group;

	const id = item.Id.S!;
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
		id,
		group,
		title,
		eventUrl,
		description,
		dateTime,
		duration,
		venue,
		host: {
			name: item.HostName.S!,
		},
		images: [], // TODO
	};

	return meetupEvent;
}

export function meetupEventToDynamoDbItem(
	meetupEvent: MeetupEvent,
): Record<string, AttributeValue> {

	const item: Record<string, AttributeValue> = {
		Id: { S: meetupEvent.id },
		MeetupGroupName: { S: meetupEvent.group.name },
		MeetupGroupUrlName: { S: meetupEvent.group.urlname },
		Title: { S: meetupEvent.title },
		EventUrl: { S: meetupEvent.eventUrl },
		Description: { S: meetupEvent.description },
		EventDateTime: { S: meetupEvent.dateTime },
		Duration: { S: meetupEvent.duration },
		VenueName: { S: meetupEvent.venue.name },
		VenueAddress: { S: meetupEvent.venue.address },
		VenueCity: { S: meetupEvent.venue.city },
		VenueState: { S: meetupEvent.venue.state },
		VenuePostalCode: { S: meetupEvent.venue.postalCode },
		HostName: { S: meetupEvent.host.name },
	};

	return item;
}

export interface MeetupFutureEventsPayload {
	data: {
		events: Events | null;
	};
}

export interface Events {
	unifiedEvents: UnifiedEvents;
}

export interface UnifiedEvents {
	count: number;
	pageInfo: PageInfo;
	edges: Edge[];
}

export interface Edge {
	node: MeetupEvent;
}

export interface Group {
	name: string;
	urlname: string;
}

export interface Host {
	name: string;
}

export interface Image {
	baseUrl: string;
	preview: null;
}

export interface Venue {
	name: string;
	address: string;
	city: string;
	state: string;
	postalCode: string;
}

export interface PageInfo {
	endCursor: string;
	hasNextPage: boolean;
}
