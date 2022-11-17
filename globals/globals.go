package globals

import (
	"cloud.google.com/go/firestore"
	googleStorage "cloud.google.com/go/storage"
	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	storage "firebase.google.com/go/v4/storage"
	"github.com/dgraph-io/ristretto"
)

func init() {
	App = &MyApp{}
}

var (
	App *MyApp
)

type MyApp struct {
	FireApp       *firebase.App
	Auth          *auth.Client
	Cache         *ristretto.Cache
	Db            *firestore.Client
	StorageClient *storage.Client
	Bucket        *googleStorage.BucketHandle
}
