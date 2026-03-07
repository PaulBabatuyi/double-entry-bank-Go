/**
 * Dashboard Module
 * Manages dashboard data loading and rendering
 */

const dashboard = {
  /**
   * Load all dashboard data
   */
  async loadData() {
    await this.loadAccounts();
    this.updateStats();
  },

  /**
   * Load user accounts
   */
  async loadAccounts() {
    try {
      const { response, data } = await api.getAccounts();

      if (response.ok) {
        state.setAccounts(data);
        this.renderAccounts();
        this.populateAccountSelects();
        await this.loadLatestTransactions();
      }
    } catch (error) {
      console.error("Error loading accounts:", error);
    }
  },

  /**
   * Render accounts list
   */
  renderAccounts() {
    const accountsList = document.getElementById("accounts-list");
    const accounts = state.getAccounts();

    if (accounts.length === 0) {
      accountsList.innerHTML = `
                <div class="text-center py-8 text-gray-400">
                    <i class="fas fa-inbox text-4xl mb-3"></i>
                    <p>No accounts yet. Create one to get started!</p>
                </div>
            `;
      return;
    }

    accountsList.innerHTML = accounts
      .map(
        (account) => `
            <div class="account-card glass rounded-xl p-5 flex justify-between items-center cursor-pointer" 
                 onclick="viewAccountDetails('${account.id}')">
                <div>
                    <div class="flex items-center space-x-2 mb-2">
                        <i class="fas fa-wallet text-purple-400"></i>
                        <h3 class="font-bold text-lg">${account.name}</h3>
                    </div>
                    <p class="text-sm text-gray-400">ID: ${utils.truncate(account.id)}</p>
                </div>
                <div class="text-right">
                    <p class="text-2xl font-bold text-green-400">${utils.formatCurrency(account.balance)}</p>
                    <p class="text-xs text-gray-400">${account.currency}</p>
                </div>
            </div>
        `,
      )
      .join("");
  },

  /**
   * Populate account select dropdowns
   */
  populateAccountSelects() {
    const selects = [
      document.getElementById("deposit-account"),
      document.getElementById("withdraw-account"),
      document.getElementById("transfer-from"),
      document.getElementById("transfer-to"),
    ];

    const accounts = state.getAccounts();

    selects.forEach((select) => {
      const currentValue = select.value;
      select.innerHTML =
        '<option value="">Select Account</option>' +
        accounts
          .map(
            (acc) =>
              `<option value="${acc.id}">${acc.name} (${utils.formatCurrency(acc.balance)})</option>`,
          )
          .join("");

      if (currentValue) {
        select.value = currentValue;
      }
    });
  },

  /**
   * Load and display transaction history
   */
  async loadLatestTransactions() {
    const transactionsList = document.getElementById("transactions-list");
    const accounts = state.getAccounts();

    if (accounts.length === 0) {
      transactionsList.innerHTML = `
                <div class="text-center py-8 text-gray-400">
                    <i class="fas fa-file-invoice text-4xl mb-3"></i>
                    <p>No transactions yet</p>
                </div>
            `;
      return;
    }

    try {
      const accountId = accounts[0].id;
      const { response, data: entries } = await api.getEntries(accountId);

      if (response.ok && entries && entries.length > 0) {
        transactionsList.innerHTML = entries
          .slice(0, 10)
          .map(
            (entry) => `
                    <div class="transaction-item glass rounded-lg p-4 flex justify-between items-center">
                        <div class="flex items-center space-x-3">
                            <div class="w-10 h-10 rounded-full ${entry.debit !== "0" ? "bg-red-500/20" : "bg-green-500/20"} 
                                 flex items-center justify-center">
                                <i class="fas ${entry.debit !== "0" ? "fa-arrow-up text-red-400" : "fa-arrow-down text-green-400"}"></i>
                            </div>
                            <div>
                                <p class="font-semibold">${entry.operation_type || "Transaction"}</p>
                                <p class="text-xs text-gray-400">${utils.formatDate(entry.created_at)}</p>
                            </div>
                        </div>
                        <div class="text-right">
                            <p class="font-bold ${entry.debit !== "0" ? "text-red-400" : "text-green-400"}">
                                ${entry.debit !== "0" ? "-" : "+"}${utils.formatCurrency(entry.debit !== "0" ? entry.debit : entry.credit)}
                            </p>
                            <p class="text-xs text-gray-400">${utils.truncate(entry.transaction_id)}</p>
                        </div>
                    </div>
                `,
          )
          .join("");
      } else {
        transactionsList.innerHTML = `
                    <div class="text-center py-8 text-gray-400">
                        <i class="fas fa-file-invoice text-4xl mb-3"></i>
                        <p>No transactions yet</p>
                    </div>
                `;
      }
    } catch (error) {
      console.error("Error loading transactions:", error);
    }
  },

  /**
   * Update dashboard statistics
   */
  updateStats() {
    const accounts = state.getAccounts();
    const totalAccounts = accounts.length;
    const totalBalance = accounts.reduce(
      (sum, acc) => sum + parseFloat(acc.balance || 0),
      0,
    );

    document.getElementById("stat-accounts").textContent = totalAccounts;
    document.getElementById("stat-balance").textContent =
      utils.formatCurrency(totalBalance);
  },

  /**
   * View account details
   */
  viewAccountDetails(accountId) {
    const account = state.getAccounts().find((acc) => acc.id === accountId);
    if (account) {
      ui.showToast(
        account.name,
        `Balance: ${utils.formatCurrency(account.balance)} ${account.currency}`,
        "info",
      );
    }
  },
};
