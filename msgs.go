package main

import (
	"fmt"
	"net/http"

	"cloud.google.com/go/firestore"
	"github.com/treeder/firetils"
	"github.com/treeder/gotils/v2"
	"github.com/treeder/quickstart/globals"
)

type Msg struct {
	firetils.Firestored
	firetils.TimeStamped
	firetils.IDed
	Msg string `firestore:"msg" json:"msg"`
}

func postMsg(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	mi := &Msg{}
	err := gotils.ParseJSONReader(r.Body, mi)
	if err != nil {
		return gotils.C(ctx).Errorf("bad input: %w", err)
	}

	v, err := firetils.Save(ctx, globals.App.Db, "msgs", mi)
	if err != nil {
		return gotils.C(ctx).Errorf("fs error: %w", err)
	}

	// TODO: store this in our own db as we build it up
	gotils.WriteObject(w, http.StatusOK, map[string]interface{}{"message": v})

	return nil
}

func getMsgs(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	userID := firetils.UserID(ctx)
	fmt.Println("USERID:", userID)
	mi := &Msg{}

	vs, err := firetils.GetAllByQuery2(ctx, globals.App.Db.Collection("msgs").OrderBy("createdAt", firestore.Desc), mi)
	if err != nil {
		return gotils.C(ctx).Errorf("fs error: %w", err)
	}

	// TODO: store this in our own db as we build it up
	gotils.WriteObject(w, http.StatusOK, map[string]interface{}{"messages": vs})

	return nil
}
