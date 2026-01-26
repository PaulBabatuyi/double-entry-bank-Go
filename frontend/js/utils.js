/**
 * Utility Functions
 * Reusable helper functions
 */

const utils = {
  /**
   * Format a number as currency
   */
  formatCurrency(amount) {
    const num = parseFloat(amount) || 0;
    return (
      CURRENCY.SYMBOL +
      num.toLocaleString(CURRENCY.LOCALE, {
        minimumFractionDigits: 2,
        maximumFractionDigits: 2,
      })
    );
  },

  /**
   * Truncate a string to a specified length
   */
  truncate(str, length = 8) {
    if (!str || str.length <= length) return str;
    return str.substring(0, length) + "...";
  },

  /**
   * Format a date
   */
  formatDate(dateString) {
    return new Date(dateString).toLocaleString();
  },

  /**
   * Validate email format
   */
  isValidEmail(email) {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return emailRegex.test(email);
  },

  /**
   * Validate amount
   */
  isValidAmount(amount) {
    const num = parseFloat(amount);
    return !isNaN(num) && num > 0;
  },
};
