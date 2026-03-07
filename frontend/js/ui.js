/**
 * UI Helper Functions
 * Manages UI state transitions and user feedback
 */

const ui = {
  /**
   * Show the authentication screen
   */
  showAuthScreen() {
    document.getElementById("auth-screen").classList.remove("hidden");
    document.getElementById("dashboard-screen").classList.add("hidden");
    document.getElementById("user-info").classList.add("hidden");
    document.getElementById("user-info").classList.remove("flex");
  },

  /**
   * Show the dashboard
   */
  showDashboard() {
    document.getElementById("auth-screen").classList.add("hidden");
    document.getElementById("dashboard-screen").classList.remove("hidden");
    document.getElementById("user-info").classList.remove("hidden");
    document.getElementById("user-info").classList.add("flex");
    document.getElementById("user-email").textContent = state.getEmail();
  },

  /**
   * Switch between login and register tabs
   */
  switchAuthTab(tab) {
    const loginTab = document.getElementById("login-tab");
    const registerTab = document.getElementById("register-tab");
    const loginForm = document.getElementById("login-form");
    const registerForm = document.getElementById("register-form");

    if (tab === "login") {
      loginTab.className =
        "flex-1 py-2 rounded-lg font-semibold transition auth-tab-active";
      registerTab.className =
        "flex-1 py-2 rounded-lg font-semibold transition auth-tab-inactive";
      loginForm.classList.remove("hidden");
      registerForm.classList.add("hidden");
    } else {
      registerTab.className =
        "flex-1 py-2 rounded-lg font-semibold transition auth-tab-active";
      loginTab.className =
        "flex-1 py-2 rounded-lg font-semibold transition auth-tab-inactive";
      registerForm.classList.remove("hidden");
      loginForm.classList.add("hidden");
    }

    this.hideAuthMessage();
  },

  /**
   * Show authentication error message
   */
  showAuthMessage(message, type) {
    const messageEl = document.getElementById("auth-message");
    messageEl.textContent = message;
    messageEl.className = `mt-4 p-3 rounded-lg text-sm text-center message-${type}`;
    messageEl.classList.remove("hidden");
  },

  /**
   * Hide authentication message
   */
  hideAuthMessage() {
    document.getElementById("auth-message").classList.add("hidden");
  },

  /**
   * Show toast notification
   */
  showToast(title, message, type) {
    const toast = document.getElementById("toast");
    const icon = document.getElementById("toast-icon");
    const titleEl = document.getElementById("toast-title");
    const messageEl = document.getElementById("toast-message");

    const icons = {
      success: { class: "fa-check-circle", color: "text-green-400" },
      error: { class: "fa-exclamation-circle", color: "text-red-400" },
      info: { class: "fa-info-circle", color: "text-blue-400" },
    };

    const iconConfig = icons[type] || icons.info;
    icon.className = `fas ${iconConfig.class} text-2xl ${iconConfig.color}`;
    titleEl.textContent = title;
    messageEl.textContent = message;

    toast.classList.remove("hidden");

    setTimeout(() => {
      toast.classList.add("hidden");
    }, TOAST_DURATION);
  },

  /**
   * Show create account modal
   */
  showCreateAccountModal() {
    document.getElementById("modal-backdrop").classList.remove("hidden");
    document.getElementById("new-account-name").value = "";
  },

  /**
   * Close modal
   */
  closeModal() {
    document.getElementById("modal-backdrop").classList.add("hidden");
  },
};
