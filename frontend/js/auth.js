/**
 * Authentication Module
 * Handles user authentication (login, register, logout)
 */

const auth = {
  /**
   * Check if user is authenticated
   */
  checkAuth() {
    if (state.initFromStorage()) {
      ui.showDashboard();
      dashboard.loadData();
    } else {
      ui.showAuthScreen();
    }
  },

  /**
   * Handle user login
   */
  async handleLogin(event) {
    event.preventDefault();

    const email = document.getElementById("login-email").value;
    const password = document.getElementById("login-password").value;

    try {
      const { response, data } = await api.login(email, password);

      if (response.ok) {
        state.setToken(data.token);
        state.setEmail(email);

        ui.showToast("Welcome back!", "Login successful", "success");
        ui.showDashboard();
        dashboard.loadData();
      } else if (response.status === 401) {
        ui.showAuthMessage(
          "Invalid email or password. Please try again.",
          "error",
        );
      } else {
        ui.showAuthMessage(
          data.error || "Login failed. Please try again.",
          "error",
        );
      }
    } catch (error) {
      ui.showAuthMessage(
        "Network error. Check your connection and try again.",
        "error",
      );
    }
  },

  /**
   * Handle user registration
   */
  async handleRegister(event) {
    event.preventDefault();

    const email = document.getElementById("register-email").value;
    const password = document.getElementById("register-password").value;

    try {
      const { response, data } = await api.register(email, password);

      if (response.ok) {
        state.setToken(data.token);
        state.setEmail(data.email);

        ui.showToast(
          "Account created!",
          "Welcome to the banking system",
          "success",
        );
        ui.showDashboard();
        dashboard.loadData();
      } else if (response.status === 409) {
        ui.showAuthMessage(
          "An account with this email already exists.",
          "error",
        );
      } else {
        ui.showAuthMessage(
          data.error || "Registration failed. Please try again.",
          "error",
        );
      }
    } catch (error) {
      ui.showAuthMessage(
        "Network error. Check your connection and try again.",
        "error",
      );
    }
  },

  /**
   * Logout user
   */
  logout() {
    state.clear();
    ui.showAuthScreen();
    ui.showToast("Logged out", "Come back soon!", "info");
  },
};
