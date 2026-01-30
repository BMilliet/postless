# Postless 🚀

A lightweight, terminal-based HTTP client with a beautiful TUI (Terminal User Interface). Think Postman, but in your terminal.

![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)
![License](https://img.shields.io/badge/license-MIT-green)

## ✨ Features

- 🎨 **Beautiful TUI** - Built with Bubbletea and Lipgloss
- 📁 **Collection-based** - Organize requests into collections
- 🔄 **Live Editing** - Edit request bodies on the fly without modifying files
- 🔐 **JWT Support** - Automatic JWT token management (stored separately in `secret.json`)
- ⚡ **Fast** - Lightweight Go binary, instant startup
- 🎯 **Fuzzy Search** - Quickly find requests with fuzzy matching
- 📝 **JSON Support** - Pretty-printed JSON responses with syntax highlighting
- ⌨️ **Keyboard-driven** - Navigate entirely with your keyboard
- 🔧 **Configurable** - Per-project settings with global defaults

## 🚀 Quick Start

### Installation

#### Prerequisites
- Go 1.25.6 or higher
- Git

#### Build from Source

```bash
# Clone the repository
git clone https://github.com/BMilliet/postless.git
cd postless

# Install dependencies
make deps

# Build and install
make build
```

The binary will be installed to `~/.postless/postless`. Add it to your PATH:

```bash
export PATH="$HOME/.postless:$PATH"
```

### Setup

1. **Initialize your project** - Create a `.postless` directory in your project root:

```bash
mkdir -p .postless/requests
```

2. **Create configuration** - Add `config.json`:

```json
{
  "baseUrl": "http://localhost:3000",
  "timeout": 30,
  "globalHeaders": {
    "Content-Type": "application/json"
  }
}
```

## 📁 Project Structure

```
.postless/
├── config.json           # Base URL, timeout, global headers
├── secret.json          # JWT token (auto-created, add to .gitignore!)
└── requests/            # Your request collections
    ├── auth/
    │   ├── login.json
    │   └── signup.json
    └── users/
        ├── get-user.json
        └── update-user.json
```

## 📝 Request Format

Create JSON files in `.postless/requests/[collection]/`:

```json
{
  "name": "Login User",
  "method": "POST",
  "url": "{{baseUrl}}/login",
  "skipAuth": true,
  "headers": {
    "X-Custom-Header": "value"
  },
  "body": {
    "email": "user@example.com",
    "password": "secret123"
  }
}
```

### Fields

- **name** (required) - Display name for the request
- **method** (required) - HTTP method: `GET`, `POST`, `PUT`, `DELETE`, `PATCH`
- **url** (required) - Endpoint URL (supports `{{baseUrl}}` variable)
- **skipAuth** (optional) - Set to `true` to skip JWT token header
- **headers** (optional) - Custom headers (overrides global headers)
- **body** (optional) - Request body (JSON object)

## 🎮 Usage

### Launch Postless

```bash
cd your-project
postless
```

### Keyboard Shortcuts

#### Navigation
- `←/→` or `h/l` - Switch between collections
- `↑/↓` or `j/k` - Navigate requests
- `/` - Search (fuzzy matching)
- `ESC` - Exit search / Cancel
- `q` - Quit

#### Actions
- `ENTER` - Execute selected request
- `e` - Edit request body fields
- `ENTER` (in settings) - Edit setting value

### Workflow Example

1. **Launch** - `postless` from your project directory
2. **Navigate** - Use arrow keys to browse collections and requests
3. **Preview** - Select a request to see details
4. **Edit** (optional) - Press `e` to edit body fields
5. **Execute** - Press `ENTER` to send the request
6. **View Response** - See status, headers, and formatted JSON body

## ⚙️ Configuration

### config.json

```json
{
  "baseUrl": "http://localhost:3000",
  "timeout": 30,
  "globalHeaders": {
    "Content-Type": "application/json",
    "X-API-Version": "v1"
  }
}
```

**Fields:**
- `baseUrl` (required) - Base URL for all requests
- `timeout` (optional) - Request timeout in seconds (default: 30)
- `globalHeaders` (optional) - Headers added to all requests

### secret.json

Auto-created on first run. Stores sensitive data separately:

```json
{
  "jwt": "your-jwt-token-here"
}
```

### Settings Page

Access the settings page (last tab in UI) to edit:
- Base URL
- JWT Token
- Timeout

Changes are saved immediately to the respective files.

## 🎨 Features in Detail

### Body Editor

Edit request bodies without modifying JSON files:

1. Select a request
2. Press `e` to enter edit mode
3. Navigate fields with arrow keys
4. Select a field and enter new value
5. Press `ENTER` to save
6. Changes persist until app restart

### JWT Management

- JWT token is stored in `secret.json` (separate from config)
- Automatically added as `Authorization: Bearer {token}` header
- Skip JWT for specific requests with `"skipAuth": true`
- Edit JWT via Settings page or directly in `secret.json`

### Response Display

- **Status Code** - Color-coded (green=2xx, coral=4xx, red=5xx)
- **Headers** - All response headers displayed
- **Body** - Pretty-printed JSON with syntax highlighting
- **Metadata** - Duration, size, timestamp

### Collections

- Organize requests by feature (auth, users, posts, etc.)
- Each subdirectory in `requests/` becomes a collection
- Collections appear as tabs in the UI
- Easy navigation with left/right arrows

## 🛠️ Development

### Run in Development

```bash
make run
```

### Format Code

```bash
make fmt
```

## 📋 Examples

### Example: Authentication Flow

**.postless/requests/auth/signup.json**
```json
{
  "name": "Sign Up",
  "method": "POST",
  "url": "{{baseUrl}}/signup",
  "skipAuth": true,
  "body": {
    "email": "new@example.com",
    "password": "secret123"
  }
}
```

**.postless/requests/auth/login.json**
```json
{
  "name": "Login",
  "method": "POST",
  "url": "{{baseUrl}}/login",
  "skipAuth": true,
  "body": {
    "email": "new@example.com",
    "password": "secret123"
  }
}
```

**.postless/requests/users/get-profile.json**
```json
{
  "name": "Get My Profile",
  "method": "GET",
  "url": "{{baseUrl}}/me",
  "skipAuth": false
}
```

## 📄 License

MIT License - See LICENSE file for details

## 🙏 Acknowledgments

- Built with [Bubbletea](https://github.com/charmbracelet/bubbletea) - TUI framework
- Styled with [Lipgloss](https://github.com/charmbracelet/lipgloss) - Terminal styling

---

**For developers who love the terminal**
