export const API_URL = "http://localhost:7878";

export async function fetchAPI(method: string, url: string, body?: any) {
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
