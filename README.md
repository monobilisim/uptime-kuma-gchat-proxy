# [![Contributors][contributors-shield]][contributors-url]
[![Forks][forks-shield]][forks-url]
[![Stargazers][stars-shield]][stars-url]
[![Issues][issues-shield]][issues-url]
[![GPLv3 License][license-shield]][license-url]

<div align="center"> 

<a href="https://mono.net.tr/">
  <img src="https://monobilisim.com.tr/images/mono-bilisim.svg" width="340"/>
</a>

<h2 align="center">uptime-kuma-gchat-proxy</h2>

<b>uptime-kuma-gchat-proxy</b> is a lightweight webhook proxy service that converts Uptime Kuma notifications to Google Chat Card format. This middleware provides better preview and rich content display in mobile notifications.

</div>

---

## Table of Contents

- [Table of Contents](#table-of-contents)
- [Features](#features)
- [Installation](#installation)
- [Configuration](#configuration)
- [Usage](#usage)
- [API Endpoints](#api-endpoints)
- [Docker Deployment](#docker-deployment)
- [Troubleshooting](#troubleshooting)
- [Security](#security)
- [Building](#building)
- [License](#license)

---

## Features

- 🚀 **Lightweight and Fast** - Written in Go for optimal performance
- 📱 **Mobile Notification Preview** - Enhanced preview support for mobile notifications
- 🎨 **Rich Card Format** - Beautiful Google Chat Card format with structured information
- 🐳 **Docker Support** - Easy deployment with Docker and Docker Compose
- 🔄 **Auto Restart** - Automatic restart on failure
- 💚 **Health Check** - Built-in health check endpoint for monitoring
- 🔒 **Simple & Secure** - Minimal dependencies and secure webhook handling

---

## Installation

### Method 1: Docker Compose (Recommended)

1. Create a `.env` file:
```bash
cp .env.example .env
```

2. Edit the `.env` file and add your Google Chat webhook URL:
```bash
GOOGLE_CHAT_WEBHOOK_URL=https://chat.googleapis.com/v1/spaces/XXXXX/messages?key=XXXXX&token=XXXXX
```

3. Start the service:
```bash
docker-compose up -d
```

### Method 2: Docker

```bash
docker build -t uptime-kuma-gchat-proxy .
docker run -d \
  -p 8080:8080 \
  -e GOOGLE_CHAT_WEBHOOK_URL="your_webhook_url" \
  uptime-kuma-gchat-proxy
```

### Method 3: Direct Go Execution

```bash
# Download dependencies
go mod download

# Set environment variable
export GOOGLE_CHAT_WEBHOOK_URL="your_webhook_url"

# Run the application
go run main.go
```

---

## Configuration

### Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `GOOGLE_CHAT_WEBHOOK_URL` | Google Chat webhook URL | - | Yes |
| `PORT` | Port for the service to listen on | 8080 | No |

### Getting Google Chat Webhook URL

1. Open a Space in Google Chat
2. Click the ▼ next to the Space name
3. Select **Apps & integrations**
4. Click **Add webhook**
5. Enter a name for the webhook (e.g., "Uptime Kuma")
6. Click **Save**
7. Copy the webhook URL

### Uptime Kuma Configuration

1. Go to **Settings** > **Notifications** in Uptime Kuma
2. Click **Setup Notification**
3. Select **Webhook** as the **Notification Type**
4. Enter the following information:
   - **Friendly Name:** Google Chat (Proxy)
   - **Post URL:** `http://localhost:8080/webhook` (or your server address)
   - **Content Type:** `application/json`
5. Click **Test** to verify the connection
6. Click **Save**

---

## Usage

### Basic Usage

Once the service is running, configure Uptime Kuma to send webhooks to:
```
http://your-server:8080/webhook
```

The proxy will automatically:
- Receive Uptime Kuma notifications
- Convert them to Google Chat Card format
- Forward them to your Google Chat Space

### Card Format Features

The converted cards include:
- **Status Indicator** - 🟢 UP / 🔴 DOWN with emoji
- **Monitor Name** - Clear identification of the monitored service
- **URL** - Direct link to the monitored service
- **Response Time** - Ping/response time in milliseconds
- **Timestamp** - When the status change occurred
- **Visit Button** - Quick access button to visit the monitored site

### Example Scenarios

#### Scenario 1: Local Docker Usage
```bash
# Set .env file
echo 'GOOGLE_CHAT_WEBHOOK_URL=your_webhook_url' > .env

# Start service
docker-compose up -d

# Configure Uptime Kuma webhook URL
# http://localhost:8080/webhook
```

#### Scenario 2: Server Deployment
```bash
# Start service
docker-compose up -d

# Configure Uptime Kuma webhook URL
# http://your-server-ip:8080/webhook

# Or with domain
# https://uptime-proxy.yourdomain.com/webhook
```

#### Scenario 3: Same Docker Network as Uptime Kuma
```yaml
# In docker-compose.yml
services:
  uptime-kuma:
    # ...
    networks:
      - monitoring

  uptime-kuma-gchat-proxy:
    # ...
    networks:
      - monitoring

networks:
  monitoring:
```

Uptime Kuma webhook URL: `http://uptime-kuma-gchat-proxy:8080/webhook`

---

## API Endpoints

### POST /webhook

Receives webhooks from Uptime Kuma and forwards them to Google Chat.

**Request Body:**
```json
{
  "heartbeat": {
    "monitorID": 1,
    "status": 1,
    "time": "2025-11-07 12:00:00",
    "msg": "OK",
    "ping": 123.45
  },
  "monitor": {
    "id": 1,
    "name": "My Website",
    "url": "https://example.com",
    "hostname": "example.com",
    "port": 443,
    "type": "http"
  },
  "msg": "Service is up"
}
```

**Response:**
```
200 OK
```

### GET /health

Health check endpoint for monitoring the service status.

**Response:**
```
200 OK
```

---

## Docker Deployment

### Docker Compose

The included `docker-compose.yml` provides:
- Automatic restart on failure
- Health check configuration
- Port mapping
- Environment variable management

### Docker Images

Build your own image:
```bash
docker build -t uptime-kuma-gchat-proxy .
```

### Health Checks

The service includes a health check endpoint that can be used with Docker health checks:
```yaml
healthcheck:
  test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/health"]
  interval: 30s
  timeout: 10s
  retries: 3
  start_period: 40s
```

---

## Troubleshooting

### Notifications Not Arriving

1. Check if the service is running:
   ```bash
   curl http://localhost:8080/health
   ```
2. Verify the webhook URL is correct in Uptime Kuma
3. Test the Google Chat webhook URL directly
4. Check logs:
   ```bash
   docker-compose logs -f uptime-kuma-gchat-proxy
   ```

### No Preview in Mobile Notifications

- Verify that cards are being sent correctly (check logs)
- Ensure you're using a Google Chat Space (card preview may not work in DMs)
- Check that the webhook URL is valid and active

### Port Conflict

Change the `PORT` environment variable in `.env`:
```bash
PORT=9090
```

### Logging

View service logs:
```bash
# Docker logs
docker-compose logs -f uptime-kuma-gchat-proxy

# Filter errors only
docker-compose logs -f uptime-kuma-gchat-proxy | grep ERROR
```

---

## Security

- **Webhook URLs are sensitive** - Keep them secure and never commit them to version control
- **Use HTTPS in production** - Always use HTTPS for production deployments
- **Firewall rules** - Restrict access to only necessary IPs
- **Environment variables** - Store sensitive data in environment variables, not in code
- **Network isolation** - Use Docker networks to isolate services when possible

---

## Building

To build the application:

```bash
go build -o uptime-kuma-gchat-proxy main.go
```

Or using Docker:
```bash
docker build -t uptime-kuma-gchat-proxy .
```

---

## License

This project is released under the GNU General Public License v3.0.

- You may run, study, share, and modify the software, provided derivative works remain GPL-compatible.
- Source code for any distributed binaries must be made available to recipients.
- The software is provided “as is”, without warranty; see [`LICENSE`](LICENSE) for the full text.

[contributors-shield]: https://img.shields.io/github/contributors/monobilisim/uptime-kuma-gchat-proxy.svg?style=for-the-badge
[contributors-url]: https://github.com/monobilisim/uptime-kuma-gchat-proxy/graphs/contributors
[forks-shield]: https://img.shields.io/github/forks/monobilisim/uptime-kuma-gchat-proxy.svg?style=for-the-badge
[forks-url]: https://github.com/monobilisim/uptime-kuma-gchat-proxy/network/members
[stars-shield]: https://img.shields.io/github/stars/monobilisim/uptime-kuma-gchat-proxy.svg?style=for-the-badge
[stars-url]: https://github.com/monobilisim/uptime-kuma-gchat-proxy/stargazers
[issues-shield]: https://img.shields.io/github/issues/monobilisim/uptime-kuma-gchat-proxy.svg?style=for-the-badge
[issues-url]: https://github.com/monobilisim/uptime-kuma-gchat-proxy/issues
[license-shield]: https://img.shields.io/github/license/monobilisim/uptime-kuma-gchat-proxy.svg?style=for-the-badge
[license-url]: https://github.com/monobilisim/uptime-kuma-gchat-proxy/blob/master/LICENSE
