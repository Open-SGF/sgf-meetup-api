// QUERY 1 - SINGLE GROUP

query ($urlname: String!, $cursor: String) {
    groupByUrlname(urlname: $urlname) {
      unifiedEvents(input: { first: 3, after: $cursor }) {
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

  { "urlname": ["open-sgf", "springfield-women-in-technology", "sgfdevs"] }


  // QUERY 2 - MULTIPLE GROUPS

  query ($cursor: String) {
    group1Events: groupByUrlname(urlname: "open-sgf") {
      unifiedEvents(input: { first: 3, after: $cursor }) {
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
      unifiedEvents(input: { first: 3, after: $cursor }) {
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
      unifiedEvents(input: { first: 3, after: $cursor }) {
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