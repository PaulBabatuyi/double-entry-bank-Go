/**
 * Main Application Entry Point
 * Initializes the application and sets up event listeners
 */

// Initialize app when DOM is ready
document.addEventListener("DOMContentLoaded", () => {
  auth.checkAuth();
});

// Make functions globally available for inline event handlers
window.switchAuthTab = ui.switchAuthTab.bind(ui);
window.handleLogin = auth.handleLogin.bind(auth);
window.handleRegister = auth.handleRegister.bind(auth);
window.logout = auth.logout.bind(auth);
window.showCreateAccountModal = ui.showCreateAccountModal.bind(ui);
window.closeModal = ui.closeModal.bind(ui);
window.handleDeposit = transactions.handleDeposit.bind(transactions);
window.handleWithdraw = transactions.handleWithdraw.bind(transactions);
window.handleTransfer = transactions.handleTransfer.bind(transactions);
window.handleCreateAccount =
  transactions.handleCreateAccount.bind(transactions);
window.viewAccountDetails = dashboard.viewAccountDetails.bind(dashboard);
