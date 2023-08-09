const meetupAccessToken = 'b017608bc6697a1ede777c2fb44d78a8';
const meetupGraphQlEndpoint = 'https://api.meetup.com/gql';
const batchSize = 10; // Number of events to fetch in each batch

const GET_FUTURE_EVENTS = `
query ($itemsNum: Int!, $cursor: String) {
    group1Events: groupByUrlname(urlname: "open-sgf") {
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
    
    group2Events: groupByUrlname(urlname: "springfield-women-in-tech") {
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
    
    group3Events: groupByUrlname(urlname: "sgfdevs") {
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

async function fetchAllFutureEvents(cursor = null) {
  const requestBody = JSON.stringify({
    query: GET_FUTURE_EVENTS,
    variables: {
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
    console.log(data.data.group1Events.unifiedEvents.edges);
    const events = data.data.groupByUrlname.unifiedEvents.edges;

    if (data.data.groupByUrlname.unifiedEvents.pageInfo.hasNextPage) {
      const nextCursor = data.data.groupByUrlname.unifiedEvents.pageInfo.endCursor;
      const nextEvents = await fetchAllFutureEvents(nextCursor);
      events.push(...nextEvents);
    }

    return events;
    
  } catch (error) {
      console.error('Error fetching future events:', error);
      return [];
  }
}

async function fetchAndPrintAllFutureEvents() {
  const futureEvents = await fetchAllFutureEvents();

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

fetchAndPrintAllFutureEvents();
  