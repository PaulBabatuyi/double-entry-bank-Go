/**
 * Application Configuration
 * Contains all configuration constants and API endpoint definitions
 */

// API Base URL - Override with VITE_API_BASE_URL environment variable or use current origin
// For Vercel deployment, backend is hosted at https://double-entry-ledger-api.onrender.com, so we default to that for production
const API_BASE_URL =
  window.VITE_API_BASE_URL ||
  (window.location.hostname === "localhost"
    ? window.location.origin
    : "https://double-entry-ledger-api.onrender.com");

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
