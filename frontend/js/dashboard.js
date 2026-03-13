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

    accountsList.innerHTML = "";
    const fragment = document.createDocumentFragment();

    accounts.forEach((account) => {
      const cardDiv = document.createElement("div");
      cardDiv.className =
        "account-card glass rounded-xl p-5 flex justify-between items-center cursor-pointer";
      cardDiv.addEventListener("click", () => viewAccountDetails(account.id));

      const leftDiv = document.createElement("div");

      const headerDiv = document.createElement("div");
      headerDiv.className = "flex items-center space-x-2 mb-2";

      const icon = document.createElement("i");
      icon.className = "fas fa-wallet text-purple-400";

      const nameH3 = document.createElement("h3");
      nameH3.className = "font-bold text-lg";
      nameH3.textContent = account.name;

      headerDiv.appendChild(icon);
      headerDiv.appendChild(nameH3);

      const idP = document.createElement("p");
      idP.className = "text-sm text-gray-400";
      idP.textContent = "ID: " + utils.truncate(account.id);

      const reconcileBtn = document.createElement("button");
      reconcileBtn.className =
        "mt-2 text-xs text-blue-300 hover:text-blue-200 transition";
      reconcileBtn.textContent = "Reconcile";
      reconcileBtn.addEventListener("click", async (event) => {
        event.stopPropagation();
        await this.reconcileAccount(account.id, account.name);
      });

      leftDiv.appendChild(headerDiv);
      leftDiv.appendChild(idP);
      leftDiv.appendChild(reconcileBtn);

      const rightDiv = document.createElement("div");
      rightDiv.className = "text-right";

      const balanceP = document.createElement("p");
      balanceP.className = "text-2xl font-bold text-green-400";
      balanceP.textContent = utils.formatCurrency(account.balance);

      const currencyP = document.createElement("p");
      currencyP.className = "text-xs text-gray-400";
      currencyP.textContent = account.currency;

      rightDiv.appendChild(balanceP);
      rightDiv.appendChild(currencyP);

      cardDiv.appendChild(leftDiv);
      cardDiv.appendChild(rightDiv);

      fragment.appendChild(cardDiv);
    });

    accountsList.appendChild(fragment);
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
    const statEl = document.getElementById("stat-transactions");

    if (accounts.length === 0) {
      transactionsList.innerHTML = "";
      const empty = document.createElement("div");
      empty.className = "text-center py-8 text-gray-400";
      const emptyIcon = document.createElement("i");
      emptyIcon.className = "fas fa-file-invoice text-4xl mb-3";
      const emptyText = document.createElement("p");
      emptyText.textContent = "No transactions yet";
      empty.appendChild(emptyIcon);
      empty.appendChild(emptyText);
      transactionsList.appendChild(empty);
      if (statEl) {
        statEl.textContent = "0";
      }
      return;
    }

    try {
      const settled = await Promise.allSettled(
        accounts.map((account) => api.getEntries(account.id)),
      );

      const entryResponses = settled
        .filter((result) => result.status === "fulfilled")
        .map((result) => result.value);

      const entries = entryResponses
        .filter(({ response }) => response.ok)
        .flatMap(({ data }) => (Array.isArray(data) ? data : []));

      if (entries.length > 0) {
        transactionsList.innerHTML = "";

        const fragment = document.createDocumentFragment();

        // Sort entries by created_at in descending order (newest first)
        const sortedEntries = [...entries].sort((a, b) => {
          return new Date(b.created_at) - new Date(a.created_at);
        });

        sortedEntries.slice(0, 10).forEach((entry) => {
          const isDebit = parseFloat(entry.debit) > 0;
          const operationText = entry.operation_type
            ? entry.operation_type.charAt(0).toUpperCase() +
              entry.operation_type.slice(1)
            : "Transaction";
          const dateText = utils.formatDate(entry.created_at);
          const amountValue = isDebit ? entry.debit : entry.credit;
          const amountText =
            (isDebit ? "-" : "+") + utils.formatCurrency(amountValue);
          const transactionIdText = utils.truncate(entry.transaction_id);

          const itemDiv = document.createElement("div");
          itemDiv.className =
            "transaction-item glass rounded-lg p-4 flex justify-between items-center";

          const leftDiv = document.createElement("div");
          leftDiv.className = "flex items-center space-x-3";

          const iconWrapper = document.createElement("div");
          iconWrapper.className =
            "w-10 h-10 rounded-full " +
            (isDebit ? "bg-red-500/20" : "bg-green-500/20") +
            " flex items-center justify-center";

          const icon = document.createElement("i");
          icon.className =
            "fas " +
            (isDebit
              ? "fa-arrow-up text-red-400"
              : "fa-arrow-down text-green-400");
          iconWrapper.appendChild(icon);

          const textWrapper = document.createElement("div");

          const operationP = document.createElement("p");
          operationP.className = "font-semibold";
          operationP.textContent = operationText;

          const dateP = document.createElement("p");
          dateP.className = "text-xs text-gray-400";
          dateP.textContent = dateText;

          textWrapper.appendChild(operationP);
          textWrapper.appendChild(dateP);

          leftDiv.appendChild(iconWrapper);
          leftDiv.appendChild(textWrapper);

          const rightDiv = document.createElement("div");
          rightDiv.className = "text-right";

          const amountP = document.createElement("p");
          amountP.className =
            "font-bold " + (isDebit ? "text-red-400" : "text-green-400");
          amountP.textContent = amountText;

          const idP = document.createElement("p");
          idP.className = "text-xs text-gray-400";
          idP.textContent = transactionIdText;

          rightDiv.appendChild(amountP);
          rightDiv.appendChild(idP);

          itemDiv.appendChild(leftDiv);
          itemDiv.appendChild(rightDiv);
          itemDiv.addEventListener("click", () =>
            this.viewTransactionDetails(entry.transaction_id),
          );

          fragment.appendChild(itemDiv);
        });

        transactionsList.appendChild(fragment);

        if (statEl) {
          statEl.textContent = String(entries.length);
        }
      } else {
        transactionsList.innerHTML = "";
        const empty = document.createElement("div");
        empty.className = "text-center py-8 text-gray-400";
        const emptyIcon = document.createElement("i");
        emptyIcon.className = "fas fa-file-invoice text-4xl mb-3";
        const emptyText = document.createElement("p");
        emptyText.textContent = "No transactions yet";
        empty.appendChild(emptyIcon);
        empty.appendChild(emptyText);
        transactionsList.appendChild(empty);
        if (statEl) {
          statEl.textContent = "0";
        }
      }
    } catch (error) {
      console.error("Error loading transactions:", error);
      transactionsList.innerHTML = "";
      const empty = document.createElement("div");
      empty.className = "text-center py-8 text-gray-400";
      const emptyIcon = document.createElement("i");
      emptyIcon.className = "fas fa-file-invoice text-4xl mb-3";
      const emptyText = document.createElement("p");
      emptyText.textContent = "Unable to load transactions";
      empty.appendChild(emptyIcon);
      empty.appendChild(emptyText);
      transactionsList.appendChild(empty);
      if (statEl) {
        statEl.textContent = "0";
      }
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

  /**
   * Trigger account reconciliation and display the result
   */
  async reconcileAccount(accountId, accountName) {
    try {
      const { response, data } = await api.reconcileAccount(accountId);

      if (response.ok) {
        ui.showToast(
          `${accountName} reconciled`,
          data.matched
            ? "Ledger and stored balance match"
            : "Balance mismatch detected",
          data.matched ? "success" : "error",
        );
      } else {
        ui.showToast(
          "Reconcile failed",
          data.error || "Please try again",
          "error",
        );
      }
    } catch (error) {
      ui.showToast("Network error", "Please try again", "error");
    }
  },

  /**
   * Fetch full transaction details for a history row
   */
  async viewTransactionDetails(transactionId) {
    try {
      const { response, data } = await api.getTransaction(transactionId);

      if (response.ok && Array.isArray(data)) {
        const totalDebit = data.reduce(
          (sum, entry) => sum + parseFloat(entry.debit || 0),
          0,
        );
        const totalCredit = data.reduce(
          (sum, entry) => sum + parseFloat(entry.credit || 0),
          0,
        );

        ui.showToast(
          "Transaction details",
          `${utils.truncate(transactionId)} | debit ${utils.formatCurrency(totalDebit)} | credit ${utils.formatCurrency(totalCredit)}`,
          "info",
        );
      } else {
        ui.showToast(
          "Transaction lookup failed",
          "Could not fetch transaction",
          "error",
        );
      }
    } catch (error) {
      ui.showToast("Network error", "Please try again", "error");
    }
  },
};
