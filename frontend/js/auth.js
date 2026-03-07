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
      } else {
        ui.showAuthMessage(data.error || "Login failed", "error");
      }
    } catch (error) {
      ui.showAuthMessage("Network error. Please try again.", "error");
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
      } else {
        ui.showAuthMessage(data.error || "Registration failed", "error");
      }
    } catch (error) {
      ui.showAuthMessage("Network error. Please try again.", "error");
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
