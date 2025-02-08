package main

import (
	"context"
	"encoding/json"
	"fmt"
	"healthmatefood-api/config"
	"healthmatefood-api/database"
	"healthmatefood-api/middleware"
	"healthmatefood-api/route"
	"log"
	"net"
	"os"
	"os/signal"

	azureai_repository "healthmatefood-api/service/azureai/repository"

	auth_repository "healthmatefood-api/service/auth/repository"
	user_handler "healthmatefood-api/service/user/http"
	user_repository "healthmatefood-api/service/user/repository"
	user_usecase "healthmatefood-api/service/user/usecase"
	user_validator "healthmatefood-api/service/user/validator"

	file_usecase "healthmatefood-api/service/file/usecase"

	_ "healthmatefood-api/docs"

	"github.com/Pheethy/sqlx"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
	openai "github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
	_ "github.com/swaggo/files"
	"google.golang.org/grpc"
)

func envPath() string {
	if len(os.Args) == 1 {
		return ".env"
	} else {
		return os.Args[1]
	}
}

func main() {
	ctx := context.Background()
	cfg := config.LoadConfig(envPath())
	azureConfig := openai.DefaultAzureConfig(cfg.AzureAI().AzureAPIKey(), cfg.AzureAI().AzureEndpoint())
	azureConfig.APIVersion = cfg.AzureAI().APIVersion()
	client := openai.NewClientWithConfig(azureConfig)

	/* Init Tracing*/
	// tracer, closer := utils_tracing.Init("healthmatefood-api")
	// defer func(c io.Closer) {
	// 	if err := c.Close(); err != nil {
	// 		logrus.Fatal(err)
	// 	}
	// }(closer)
	// opentracing.SetGlobalTracer(tracer)
	/* Database Connection */
	psqlDB := database.DBConnect(ctx, cfg.Db(), nil)
	defer func(sql *sqlx.DB) {
		if err := sql.Close(); err != nil {
			logrus.Fatal(err)
		}
	}(psqlDB)

	/* Init Repository */
	userRepo := user_repository.NewUserRepository(psqlDB)
	azureAIRepo := azureai_repository.NewAzureAIRepository(client, cfg.AzureAI())
	authRepo := auth_repository.NewAuthRepository(cfg.Jwt(), psqlDB)
	_ = azureAIRepo
	resp, err := azureAIRepo.GetChatCompletion(context.Background(), "test")
	if err != nil {
		log.Println(err)
	}
	fmt.Println("resp ai", resp)

	/* Init Usecase */
	fileUs := file_usecase.NewFileUsecase(cfg)
	userUs := user_usecase.NewUserUsecase(cfg, userRepo, fileUs, authRepo)

	/* Init Handler */
	userHand := user_handler.NewUserHandler(userUs)

	/* Init Validate */
	userValidate := user_validator.Validation{}

	/* Init Fiber Server */
	app := fiber.New(fiber.Config{
		AppName:      cfg.App().Name(),
		BodyLimit:    cfg.App().BodyLimit(),
		ReadTimeout:  cfg.App().ReadTimeOut(),
		WriteTimeout: cfg.App().WriteTimeOut(),
		JSONEncoder:  json.Marshal,
		JSONDecoder:  json.Unmarshal,
	})
	/* Init Middleware */
	middlewareInf := middleware.InitMiddleware(cfg, authRepo)
	/* Setup Middleware */
	app.Use(middlewareInf.SetTracer())
	app.Use(middlewareInf.Cors())
	app.Use(middlewareInf.Logger())
	app.Use(middlewareInf.InputForm())
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})
	/* Swagger Route */
	app.Get("/swagger/*", swagger.HandlerDefault)

	/* Init Routing */
	router := app.Group("/v1")
	r := route.NewRoute(router)
	r.RegisterUser(userHand, userValidate)

	/* Graceful Shutdown */
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		ticker := <-c
		log.Println("Server is shutting down...")
		if ticker != nil {
			app.Shutdown()
		}
	}()

	/* Start Server */
	if err := app.Listen(cfg.App().Url()); err != nil {
		logrus.Fatal(err)
	}
}

func startGRPCServer(cfg config.Iconfig, server *grpc.Server) {
	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPC().Port()))
	if err != nil {
		panic("failed to listen: " + err.Error())
	}

	/* serve grpc */
	log.Printf("start grpc server [::%d]", cfg.GRPC().Port())
	if err := server.Serve(listen); err != nil {
		panic(err)
	}
}
