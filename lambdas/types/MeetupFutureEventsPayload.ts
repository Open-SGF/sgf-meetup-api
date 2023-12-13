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
