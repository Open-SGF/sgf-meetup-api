export class MeetupEvent {
  group: string;
  title: string;
  eventUrl: string;
  description: string;
  startTime: Date;
  duration: string;
  // venue: {
  //     name: string,
  //     address: string,
  //     city: string,
  //     state: string,
  //     postalCode: string
  // };

  constructor(
    group: string,
    title: string,
    eventUrl: string,
    description: string,
    startTime: Date,
    duration: string,
    // venue: {
    //     name: string,
    //     address: string,
    //     city: string,
    //     state: string,
    //     postalCode: string
    // },
  ) {
    this.group = group;
    this.title = title;
    this.eventUrl = eventUrl;
    this.description = description;
    this.startTime = startTime;
    this.duration = duration;
    // this.venue = venue;
  }

  static fromDynamoDBItem(item: any) {
    const group = item.MeetupGroup.S;
    const title = item.Title.S;
    const eventUrl = item.EventUrl.S;
    const description = item.Description.S;
    const startTime = new Date(item.StartTime.S);
    const duration = item.Duration.S;
    // const venue = item.venue.S;
    const meetupEvent = new MeetupEvent(group, title, eventUrl, description, startTime, duration);
    return meetupEvent;
  }
}
