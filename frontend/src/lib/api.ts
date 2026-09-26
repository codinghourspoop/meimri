export interface ApiErrorPayload {
  error?: string;
  [key: string]: unknown;
}

export class ApiError extends Error {
  status: number;
  payload: ApiErrorPayload | null;

  constructor(status: number, payload: ApiErrorPayload | null) {
    super(payload?.error || `request failed (${status})`);
    this.status = status;
    this.payload = payload;
  }
}

interface RequestOptions {
  method?: string;
  body?: unknown;
}

async function request<T = any>(path: string, { method = "GET", body }: RequestOptions = {}): Promise<T> {
  const res = await fetch(path, {
    method,
    credentials: "same-origin",
    headers: body !== undefined ? { "Content-Type": "application/json" } : undefined,
    body: body !== undefined ? JSON.stringify(body) : undefined,
  });
  const text = await res.text();
  let data: any = null;
  try {
    data = text ? JSON.parse(text) : null;
  } catch {
    data = null;
  }
  if (!res.ok) throw new ApiError(res.status, data);
  return data as T;
}

export interface User {
  username: string;
  display_name: string;
}

export interface DeckSummary {
  id: string;
  title: string;
  description: string;
  card_count: number;
}

export interface DeckCard {
  front: string;
  back: string;
}

export interface Deck extends DeckSummary {
  cards: DeckCard[];
}

export const api = {
  get: <T = any>(path: string) => request<T>(path),
  post: <T = any>(path: string, body: unknown) => request<T>(path, { method: "POST", body: body ?? {} }),
  me: () => request<{ user: User; unread: number }>("/api/me"),
  login: (username: string, password: string) =>
    request("/api/auth/login", { method: "POST", body: { username, password } }),
  register: (username: string, display_name: string, password: string) =>
    request("/api/auth/register", { method: "POST", body: { username, display_name, password } }),
  recover: (username: string, code: string, password: string) =>
    request("/api/auth/recover", { method: "POST", body: { username, code, password } }),
  logout: () => request("/api/auth/logout", { method: "POST", body: {} }),
  searchDecks: (q: string) => request<DeckSummary[]>(`/api/decks/search?q=${encodeURIComponent(q ?? "")}`),
  createDeck: (title: string, description: string, cards: DeckCard[]) =>
    request("/api/decks", { method: "POST", body: { title, description, cards } }),
  getDeck: (id: string) => request<Deck>(`/api/decks/${encodeURIComponent(id)}`),
  getLearn: (deckId: string, passes: number) =>
    request(`/api/learn?deck_id=${encodeURIComponent(deckId)}&passes=${encodeURIComponent(passes ?? 3)}`),
  notifications: () => request("/api/notifications"),
  markNotificationsRead: () => request("/api/notifications/read", { method: "POST", body: {} }),
};
