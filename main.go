package main

import (
	"context"
	"log"

	"github.com/dgraph-io/ristretto"
	"github.com/joho/godotenv"
	"github.com/treeder/firetils"
	"github.com/treeder/gcputils"
	"github.com/treeder/goapibase"
	"github.com/treeder/gotils/v2"
	"github.com/treeder/quickstart/globals"
)

func main() {
	ctx := context.Background()

	err := godotenv.Load()
	if err != nil {
		gotils.L(ctx).Info().Println("Warning: error loading .env file:", err)
	}

	// GOOGLE CREDS
	opts, projectID, err := gcputils.CredentialsAndProjectIDFromEnv("G_KEY", "G_PROJECT_ID")
	if err != nil {
		log.Fatalln(err)
	}

	// GET OTHER ENV VARS HERE
	env := gcputils.GetEnvVar("ENV", "dev")
	if env == "prod" {
		gotils.SetLoggable(gcputils.NewLogger())
	}

	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 10000000,
		MaxCost:     100000000,
		BufferItems: 64,
	})
	if err != nil {
		log.Fatalln("error creating cache", err)
	}
	globals.App.Cache = cache

	firebaseApp, err := firetils.New(ctx, projectID, opts)
	if err != nil {
		gotils.Logf(ctx, "error", "couldn't init firebase newapp: %v\n", err)
		return
	}
	globals.App.FireApp = firebaseApp
	firestore, err := firebaseApp.Firestore(ctx)
	if err != nil {
		gotils.Logf(ctx, "error", "couldn't init firestore: %v\n", err)
		return
	}
	globals.App.Db = firestore
	globals.App.Auth, err = firebaseApp.Auth(ctx)
	if err != nil {
		gotils.L(ctx).Error().Printf("error getting firebase auth client: %v\n", err)
		return
	}

	client, err := firebaseApp.Storage(context.Background())
	if err != nil {
		gotils.L(ctx).Error().Printf("error getting firebase auth client: %v\n", err)
		return
	}
	globals.App.StorageClient = client
	/*
		bucket, err := client.DefaultBucket()
		if err != nil {
			gotils.L(ctx).Error().Printf("error getting firebase bucket: %v\n", err)
			return
		}
		globals.App.Bucket = bucket
	*/
	bucket := "tokenizer-dev-bae64.appspot.com"
	globals.App.Bucket, err = client.Bucket(bucket)
	if err != nil {
		gotils.Logf(ctx, "error", "couldn't create firestore bucket: %v\n", err)
	}

	// add something to firestore just to be sure it's working
	tmp := firestore.Collection("tmp")
	_, _, err = tmp.Add(ctx, TmpType{Name: "wall-e"})
	if err != nil {
		gotils.Logf(ctx, "error", "couldn't write to firestore: %v\n", err)
	}

	r := goapibase.InitRouter(ctx)
	// Setup your routes
	setupRoutes(ctx, r)
	// Start server
	_ = goapibase.Start(ctx, gotils.Port(8080), r)
}

type TmpType struct {
	Name string `firestore:"name"`
}
