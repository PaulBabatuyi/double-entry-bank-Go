/**
 * Application Configuration
 * Contains all configuration constants and API endpoint definitions
 */

// API Base URL - override in index.html by setting window.API_BASE_URL before this file loads.
// Default to current origin so frontend and backend can be served together from the same host.
const API_BASE_URL =
  window.API_BASE_URL || window.VITE_API_BASE_URL || window.location.origin;

// API Endpoints
const API_ENDPOINTS = {
  REGISTER: "/register",
  LOGIN: "/login",
  ACCOUNTS: "/accounts",
  DEPOSIT: (accountId) => `/accounts/${accountId}/deposit`,
  WITHDRAW: (accountId) => `/accounts/${accountId}/withdraw`,
  TRANSFERS: "/transfers",
  ENTRIES: (accountId) => `/accounts/${accountId}/entries`,
  RECONCILE: (accountId) => `/accounts/${accountId}/reconcile`,
  TRANSACTIONS: (txId) => `/transactions/${txId}`,
  HEALTH: "/health",
};

// Toast notification duration (milliseconds)
const TOAST_DURATION = 4000;

// Currency settings
const CURRENCY = {
  CODE: "USD",
  SYMBOL: "$",
  LOCALE: "en-US",
};

// Local storage keys
const STORAGE_KEYS = {
  TOKEN: "token",
  EMAIL: "email",
};
