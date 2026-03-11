# Frontend Demo - Double-Entry Bank Ledger

Modern web interface for the banking ledger API built with vanilla JavaScript, Tailwind CSS, and Font Awesome.

## Features

###  Modern UI
- Glassmorphism design with gradient backgrounds
- Responsive layout (mobile & desktop)
- Smooth animations and transitions
- Toast notifications for user feedback

###  Authentication
- User registration with email/password
- Secure login with JWT tokens
- Token persistence (localStorage)
- Auto-logout on token expiration

###  Banking Operations
- **Account Management**: Create and view multiple accounts
- **Deposits**: Add funds to accounts
- **Withdrawals**: Remove funds from accounts
- **Transfers**: Move money between accounts
- **Transaction History**: View all account entries

###  Dashboard
- Total accounts count
- Combined balance across all accounts
- Transaction counter
- Real-time updates after operations

## Technology Stack

- **HTML5**: Semantic markup
- **Tailwind CSS** (CDN): Utility-first styling
- **Font Awesome** (CDN): Icons
- **Vanilla JavaScript**: No frameworks, pure ES6+
- **Fetch API**: RESTful API communication

## File Structure

```
frontend/
├── index.html      # Main HTML structure
├── styles.css      # Custom CSS (animations, effects)
├── js/             # Modular JavaScript files
│   ├── config.js   # Configuration and constants
│   ├── state.js    # State management
│   ├── api.js      # API service layer
│   ├── auth.js     # Authentication logic
│   ├── dashboard.js # Dashboard operations
│   ├── transactions.js # Transaction handlers
│   ├── ui.js       # UI helper functions
│   ├── utils.js    # Utility functions
│   └── main.js     # Application entry point
└── README.md       # This file
```

## API Integration

The frontend communicates with the Go backend at `http://localhost:8080`:

### Endpoints Used
- `POST /register` - Create new user
- `POST /login` - Authenticate user
- `GET /accounts` - List user accounts
- `POST /accounts` - Create new account
- `POST /accounts/{id}/deposit` - Deposit funds
- `POST /accounts/{id}/withdraw` - Withdraw funds
- `POST /transfers` - Transfer between accounts
- `GET /accounts/{id}/entries` - Get transaction history

## State Management

```javascript
currentUser = {
    email: '',      // User email
    token: '',      // JWT token
    accounts: []    // Array of account objects
}
```

## Local Storage

- `token`: JWT authentication token
- `email`: User email for display

## Development Notes

### CORS Configuration
CORS is required only when the frontend is served from a different origin than the backend (for example, frontend at `http://localhost:3000` and API at `http://localhost:8080`). When both frontend and backend are served from `http://localhost:8080`, CORS is not involved.

### API Base URL
Currently hardcoded to `http://localhost:8080`. For production, use environment variables or config file.

### Security Considerations
- Tokens stored in localStorage (consider httpOnly cookies for production)
- No input sanitization (backend handles validation)
- Password minimum length: 6 characters

## Future Enhancements

- [ ] Account reconciliation UI
- [ ] Advanced transaction filtering and search
- [ ] Export transaction history (CSV, PDF)
- [ ] Multi-currency support display
- [ ] Dark/light theme toggle
- [ ] Transaction analytics charts (Chart.js)
- [ ] Notification preferences
- [ ] Two-factor authentication UI

## Browser Support

Tested on:
- ✅ Chrome 90+
- ✅ Firefox 88+
- ✅ Safari 14+
- ✅ Edge 90+

Requires ES6+ support (modern browsers only).

## Quick Start

1. Start the Go backend server
2. Navigate to http://localhost:8080
3. Register a new account
4. Create accounts and start transacting!
