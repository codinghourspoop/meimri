import { writable } from "svelte/store";
import { api } from "./api.js";

export const session = writable({ status: "loading", user: null, unread: 0 });
export const pendingRecoveryCodes = writable(null);

export async function loadMe() {
  try {
    const data = await api.me();
    session.set({ status: "authed", user: data.user, unread: data.unread ?? 0 });
  } catch (e) {
    if (e?.status === 401) session.set({ status: "guest", user: null, unread: 0 });
    else session.set({ status: "guest", user: null, unread: 0 });
  }
}

export async function logout() {
  try {
    await api.logout();
  } finally {
    session.set({ status: "guest", user: null, unread: 0 });
  }
}
