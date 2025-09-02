package bedrock

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-xray-sdk-go/xray"
)

func InitializeXRay() {
	environment := getEnv("ENVIRONMENT", "development")
	useXRay := getEnv("USE_XRAY_SDK", "false")
	
	if environment == "production" && useXRay == "true" {
		fmt.Println("Initializing AWS X-Ray SDK...")
		
		// Configure X-Ray to use CloudWatch Agent
		err := xray.Configure(xray.Config{
			ServiceVersion:    "1.0.0",
			DaemonAddr:       "cloudwatch-agent.amazon-cloudwatch.svc.cluster.local:2000",
		})
		if err != nil {
			fmt.Printf("Failed to configure X-Ray: %v\n", err)
			return
		}
		
		fmt.Println("AWS X-Ray SDK configured with CloudWatch Agent")
		fmt.Println("AWS X-Ray SDK initialized successfully")
	} else {
		fmt.Println("X-Ray disabled for development environment")
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// XRayMiddleware wraps HTTP handlers with X-Ray tracing
func XRayMiddleware(name string, handler func(context.Context) error) func(context.Context) error {
	return func(ctx context.Context) error {
		if getEnv("USE_XRAY_SDK", "false") == "true" {
			ctx, seg := xray.BeginSegment(ctx, name)
			defer seg.Close(nil)
			
			err := handler(ctx)
			if err != nil {
				seg.AddError(err)
			}
			return err
		}
		return handler(ctx)
	}
}
