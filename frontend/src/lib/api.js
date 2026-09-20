export class ApiError extends Error {
  constructor(status, payload) {
    super(payload?.error || `request failed (${status})`);
    this.status = status;
    this.payload = payload;
  }
}

async function request(path, { method = "GET", body } = {}) {
  const res = await fetch(path, {
    method,
    credentials: "same-origin",
    headers: body !== undefined ? { "Content-Type": "application/json" } : undefined,
    body: body !== undefined ? JSON.stringify(body) : undefined,
  });
  const text = await res.text();
  let data = null;
  try {
    data = text ? JSON.parse(text) : null;
  } catch {
    data = null;
  }
  if (!res.ok) throw new ApiError(res.status, data);
  return data;
}

export const api = {
  get: (path) => request(path),
  post: (path, body) => request(path, { method: "POST", body: body ?? {} }),
  me: () => request("/api/me"),
  login: (username, password) => request("/api/auth/login", { method: "POST", body: { username, password } }),
  register: (username, display_name, password) =>
    request("/api/auth/register", { method: "POST", body: { username, display_name, password } }),
  recover: (username, code, password) =>
    request("/api/auth/recover", { method: "POST", body: { username, code, password } }),
  logout: () => request("/api/auth/logout", { method: "POST", body: {} }),
  searchDecks: (q) => request(`/api/decks/search?q=${encodeURIComponent(q ?? "")}`),
  createDeck: (title, description, cards) =>
    request("/api/decks", { method: "POST", body: { title, description, cards } }),
  getDeck: (id) => request(`/api/decks/${encodeURIComponent(id)}`),
  getLearn: (deckId, passes) =>
    request(`/api/learn?deck_id=${encodeURIComponent(deckId)}&passes=${encodeURIComponent(passes ?? 3)}`),
  notifications: () => request("/api/notifications"),
  markNotificationsRead: () => request("/api/notifications/read", { method: "POST", body: {} }),
};
