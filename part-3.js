
//    Prototype
//    Start with array of 2 group IDs and get events (print to console)
//    Use Auth tokens to make Auth requests (blank strings for now)
//    Packages for GraphQL API call?

//    1. Loop over IDs
//    2. API call
//    3. Print

//    See https://www.meetup.com/api/guide/#graphQl-guide for documentation


const meetupAccessToken = 'YOUR_MEETUP_ACCESS_TOKEN';
const meetupGraphQlEndpoint = 'https://api.meetup.com/gql';
const batchSize = 10; // Number of events to fetch in each batch

const GET_FUTURE_EVENTS = `
query ($urlname: String!, $itemsNum: Int!, $cursor: String) {
    proNetworkByUrlname(urlname: $urlname) {
      eventsSearch(filter: { status: UPCOMING }, input: { first: $itemsNum, after: $cursor }) {
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
              id
              name
              address
              city
              state
              postalCode
            }
            group {
              id
              name
              urlname
            }
            host {
              id
              name
            }
            images {
              id
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
    const events = data.data.proNetworkByUrlname.eventsSearch.edges;

    if (data.data.proNetworkByUrlname.eventsSearch.pageInfo.hasNextPage) {
      const nextCursor = data.data.proNetworkByUrlname.eventsSearch.pageInfo.endCursor;
      const nextEvents = await fetchAllFutureEvents(urlname, nextCursor);
      events.push(...nextEvents);
    }

    return events;
  } catch (error) {
    console.error('Error fetching future events:', error);
    return [];
  }
}

async function fetchAndPrintAllFutureEvents(urlname) {
  const futureEvents = await fetchAllFutureEvents(urlname);

  futureEvents.forEach(event => {
    console.log('--------------');
    console.log('Event:', event.node.title);
    console.log('Event Link:', event.node.eventUrl);
    console.log('Description:', event.node.description);
    console.log('Time:', new Date(event.node.dateTime));
    console.log('Duration:', event.node.duration);
    console.log('Location:', event.node.venue.name + '\n' + event.node.venue.address + '\n' + event.node.venue.city + ', ' + event.node.venue.state + '\n' + event.node.venue.postalCode);
    console.log('Group:', event.node.group.name);
    console.log('--------------');
  });
}

const proNetworkUrlname = 'YOUR_PRO_NETWORK_URLNAME';
fetchAndPrintAllFutureEvents(proNetworkUrlname);
