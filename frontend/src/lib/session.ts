import { writable } from "svelte/store";
import { api, type User } from "./api.js";

export type SessionStatus = "loading" | "authed" | "guest";

export interface SessionState {
  status: SessionStatus;
  user: User | null;
  unread: number;
}

export const session = writable<SessionState>({ status: "loading", user: null, unread: 0 });
export const pendingRecoveryCodes = writable<string[] | null>(null);

export async function loadMe(): Promise<void> {
  try {
    const data = await api.me();
    session.set({ status: "authed", user: data.user, unread: data.unread ?? 0 });
  } catch {
    session.set({ status: "guest", user: null, unread: 0 });
  }
}

export async function logout(): Promise<void> {
  try {
    await api.logout();
  } finally {
    session.set({ status: "guest", user: null, unread: 0 });
  }
}
