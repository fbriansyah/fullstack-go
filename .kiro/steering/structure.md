# Project Structure

## Standard Go Fullstack Layout

```
/
├── cmd/                    # Main applications
│   └── server/            # Server entry point
│       └── main.go
├── internal/              # Private application code
│   ├── modules/          # Business modules (modular monolith)
│   │   ├── user/         # User management module
│   │   ├── auth/         # Authentication module
│   │   └── notification/ # Notification module (future)
│   ├── shared/           # Shared infrastructure
│   │   ├── events/       # Event bus and RabbitMQ integration
│   │   ├── database/     # Database connection and migrations
│   │   ├── middleware/   # HTTP middleware
│   │   └── errors/       # Error handling utilities
│   └── config/           # Configuration management
├── pkg/                   # Public library code
├── web/                   # Frontend assets
│   ├── static/           # Static files (CSS, JS, images)
│   ├── templates/        # Templ template files (.templ)
│   └── components/       # Reusable Templ components
├── migrations/            # Database migrations
├── tests/                 # Integration and e2e tests
├── docker/               # Docker-related files
├── docs/                 # Documentation
├── .air.toml             # Air configuration for hot reload
├── docker-compose.yml    # Local development setup
├── Dockerfile            # Container definition
├── go.mod                # Go module definition
├── go.sum                # Go module checksums
└── README.md             # Project documentation
```

## Naming Conventions

### Go Code
- **Packages**: lowercase, single word when possible
- **Files**: snake_case for multi-word files
- **Functions**: PascalCase for exported, camelCase for private
- **Variables**: camelCase
- **Constants**: PascalCase or SCREAMING_SNAKE_CASE for package-level

### API Endpoints
- RESTful naming: `/api/v1/users`, `/api/v1/users/{id}`
- Use HTTP verbs appropriately (GET, POST, PUT, DELETE)
- Version your APIs: `/api/v1/`

### Database
- **Tables**: snake_case, plural nouns (`users`, `user_profiles`)
- **Columns**: snake_case (`created_at`, `user_id`)
- **Indexes**: descriptive names (`idx_users_email`)

## File Organization Rules

1. **Separation of Concerns**: Keep business logic in `services/`, data access in `repository/`
2. **Handler Pattern**: API handlers should be thin, delegating to services
3. **Configuration**: Environment-based config in `internal/config/`
4. **Testing**: Test files alongside source files with `_test.go` suffix
5. **Static Assets**: Frontend builds go to `web/dist/`, source in `web/static/`