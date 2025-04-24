# Bridgetunes MTN MyNumba Don Win - Backend Package

This repository contains the backend implementation for the Bridgetunes MTN MyNumba Don Win promotion platform.

## Project Structure

The project follows a clean architecture approach with the following structure:

```
├── api/
│   └── routes/           # API route definitions
├── cmd/
│   └── api/              # Application entry points
├── internal/
│   ├── config/           # Configuration handling
│   ├── handlers/         # HTTP request handlers
│   ├── middleware/       # HTTP middleware
│   ├── models/           # Data models
│   ├── repositories/     # Data access layer
│   ├── services/         # Business logic
│   └── utils/            # Utility functions
├── pkg/
│   ├── mongodb/          # MongoDB client
│   ├── mtnapi/           # MTN API client
│   └── smsgateway/       # SMS gateway interfaces
├── .env.example          # Example environment variables
├── Dockerfile            # Docker configuration
├── docker-compose.yml    # Docker Compose configuration
├── go.mod                # Go module definition
└── README.md             # Project documentation
```

## Features

- User management with opt-in/opt-out functionality
- Topup processing and points allocation
- Draw management with configurable number ending selection
- Winner selection and prize distribution
- Notification management with SMS gateway integration
- Template management for notifications
- Campaign management for bulk notifications
- JWT authentication for API security

## Prerequisites

- Go 1.19 or higher
- MongoDB Atlas account
- Docker and Docker Compose (optional, for containerized deployment)

## Configuration

Create a `.env` file in the root directory with the following variables:

```
# Server configuration
SERVER_PORT=8080
SERVER_MODE=debug
SERVER_ALLOWED_HOSTS=*

# MongoDB configuration
MONGODB_URI=mongodb+srv://username:password@cluster.mongodb.net/?retryWrites=true&w=majority
MONGODB_DATABASE=bridgetunes

# JWT configuration
JWT_SECRET=your-secret-key
JWT_EXPIRES_IN=86400

# MTN API configuration
MTN_BASE_URL=https://api.mtn.com
MTN_API_KEY=your-api-key
MTN_API_SECRET=your-api-secret
MTN_MOCK_API=true

# SMS gateway configuration
SMS_DEFAULT_GATEWAY=MTN
SMS_MOCK_SMS=true
SMS_MTN_BASE_URL=https://sms.mtn.com
SMS_MTN_API_KEY=your-api-key
SMS_MTN_API_SECRET=your-api-secret
SMS_KODOBE_BASE_URL=https://api.kodobe.net
SMS_KODOBE_API_KEY=your-api-key
```

## Getting Started

### Local Development

1. Clone the repository:
```bash
git clone https://github.com/bridgetunes/mtn-backend.git
cd mtn-backend
```

2. Install dependencies:
```bash
go mod download
```

3. Create a `.env` file with your configuration (see above)

4. Run the application:
```bash
go run cmd/api/main.go
```

The API will be available at http://localhost:8080

### Docker Deployment

1. Clone the repository:
```bash
git clone https://github.com/bridgetunes/mtn-backend.git
cd mtn-backend
```

2. Create a `.env` file with your configuration (see above)

3. Build and run with Docker Compose:
```bash
docker-compose up -d
```

The API will be available at http://localhost:8080

## API Documentation

### Authentication

- `POST /api/v1/auth/login` - Login and get JWT token

### User Management

- `GET /api/v1/users` - Get all users
- `GET /api/v1/users/:id` - Get user by ID
- `GET /api/v1/users/msisdn/:msisdn` - Get user by MSISDN
- `POST /api/v1/users` - Create a new user
- `PUT /api/v1/users/:id` - Update a user
- `DELETE /api/v1/users/:id` - Delete a user
- `POST /api/v1/opt-in` - Opt in to the promotion
- `POST /api/v1/opt-out` - Opt out of the promotion

### Topup Management

- `GET /api/v1/topups` - Get topups by date range
- `GET /api/v1/topups/:id` - Get topup by ID
- `GET /api/v1/topups/msisdn/:msisdn` - Get topups by MSISDN
- `POST /api/v1/topups` - Create a new topup
- `POST /api/v1/topups/process` - Process topups from MTN API

### Draw Management

- `GET /api/v1/draws` - Get draws by date range
- `GET /api/v1/draws/:id` - Get draw by ID
- `GET /api/v1/draws/date/:date` - Get draw by date
- `GET /api/v1/draws/status/:status` - Get draws by status
- `GET /api/v1/draws/default-digits/:day` - Get default eligible digits for a day
- `POST /api/v1/draws/schedule` - Schedule a new draw
- `POST /api/v1/draws/:id/execute` - Execute a scheduled draw

### Notification Management

- `GET /api/v1/notifications` - Get notifications by status
- `GET /api/v1/notifications/:id` - Get notification by ID
- `GET /api/v1/notifications/msisdn/:msisdn` - Get notifications by MSISDN
- `GET /api/v1/notifications/campaign/:id` - Get notifications by campaign ID
- `GET /api/v1/notifications/status/:status` - Get notifications by status
- `POST /api/v1/notifications/send-sms` - Send an SMS notification

#### Campaign Management

- `POST /api/v1/notifications/campaigns` - Create a new campaign
- `POST /api/v1/notifications/campaigns/:id/execute` - Execute a campaign

#### Template Management

- `GET /api/v1/notifications/templates` - Get all templates
- `GET /api/v1/notifications/templates/:id` - Get template by ID
- `GET /api/v1/notifications/templates/name/:name` - Get template by name
- `GET /api/v1/notifications/templates/type/:type` - Get templates by type
- `POST /api/v1/notifications/templates` - Create a new template
- `PUT /api/v1/notifications/templates/:id` - Update a template
- `DELETE /api/v1/notifications/templates/:id` - Delete a template

## License

This project is proprietary and confidential. Unauthorized copying, distribution, or use is strictly prohibited.
