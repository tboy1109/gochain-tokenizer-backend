package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/treeder/firetils"
	"github.com/treeder/gotils/v2"
	"github.com/treeder/quickstart/globals"
)

type Organization struct {
	firetils.Firestored
	firetils.TimeStamped
	firetils.IDed
	Name        string `firestore:"name" json:"name"`
	Logo 	 	    string `firestore:"logo" json:"logo"`
	Admin				string `firestore:"admin" json:"admin"`
}

type Member struct {
	firetils.Firestored
	firetils.TimeStamped
	firetils.IDed
	Email        string `firestore:"email" json:"email"`
	Role	        string `firestore:"role" json:"role"`
	OrgID 	 	    string `firestore:"orgid" json:"orgid"`
}

type AddResult struct{
	ID	string	`firestore:"id" json:"id"`
}

type InviteReq struct {
	Email            string `json:"email"`
}

type LeaveReq struct {
	Email            string `json:"email"`
}

func addOrganization(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	email := r.FormValue("email")

	org := &Organization{}
	org.Name = r.FormValue("name")
	org.Admin = email

	member := &Member{}
	member.Email = email
	member.Role = "Admin"

	file, handler, err := r.FormFile("logo")
	if err != nil {
		return gotils.C(ctx).Errorf("Error Retrieving the File: %w", err)
	}
	fmt.Printf("Uploaded File: %+v\n", handler.Filename)
	fmt.Printf("File Size: %+v\n", handler.Size)
	fmt.Printf("MIME Header: %+v\n", handler.Header)
	id := uuid.New()

	fileName := fmt.Sprintf("%v", time.Now().Unix())

	wc := globals.App.Bucket.Object(fileName).NewWriter(context.Background())
	wc.ObjectAttrs.Metadata = map[string]string{"firebaseStorageDownloadTokens": id.String()}
	_, err = io.Copy(wc, file)
	if err != nil {
		return gotils.C(ctx).Errorf("Error uploading the File: %w", err)
	}

	if err := wc.Close(); err != nil {
		return gotils.C(ctx).Errorf("Error closing the File: %w", err)
	}
	// TODO: store this in our own db as we build it up
	json.Marshal(wc.ObjectAttrs.MediaLink)
	// url, err := globals.App.Bucket.SignedURL("tmp.bin", opts)
	// if err != nil {
	// 	return gotils.C(ctx).Errorf("Bucket.SignedURL: %v", err)
	// }
	org.Logo = "https://firebasestorage.googleapis.com/v0/b/tokenizer-dev-bae64.appspot.com/o/" + fileName + "?alt=media&token=" + wc.ObjectAttrs.Metadata["firebaseStorageDownloadTokens"]
	fmt.Printf("Img URL:%v\n", org.Logo)
	v, err := firetils.Save(ctx, globals.App.Db, "organizations", org)
	if err != nil {
		return gotils.C(ctx).Errorf("fs error: %w", err)
	}
	member.OrgID = v.GetID()
	firetils.Save(ctx, globals.App.Db, "members", member)
	gotils.WriteObject(w, http.StatusOK, map[string]interface{}{"organization": v})

	return nil
}

func getOrganization(w http.ResponseWriter, r *http.Request) error {
	paths := strings.Split(r.URL.Path, "/")
	orgID := paths[3]
	ctx := r.Context()

	//userID := firetils.UserID(ctx)
	fmt.Println("ORGID:", orgID)
	org := &Organization{}

	firetils.GetByID(ctx, globals.App.Db, "organizations", orgID, org)
	
	// if err != nil {
	// 	return gotils.C(ctx).Errorf("fs error: %w", err)
	// }

	// TODO: store this in our own db as we build it up
	gotils.WriteObject(w, http.StatusOK, map[string]interface{}{"organization": org})

	return nil
}

func getOrganizations(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	paths := strings.Split(r.URL.Path, "/")
	email := paths[4]

	fmt.Printf("User Id:%v\n", email)

	org := &Organization{}
	orgs, err := firetils.GetAllByQuery2(ctx, globals.App.Db.Collection("organizations").Offset(0), org)

	member := &Member{}
	members, err := firetils.GetAllByQuery2(ctx, globals.App.Db.Collection("members").Where("email","==",email), member)

	if err != nil {
		return gotils.C(ctx).Errorf("fs error: %w", err)
	}

	// TODO: store this in our own db as we build it up
	gotils.WriteObject(w, http.StatusOK, map[string]interface{}{"organizations": orgs, "members": members})

	return nil
}

func getOrgUsers(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	paths := strings.Split(r.URL.Path, "/")
	orgID := paths[3]

	fmt.Printf("ORG Id:%v\n", orgID)

	member := &Member{}

	users, err := firetils.GetAllByQuery2(ctx, globals.App.Db.Collection("members").Where("orgid", "==", orgID), member)
	if err != nil {
		return gotils.C(ctx).Errorf("fs error: %w", err)
	}

	// TODO: store this in our own db as we build it up
	gotils.WriteObject(w, http.StatusOK, map[string]interface{}{"users": users})

	return nil
}

func getAdminOrgs(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	paths := strings.Split(r.URL.Path, "/")
	email := paths[4]

	fmt.Printf("User Id:%v\n", email)

	org := &Organization{}
	orgs, err := firetils.GetAllByQuery2(ctx, globals.App.Db.Collection("organizations").Where("admin","==",email), org)

	if err != nil {
		return gotils.C(ctx).Errorf("fs error: %w", err)
	}

	// TODO: store this in our own db as we build it up
	gotils.WriteObject(w, http.StatusOK, map[string]interface{}{"organizations": orgs})

	return nil
}

func inviteUser(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	paths := strings.Split(r.URL.Path, "/")
	orgID := paths[3]

	inviteReq := &InviteReq{}
	err := gotils.ParseJSONReader(r.Body, inviteReq)
	if err != nil {
		return gotils.C(ctx).Errorf("bad input: %w", err)
	}

	member := &Member{}
	member.Email = inviteReq.Email
	member.Role = "User"
	member.OrgID = orgID
	firetils.Save(ctx, globals.App.Db, "members", member)
	gotils.WriteObject(w, http.StatusOK, map[string]interface{}{"organization": member})

	return nil
}

func leaveUser(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	paths := strings.Split(r.URL.Path, "/")
	orgID := paths[3]

	member := &Member{}

	leaveReq := &LeaveReq{}
	err := gotils.ParseJSONReader(r.Body, leaveReq)
	if err != nil {
		return gotils.C(ctx).Errorf("bad input: %w", err)
	}

	firetils.GetOneByQuery(ctx, globals.App.Db.Collection("members").Where("orgid", "==", orgID).Where("email","==",leaveReq.Email), member)

	fmt.Printf("Searched ID:%v",member.IDed)

	firetils.Delete(ctx, globals.App.Db, "members", member.GetID())

	gotils.WriteObject(w, http.StatusOK, map[string]interface{}{"status": "success"})

	return nil
}