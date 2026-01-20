# CLAUDE.md

File Management System - A Go-based MCP server for managing files, folders, tags, and semantic search.

## Architecture

- **MCP Server**: `github.com/mark3labs/mcp-go` for AI tool integration
- **REST API**: Fiber framework with oapi-codegen generated handlers
- **Database**: Turso (production) with vector search / SQLite in-memory (testing) with GORM ORM
- **Authentication**: MCPRouter via FiberApikeyMiddleware or OAuth
- **File Storage**: S3-compatible storage (AWS S3, Cloudflare R2, MinIO)
- **Embeddings**: Vercel AI Gateway for text-embedding-3-small (1536 dimensions)
- **Content Parsing**: External Python service for document text extraction

## Data Models

### Tag

- `id` (uint) - Primary key
- `user_id` (string) - Index, required
- `name` (string) - Required
- `color` (varchar(7)) - Hex color, optional
- `created_at`, `updated_at`, `deleted_at` - Timestamps with soft delete

### Folder

- `id` (uint) - Primary key
- `user_id` (string) - Index, required
- `name` (string) - Required
- `description` (text) - Optional
- `parent_id` (uint\*) - Self-referential for tree structure
- `tags` - Many-to-many relationship via `folder_tags`
- `children` - Has many folders (self-referential)
- `created_at`, `updated_at`, `deleted_at` - Timestamps with soft delete

### File

- `id` (uint) - Primary key
- `user_id` (string) - Index, required
- `title` (string) - Required
- `summary` (text) - AI-generated summary
- `content` (text) - Parsed text content for search
- `file_type` (enum) - music, photo, video, document, invoice
- `folder_id` (uint\*) - Foreign key to folder
- `tags` - Many-to-many relationship via `file_tags`
- `s3_key` (string) - S3 object key
- `original_filename` (string) - Original upload filename
- `mime_type` (string) - MIME type
- `size` (int64) - File size in bytes
- `processing_status` (enum) - pending, processing, completed, failed
- `processing_error` (text) - Error message if processing failed
- `has_embedding` (bool) - Whether vector embedding exists
- `created_at`, `updated_at`, `deleted_at` - Timestamps with soft delete

### FileEmbedding

- `id` (uint) - Primary key
- `file_id` (uint) - Foreign key, unique
- `user_id` (string) - For user isolation
- `embedding` (F32_BLOB) - 1536-dimension vector for Turso vector search

## MCP Tools (25 total)

**Tags**: `create_tag`, `list_tags`, `get_tag`, `update_tag`, `delete_tag`
**Folders**: `create_folder`, `list_folders`, `get_folder`, `update_folder`, `delete_folder`, `move_folder`, `get_folder_tree`, `add_tags_to_folder`, `remove_tags_from_folder`
**Files**: `create_file`, `list_files`, `get_file`, `update_file`, `delete_file`, `move_files`, `add_tags_to_file`, `remove_tags_from_file`, `get_file_download_url`
**Search**: `search_files` (supports fulltext, semantic, hybrid)
**Upload**: `upload_file`

## API Endpoints

### Tags

- `POST /api/tags` - Create tag (201)
- `GET /api/tags` - List with search (`?keyword=`)
- `GET /api/tags/{id}` - Get by ID
- `PUT /api/tags/{id}` - Update
- `DELETE /api/tags/{id}` - Delete (204)

### Folders

- `POST /api/folders` - Create folder (201)
- `GET /api/folders` - List with filter (`?parent_id=`)
- `GET /api/folders/{id}` - Get by ID
- `PUT /api/folders/{id}` - Update
- `DELETE /api/folders/{id}` - Delete (204)
- `POST /api/folders/{id}/move` - Move folder to new parent
- `GET /api/folders/tree` - Get hierarchical tree structure
- `POST /api/folders/{id}/tags` - Add tags to folder
- `DELETE /api/folders/{id}/tags` - Remove tags from folder

### Files

- `POST /api/files` - Create file record (201)
- `GET /api/files` - List with filters (`?folder_id=`, `?file_type=`, `?keyword=`)
- `GET /api/files/{id}` - Get by ID
- `PUT /api/files/{id}` - Update
- `DELETE /api/files/{id}` - Delete (204)
- `POST /api/files/move` - Batch move files to folder
- `POST /api/files/{id}/tags` - Add tags to file
- `DELETE /api/files/{id}/tags` - Remove tags from file
- `GET /api/files/{id}/download` - Get presigned download URL
- `POST /api/files/{id}/process` - Trigger async content processing (202)

### Search

- `GET /api/search?q=...&type=fulltext|semantic|hybrid` - Search files

### Upload

- `POST /api/upload` - Upload file to S3 (201)
- `GET /api/upload/presigned?filename=...` - Get presigned upload URL

### Health

- `GET /health` - Health check (no auth)

## File Processing Flow

1. Upload file to S3 via `/api/upload` -> returns S3 key
2. Create file record via `POST /api/files` with S3 key (status: pending)
3. Trigger processing via `POST /api/files/{id}/process` -> returns 202 immediately
4. **Background goroutine:**
   - Update status to "processing"
   - Get presigned download URL for the S3 file
   - Call Python content parser with the URL
   - Store parsed content and summary in file record
   - Detect FileType from content (invoice detection)
   - Call Vercel AI Gateway to generate embedding (1536 dimensions)
   - Store embedding in file_embeddings table (Turso F32_BLOB)
   - Update status to "completed" (or "failed" with error message)
5. Client polls `GET /api/files/{id}` to check processing_status

## Search Types

- **fulltext**: LIKE search on title and content fields
- **semantic**: Turso vector_distance_cos on embeddings
- **hybrid**: Combines fulltext and vector results with weighted scoring

## Development Commands

```bash
# Build
make build       # Build the project
make run         # Run the server

# Code Generation
make generate    # Regenerate API handlers from OpenAPI spec

# Testing
make test        # Run all tests
go test ./e2e/api/... -v -timeout 30s  # Run E2E tests

# Code Quality
make fmt         # Format code
make lint        # Run linter
make deps        # Download dependencies
```

## File Structure

```
files-management/
├── cmd/
│   └── server/main.go              # Server entry point
├── internal/
│   ├── api/
│   │   ├── server.go               # Fiber server setup
│   │   ├── converters.go           # Model to generated type converters
│   │   ├── generated/              # oapi-codegen generated code
│   │   │   └── server.gen.go
│   │   ├── handlers/               # Strict handler implementations
│   │   │   ├── tag_handlers.go
│   │   │   ├── folder_handlers.go
│   │   │   ├── file_handlers.go
│   │   │   ├── search_handlers.go
│   │   │   └── upload_handlers.go
│   │   └── middleware/
│   │       └── auth.go             # Auth middleware
│   ├── assets/
│   │   └── openapi.yaml            # OpenAPI 3.0.3 specification
│   ├── mcp/
│   │   └── server.go               # MCP tools registration
│   ├── models/
│   │   ├── tag.go
│   │   ├── folder.go
│   │   ├── file.go
│   │   └── file_embedding.go
│   ├── services/
│   │   ├── db_service.go           # Turso/SQLite connection + migrations
│   │   ├── tag_service.go
│   │   ├── folder_service.go
│   │   ├── file_service.go
│   │   ├── search_service.go       # Fulltext, vector, hybrid search
│   │   ├── embedding_service.go    # Vercel AI Gateway integration
│   │   ├── content_parser_service.go  # Python parser integration
│   │   └── upload_service.go
│   ├── tools/                      # MCP tool implementations
│   │   ├── tag_tools.go
│   │   ├── folder_tools.go
│   │   ├── file_tools.go
│   │   ├── search_tools.go
│   │   └── upload_tools.go
│   └── utils/
│       ├── context.go              # Auth context helpers
│       └── jwt_authenticator.go    # JWT validation
├── e2e/
│   └── api/
│       ├── test_helpers.go
│       ├── auth_test.go
│       ├── tag_test.go
│       ├── folder_test.go
│       ├── file_test.go
│       ├── search_test.go
│       └── upload_test.go
├── k8s/
│   ├── deployment.yaml
│   ├── secrets.yaml
│   └── service.yaml
├── go.mod
├── Makefile
└── CLAUDE.md
```

## Environment Variables

```bash
# Database (Turso with vector support)
TURSO_DATABASE_URL=libsql://your-db.turso.io
TURSO_AUTH_TOKEN=your-token

# Local SQLite (fallback)
SQLITE_DB_PATH=files.db

# S3-compatible storage
S3_ENDPOINT=https://s3.amazonaws.com
S3_BUCKET=files-management
S3_ACCESS_KEY=your-access-key
S3_SECRET_KEY=your-secret-key
S3_REGION=us-east-1
S3_USE_PATH_STYLE=false

# Authentication
MCPROUTER_SERVER_URL=https://your-mcprouter.com
MCPROUTER_SERVER_API_KEY=your-api-key
OAUTH_SERVER_URL=https://your-oauth-provider.com

# AI Gateway (for embeddings)
AI_GATEWAY_URL=https://ai-gateway.vercel.sh/v1
AI_GATEWAY_API_KEY=your-key
EMBEDDING_MODEL=openai/text-embedding-3-small
EMBEDDING_DIMENSIONS=1536

# Content Parser Service
CONTENT_PARSER_ENDPOINT=https://your-python-service/convert
ADMIN_API_KEY=your-admin-key

# Server
PORT=8080
```

## Authentication

The system supports multiple authentication methods:

1. **MCPRouter**: API key-based authentication via `FiberApikeyMiddleware`
2. **OAuth/JWKS**: OAuth 2.0 with JWKS validation via `OAUTH_SERVER_URL`

Authentication is optional - if no authentication environment variables are set, the API runs without authentication.

### Authentication Flow

#### HTTP API Authentication

1. Client sends request with `Authorization: Bearer <jwt-token>` header
2. Auth middleware validates token using JwtAuthenticator
3. Authenticated user stored in Fiber context via `c.Locals(AuthenticatedUserContextKey, user)`
4. Handlers access user with: `user := c.Locals(AuthenticatedUserContextKey).(*utils.AuthenticatedUser)`

#### MCP Tool Authentication

1. HTTP request to `/mcp/*` endpoints includes `Authorization: Bearer <jwt-token>` header
2. Custom handler extracts and validates JWT token
3. Authenticated user added to Go context via `utils.WithAuthenticatedUser(ctx, user)`
4. MCP tools access user with: `user, ok := utils.GetAuthenticatedUser(ctx)`

## Testing

Tests use in-memory SQLite databases and mock services:

```go
// Create test setup
setup := NewTestSetup(t)
defer setup.Cleanup()

// Make authenticated requests
resp, err := setup.MakeRequest("POST", "/api/tags", payload)
```

Test authentication is handled via `X-Test-User-ID` header in tests.

### Running Tests

```bash
# Run all tests with 30s timeout
go test ./... -timeout 30s

# Run E2E API tests
go test ./e2e/api/... -v -timeout 30s

# Run specific test suite
go test ./e2e/api -run TestTagSuite -v -timeout 30s
go test ./e2e/api -run TestFolderSuite -v -timeout 30s
go test ./e2e/api -run TestFileSuite -v -timeout 30s
go test ./e2e/api -run TestSearchSuite -v -timeout 30s
```

## Tool Implementation Pattern

All MCP tools follow a consistent structure:

```go
type CreateTagTool struct {
    service services.TagService
}

func NewCreateTagTool(service services.TagService) *CreateTagTool {
    return &CreateTagTool{service: service}
}

func (t *CreateTagTool) GetTool() mcp.Tool {
    return mcp.NewTool("create_tag",
        mcp.WithDescription("Create a new tag"),
        mcp.WithString("name", mcp.Required(), mcp.Description("Tag name")),
        mcp.WithString("color", mcp.Description("Hex color code")),
    )
}

func (t *CreateTagTool) GetHandler() server.ToolHandlerFunc {
    return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        userID := getUserIDFromContext(ctx)
        if userID == "" {
            return mcp.NewToolResultError("Authentication required"), nil
        }

        args := getArgsMap(request.Params.Arguments)
        // ... implementation
    }
}
```

## Code Guidelines

1. Never use `fmt.Println` for logging - use structured logging
2. Test timeout policy: Never run tests longer than 30 seconds
3. All handlers require user authentication via context
4. Services are user-scoped - all operations filter by `user_id`
5. Use service methods for data access, never raw GORM queries in handlers/tools
6. OpenAPI spec is the source of truth - run `make generate` after changes

## Key Dependencies

- `github.com/mark3labs/mcp-go` - MCP server framework
- `github.com/gofiber/fiber/v2` - HTTP framework
- `github.com/oapi-codegen/oapi-codegen` - OpenAPI code generator
- `gorm.io/gorm` and `gorm.io/driver/sqlite` - ORM and database
- `github.com/tursodatabase/libsql-client-go/libsql` - Turso client with vector support
- `github.com/aws/aws-sdk-go-v2` - S3-compatible storage
- `github.com/rxtech-lab/mcprouter-authenticator` - MCPRouter authentication
- `github.com/stretchr/testify` - Testing utilities

## Turso Vector Search

The system uses Turso's native vector search capabilities:

```sql
-- Create embeddings table
CREATE TABLE file_embeddings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    file_id INTEGER NOT NULL UNIQUE,
    user_id TEXT NOT NULL,
    embedding F32_BLOB(1536),
    FOREIGN KEY (file_id) REFERENCES files(id)
);

-- Vector index
CREATE INDEX file_embeddings_idx ON file_embeddings(libsql_vector_idx(embedding));

-- Search query
SELECT f.*, vector_distance_cos(fe.embedding, vector32(?)) AS distance
FROM file_embeddings fe
JOIN files f ON f.id = fe.file_id
WHERE f.user_id = ? AND f.deleted_at IS NULL
ORDER BY distance ASC LIMIT ?;
```
