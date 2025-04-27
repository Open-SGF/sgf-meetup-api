package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"sgf-meetup-api/pkg/infra"
	"sgf-meetup-api/pkg/upsertuser"
	"strings"
	"time"
)

const (
	serviceInitTimeout = 10 * time.Second
	upsertTimeout      = 5 * time.Second
)

func main() {
	var clientID, clientSecret string

	flag.StringVar(&clientID, "clientId", "", "Client ID for the user (required)")
	flag.StringVar(&clientSecret, "clientSecret", "", "Client Secret for the user (required)")
	flag.Usage = createUsageFunc()
	flag.Parse()

	validateRequiredFlags(clientID, clientSecret)

	ctx, cancel := context.WithTimeout(context.Background(), serviceInitTimeout)
	defer cancel()

	service, err := upsertuser.InitService(ctx)
	if err != nil {
		log.Fatalf("service initialization failed: %v", err)
	}

	upsertCtx, upsertCancel := context.WithTimeout(context.Background(), upsertTimeout)
	defer upsertCancel()

	if err := service.UpsertUser(upsertCtx, *infra.ApiUsersTableProps.TableName, clientID, clientSecret); err != nil {
		log.Fatalf("user upsert operation failed: %v", err)
	}

	log.Printf("successfully upserted user %q", clientID)
}

func createUsageFunc() func() {
	return func() {
		_, _ = fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s -clientId <ID> -clientSecret <SECRET>\n\n", os.Args[0])
		_, _ = fmt.Fprintln(flag.CommandLine.Output(), "Required flags:")
		flag.VisitAll(func(f *flag.Flag) {
			_, _ = fmt.Fprintf(flag.CommandLine.Output(), "  -%-12s%s\n", f.Name, f.Usage)
		})
	}
}

func validateRequiredFlags(flags ...string) {
	var missing []string
	for i := 0; i < len(flags); i += 2 {
		if flags[i] == "" {
			missing = append(missing, strings.TrimPrefix(flag.CommandLine.Name(), "-"))
		}
	}

	if len(missing) > 0 {
		_, _ = fmt.Fprintf(os.Stderr, "missing required parameters: %s\n\n", strings.Join(missing, ", "))
		flag.Usage()
		os.Exit(1)
	}
}
