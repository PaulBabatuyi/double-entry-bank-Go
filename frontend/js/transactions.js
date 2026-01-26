/**
 * Transactions Module
 * Handles all transaction operations (deposit, withdraw, transfer, create account)
 */

const transactions = {
  /**
   * Handle deposit
   */
  async handleDeposit(event) {
    event.preventDefault();

    const accountId = document.getElementById("deposit-account").value;
    const amount = parseFloat(document.getElementById("deposit-amount").value);

    if (!accountId || !amount) return;

    try {
      const { response, data } = await api.deposit(accountId, amount);

      if (response.ok) {
        ui.showToast(
          "Deposit successful!",
          `Added ${utils.formatCurrency(amount)}`,
          "success",
        );
        document.getElementById("deposit-amount").value = "";
        await dashboard.refreshAfterMutation();
      } else {
        ui.showToast(
          "Deposit failed",
          data.error || "Please try again",
          "error",
        );
      }
    } catch (error) {
      ui.showToast("Network error", "Please try again", "error");
    }
  },

  /**
   * Handle withdrawal
   */
  async handleWithdraw(event) {
    event.preventDefault();

    const accountId = document.getElementById("withdraw-account").value;
    const amount = parseFloat(document.getElementById("withdraw-amount").value);

    if (!accountId || !amount) return;

    try {
      const { response, data } = await api.withdraw(accountId, amount);

      if (response.ok) {
        ui.showToast(
          "Withdrawal successful!",
          `Withdrew ${utils.formatCurrency(amount)}`,
          "success",
        );
        document.getElementById("withdraw-amount").value = "";
        await dashboard.refreshAfterMutation();
      } else {
        ui.showToast(
          "Withdrawal failed",
          data.error || "Insufficient funds or invalid request",
          "error",
        );
      }
    } catch (error) {
      ui.showToast("Network error", "Please try again", "error");
    }
  },

  /**
   * Handle transfer between accounts
   */
  async handleTransfer(event) {
    event.preventDefault();

    const fromAccountId = document.getElementById("transfer-from").value;
    const toAccountId = document.getElementById("transfer-to").value;
    const amount = parseFloat(document.getElementById("transfer-amount").value);

    if (!fromAccountId || !toAccountId || !amount) return;

    if (fromAccountId === toAccountId) {
      ui.showToast(
        "Invalid transfer",
        "Cannot transfer to the same account",
        "error",
      );
      return;
    }

    try {
      const { response, data } = await api.transfer(
        fromAccountId,
        toAccountId,
        amount,
      );

      if (response.ok) {
        ui.showToast(
          "Transfer successful!",
          `Transferred ${utils.formatCurrency(amount)}`,
          "success",
        );
        document.getElementById("transfer-amount").value = "";
        await dashboard.refreshAfterMutation();
      } else {
        ui.showToast(
          "Transfer failed",
          data.error || "Please try again",
          "error",
        );
      }
    } catch (error) {
      ui.showToast("Network error", "Please try again", "error");
    }
  },

  /**
   * Handle account creation
   */
  async handleCreateAccount(event) {
    event.preventDefault();

    const name = document.getElementById("new-account-name").value;

    try {
      const { response, data } = await api.createAccount(name);

      if (response.ok) {
        ui.showToast("Account created!", `${name} is ready to use`, "success");
        ui.closeModal();
        await dashboard.refreshAfterMutation();
      } else {
        ui.showToast(
          "Creation failed",
          data.error || "Please try again",
          "error",
        );
      }
    } catch (error) {
      ui.showToast("Network error", "Please try again", "error");
    }
  },
};
