/**
 * Application Configuration
 * Contains all configuration constants and API endpoint definitions
 */

// API Base URL - Uses current origin to work across environments
const API_BASE_URL = window.location.origin;

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
  CODE: "NGN",
  SYMBOL: "₦",
  LOCALE: "en-NG",
};

// Local storage keys
const STORAGE_KEYS = {
  TOKEN: "token",
  EMAIL: "email",
};
