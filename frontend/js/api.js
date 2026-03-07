/**
 * API Service
 * Handles all HTTP requests to the backend API
 */

const api = {
  /**
   * Make an authenticated API request
   */
  async request(endpoint, options = {}) {
    const headers = {
      "Content-Type": "application/json",
      ...options.headers,
    };

    // Add authorization header if token exists
    const token = state.getToken();
    if (token) {
      headers["Authorization"] = `Bearer ${token}`;
    }

    const response = await fetch(`${API_BASE_URL}${endpoint}`, {
      ...options,
      headers,
    });

    const data = await response.json();

    // Handle 401 Unauthorized
    if (response.status === 401) {
      auth.logout();
      throw new Error("Unauthorized - please login again");
    }

    return { response, data };
  },

  /**
   * Register a new user
   */
  async register(email, password) {
    return this.request(API_ENDPOINTS.REGISTER, {
      method: "POST",
      body: JSON.stringify({ email, password }),
    });
  },

  /**
   * Login a user
   */
  async login(email, password) {
    return this.request(API_ENDPOINTS.LOGIN, {
      method: "POST",
      body: JSON.stringify({ email, password }),
    });
  },

  /**
   * Get all user accounts
   */
  async getAccounts() {
    return this.request(API_ENDPOINTS.ACCOUNTS);
  },

  /**
   * Create a new account
   */
  async createAccount(name) {
    return this.request(API_ENDPOINTS.ACCOUNTS, {
      method: "POST",
      body: JSON.stringify({ name }),
    });
  },

  /**
   * Deposit funds into an account
   */
  async deposit(accountId, amount) {
    return this.request(API_ENDPOINTS.DEPOSIT(accountId), {
      method: "POST",
      body: JSON.stringify({ amount }),
    });
  },

  /**
   * Withdraw funds from an account
   */
  async withdraw(accountId, amount) {
    return this.request(API_ENDPOINTS.WITHDRAW(accountId), {
      method: "POST",
      body: JSON.stringify({ amount }),
    });
  },

  /**
   * Transfer funds between accounts
   */
  async transfer(fromAccountId, toAccountId, amount) {
    return this.request(API_ENDPOINTS.TRANSFERS, {
      method: "POST",
      body: JSON.stringify({
        from_account_id: fromAccountId,
        to_account_id: toAccountId,
        amount,
      }),
    });
  },

  /**
   * Get account entries (transaction history)
   */
  async getEntries(accountId) {
    return this.request(API_ENDPOINTS.ENTRIES(accountId));
  },
};
