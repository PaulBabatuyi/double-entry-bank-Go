/**
 * State Management
 * Manages the application's global state
 */

const state = {
  currentUser: {
    email: "",
    token: "",
    accounts: [],
  },

  // Getters
  getToken() {
    return this.currentUser.token;
  },

  getEmail() {
    return this.currentUser.email;
  },

  getAccounts() {
    return this.currentUser.accounts;
  },

  // Setters
  setToken(token) {
    this.currentUser.token = token;
    localStorage.setItem(STORAGE_KEYS.TOKEN, token);
  },

  setEmail(email) {
    this.currentUser.email = email;
    localStorage.setItem(STORAGE_KEYS.EMAIL, email);
  },

  setAccounts(accounts) {
    this.currentUser.accounts = accounts || [];
  },

  // Initialize from localStorage
  initFromStorage() {
    const token = localStorage.getItem(STORAGE_KEYS.TOKEN);
    const email = localStorage.getItem(STORAGE_KEYS.EMAIL);

    if (token && email) {
      this.currentUser.token = token;
      this.currentUser.email = email;
      return true;
    }
    return false;
  },

  // Clear all state
  clear() {
    this.currentUser = {
      email: "",
      token: "",
      accounts: [],
    };
    localStorage.removeItem(STORAGE_KEYS.TOKEN);
    localStorage.removeItem(STORAGE_KEYS.EMAIL);
  },
};
