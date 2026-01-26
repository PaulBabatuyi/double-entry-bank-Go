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

    let response;
    try {
      response = await fetch(`${API_BASE_URL}${endpoint}`, {
        cache: "no-store",
        ...options,
        headers,
      });
    } catch (error) {
      const reason = error && error.message ? error.message : "request failed";
      throw new Error(
        `Request to ${API_BASE_URL}${endpoint} failed: ${reason}`,
      );
    }

    // Handle 401: if the user already has a session token, it has expired — log
    // them out. During login/register there is no token, so let the response
    // fall through to the caller so it can show the proper error message.
    if (response.status === 401 && state.getToken()) {
      auth.logout();
      throw new Error("Session expired - please login again");
    }

    let data = null;
    const contentType = response.headers.get("Content-Type") || "";
    if (contentType.includes("application/json")) {
      data = await response.json();
    } else {
      const text = await response.text();
      try {
        data = JSON.parse(text);
      } catch {
        data = { message: text };
      }
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
        from_id: fromAccountId,
        to_id: toAccountId,
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

  /**
   * Reconcile an account balance against ledger entries
   */
  async reconcileAccount(accountId) {
    return this.request(API_ENDPOINTS.RECONCILE(accountId));
  },

  /**
   * Get a full transaction view by transaction ID
   */
  async getTransaction(txId) {
    return this.request(API_ENDPOINTS.TRANSACTIONS(txId));
  },
};
