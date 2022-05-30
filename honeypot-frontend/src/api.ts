export let API_URL = "http://localhost:7878";

if (window.location.hostname !== "localhost") {
  API_URL = window.location.protocol + "//" + window.location.hostname;
  if (window.location.port) {
    API_URL += ":" + window.location.port;
  }
}

// parse the location query
const query = new URLSearchParams(window.location.search);
if(query.has("api")) {
  API_URL = query.get("api");
}

export async function fetchAPI(
  method: string,
  url: string,
  body?: any,
  query?: any
) {
  if (query) {
    url += "?";
    for (const key in query) {
      if (Array.isArray(query[key])) {
        for (const value of query[key]) {
          url += `${encodeURIComponent(key)}=${encodeURIComponent(value)}&`;
        }
      } else {
        url += `${encodeURIComponent(key)}=${encodeURIComponent(query[key])}&`;
      }
    }
  }
  const response = await fetch(`${API_URL}${url}`, {
    method,
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(body),
  });
  if (!response.ok) {
    throw new Error(response.statusText + ": " + (await response.text()));
  }
  return response.json();
}

export function WebSocketAPI(endpoint: string) {
  let websocketBaseUrl = API_URL.replace(/^http/, "ws");
  const ws = new WebSocket(`${websocketBaseUrl}${endpoint}`);
  return ws;
}
