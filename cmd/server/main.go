package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/rxtech-lab/invoice-management/internal/api"
	mcpserver "github.com/rxtech-lab/invoice-management/internal/mcp"
	"github.com/rxtech-lab/invoice-management/internal/services"
	"gorm.io/gorm"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found, using environment variables")
	}

	// Validate required environment variables
	validateRequiredEnvVars()

	// Initialize database
	dbService, err := initDatabase()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer dbService.Close()

	// Get the underlying GORM DB for service creation
	db := dbService.GetDB()

	// Initialize services
	tagService := services.NewTagService(db)
	folderService := services.NewFolderService(db)
	fileService := services.NewFileService(db)
	uploadService := initUploadService()
	embeddingService := initEmbeddingService(db)
	contentParserService := initContentParserService()
	summaryService := initSummaryService()
	searchService := services.NewSearchService(db, embeddingService)
	agentService := initAgentService(tagService, fileService, folderService)

	// Initialize MCP server
	mcpSrv := mcpserver.NewMCPServer(
		dbService,
		tagService,
		folderService,
		fileService,
		uploadService,
		searchService,
	)

	// Initialize API server
	port := getEnvOrDefault("PORT", "8080")
	apiServer := api.NewAPIServer(
		dbService,
		tagService,
		folderService,
		fileService,
		uploadService,
		embeddingService,
		contentParserService,
		searchService,
		summaryService,
		agentService,
		mcpSrv.GetServer(),
	)

	// Enable authentication if configured (must be before routes)
	if os.Getenv("OAUTH_SERVER_URL") != "" || os.Getenv("MCPROUTER_SERVER_URL") != "" {
		if err := apiServer.EnableAuthentication(); err != nil {
			log.Printf("Warning: Failed to enable authentication: %v", err)
		} else {
			log.Println("Authentication enabled")
		}
	}

	// Setup routes (after authentication middleware)
	apiServer.SetupRoutes()

	// Enable StreamableHTTP for MCP
	apiServer.EnableStreamableHTTP()

	// Handle graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("Shutting down server...")
		cancel()
		if err := apiServer.Shutdown(); err != nil {
			log.Printf("Error during shutdown: %v", err)
		}
	}()

	// Start server
	log.Printf("Starting server on port %s", port)
	if err := apiServer.Start(":" + port); err != nil {
		log.Fatalf("Server error: %v", err)
	}

	<-ctx.Done()
	log.Println("Server stopped")
}

func initDatabase() (services.DBService, error) {
	tursoURL := os.Getenv("TURSO_DATABASE_URL")
	tursoToken := os.Getenv("TURSO_AUTH_TOKEN")

	if tursoURL != "" {
		log.Println("Connecting to Turso database...")
		return services.NewTursoDBService(tursoURL, tursoToken)
	}

	// Fall back to local SQLite
	dbPath := getEnvOrDefault("SQLITE_DB_PATH", "files.db")
	log.Printf("Using local SQLite database: %s", dbPath)
	return services.NewSqliteDBService(dbPath)
}

func initUploadService() services.UploadService {
	bucket := os.Getenv("S3_BUCKET")

	if bucket == "" {
		log.Println("Warning: S3_BUCKET not configured, file uploads will not work")
		return nil
	}

	cfg := services.S3Config{
		Endpoint:        os.Getenv("S3_ENDPOINT"),
		Bucket:          bucket,
		AccessKeyID:     os.Getenv("S3_ACCESS_KEY"),
		SecretAccessKey: os.Getenv("S3_SECRET_KEY"),
		Region:          getEnvOrDefault("S3_REGION", "us-east-1"),
		UsePathStyle:    os.Getenv("S3_USE_PATH_STYLE") == "true",
	}

	service, err := services.NewUploadService(cfg)
	if err != nil {
		log.Printf("Warning: Failed to initialize S3 upload service: %v", err)
		return nil
	}

	log.Printf("S3 upload service initialized (bucket: %s)", bucket)
	return service
}

func initEmbeddingService(db *gorm.DB) services.EmbeddingService {
	gatewayURL := os.Getenv("AI_GATEWAY_URL")
	apiKey := os.Getenv("AI_GATEWAY_API_KEY")

	model := getEnvOrDefault("EMBEDDING_MODEL", "text-embedding-3-small")
	dimensions := 1536
	if dimStr := os.Getenv("EMBEDDING_DIMENSIONS"); dimStr != "" {
		if dim, err := strconv.Atoi(dimStr); err == nil {
			dimensions = dim
		}
	}

	config := services.EmbeddingConfig{
		GatewayURL: gatewayURL,
		APIKey:     apiKey,
		Model:      model,
		Dimensions: dimensions,
	}

	log.Printf("Embedding service initialized (model: %s, dimensions: %d)", model, dimensions)
	return services.NewEmbeddingService(db, config)
}

func initContentParserService() services.ContentParserService {
	endpoint := os.Getenv("CONTENT_PARSER_ENDPOINT")
	apiKey := os.Getenv("ADMIN_API_KEY")

	config := services.ContentParserConfig{
		EndpointURL: endpoint,
		APIKey:      apiKey,
	}

	log.Printf("Content parser service initialized (endpoint: %s)", endpoint)
	return services.NewContentParserService(config)
}

func initSummaryService() services.SummaryService {
	gatewayURL := os.Getenv("AI_GATEWAY_URL")
	apiKey := os.Getenv("AI_GATEWAY_API_KEY")
	model := getEnvOrDefault("SUMMARY_MODEL", "gpt-4o-mini")

	config := services.SummaryConfig{
		GatewayURL: gatewayURL,
		APIKey:     apiKey,
		Model:      model,
	}

	log.Printf("Summary service initialized (model: %s)", model)
	return services.NewSummaryService(config)
}

func initAgentService(
	tagService services.TagService,
	fileService services.FileService,
	folderService services.FolderService,
) services.AgentService {
	// Check if agent is enabled (default: true)
	enabled := os.Getenv("AGENT_ENABLED") != "false"
	if !enabled {
		log.Println("AI Agent service disabled")
		return nil
	}

	gatewayURL := os.Getenv("AI_GATEWAY_URL")
	apiKey := os.Getenv("AI_GATEWAY_API_KEY")
	model := getEnvOrDefault("AGENT_MODEL", "gpt-4o-mini")

	maxTurns := 10
	if maxTurnsStr := os.Getenv("AGENT_MAX_TURNS"); maxTurnsStr != "" {
		if mt, err := strconv.Atoi(maxTurnsStr); err == nil && mt > 0 {
			maxTurns = mt
		}
	}

	config := services.AgentConfig{
		GatewayURL: gatewayURL,
		APIKey:     apiKey,
		Model:      model,
		MaxTurns:   maxTurns,
		Enabled:    enabled,
	}

	log.Printf("AI Agent service initialized (model: %s, maxTurns: %d)", model, maxTurns)
	return services.NewAgentService(config, tagService, fileService, folderService)
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func validateRequiredEnvVars() {
	required := map[string]string{
		"AI_GATEWAY_URL":          os.Getenv("AI_GATEWAY_URL"),
		"AI_GATEWAY_API_KEY":      os.Getenv("AI_GATEWAY_API_KEY"),
		"CONTENT_PARSER_ENDPOINT": os.Getenv("CONTENT_PARSER_ENDPOINT"),
	}

	var missing []string
	for name, value := range required {
		if value == "" {
			missing = append(missing, name)
		}
	}

	if len(missing) > 0 {
		sort.Strings(missing)
		log.Fatalf("Missing required environment variables: %s", strings.Join(missing, ", "))
	}
}
