package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/treeder/firetils"
	"github.com/treeder/gotils/v2"
	"github.com/treeder/quickstart/globals"
)

type Assets struct {
	firetils.Firestored
	firetils.TimeStamped
	firetils.IDed
	Name        string 		`firestore:"name" json:"name"`
	Description string 		`firestore:"description" json:"description"`
	Equity      int    		`firestore:"equity" json:"equity"`
	Seeking     int    		`firestore:"seeking" json:"seeking"`
	Location    string 		`firestore:"location" json:"location"`
	Category    string 		`firestore:"category" json:"category"`
	Valuation   int    		`firestore:"valuation" json:"valuation"`
	SharePrice  int    		`firestore:"sharePrice" json:"sharePrice"`
	Creator     string 		`firestore:"creator" json:"creator"`
	ImgURL      string 		`firestore:"imgUrl" json:"imgUrl"`
	Map      		string 		`firestore:"map" json:"map"`
	TokenId	 	  int	   		`firestore:"tokenId" json:"tokenId"`
	Owner				string		`firestore:"owner" json:"owner"`
	FieldNames	[]string 	`firestore:"fieldNames" json:"fieldNames"`
	Values			[]string 	`firestore:"values" json:"values"`
}

type ImgData struct {
	ImgData string `firestore:"imgData" json:"imgData"`
}

type IPFSResp struct {
	IpfsHash    string `json:"IpfsHash"`
	PinSize     string `json:"PinSize"`
	TimeStamp   string `json:"Timestamp"`
	IsDuplicate string `json:"isDuplicate"`
}

type TokenizeReq struct {
	Id            string `json:"id"`
	WalletAddress string `json:"walletAddress"`
}

type CompleteTokenizeReq struct {
	Id 			string 	`json:"id"`
	TokenId	int			`json:"tokenId"`
}

type Attribute struct {
	TraitType string      `json:"trait_type"`
	Value     interface{} `json:"value"`
}

type Collection struct {
	Name   string `json:"name"`
	Family string `json:"family"`
}

type File struct {
	URI  string `json:"uri"`
	Type string `json:"type"`
}

type Creator struct {
	Address string `json:"address"`
	Share   int    `json:"share"`
}

type Properties struct {
	Files    []File    `json:"files"`
	Category string    `json:"category"`
	Creators []Creator `json:"creators"`
}

type Metadata struct {
	Name                 string      `json:"name"`
	Edition              int         `json:"edition"`
	Description          string      `json:"description"`
	SellerFeeBasicPoints int         `json:"seller_fee_basis_points"`
	Image                string      `json:"image"`
	ExternalURL          string      `json:"external_url"`
	Attributes           []Attribute `json:"attributes"`
	Collection           Collection  `json:"collection"`
	Date                 int64       `json:"date"`
	Properties           Properties  `json:"properties"`
	Symbol               string      `json:"symbol"`
}

func addAssets(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	mi := &Assets{}
	mi.Name = r.FormValue("name")
	mi.Description = r.FormValue("description")
	equity, err := strconv.ParseInt(r.FormValue("equity"), 10, 32)
	if err != nil {
		return gotils.C(ctx).Errorf("Parsing Error: %w", err)
	}
	mi.Equity = int(equity)
	seeking, err := strconv.ParseInt(r.FormValue("seeking"), 10, 32)
	if err != nil {
		return gotils.C(ctx).Errorf("Parsing Error: %w", err)
	}
	mi.Seeking = int(seeking)
	mi.Location = r.FormValue("location")
	mi.Category = r.FormValue("category")
	valuation, err := strconv.ParseInt(r.FormValue("valuation"), 10, 32)
	if err != nil {
		return gotils.C(ctx).Errorf("Parsing Error: %w", err)
	}
	mi.Valuation = int(valuation)
	sharePrice, err := strconv.ParseInt(r.FormValue("sharePrice"), 10, 32)
	if err != nil {
		return gotils.C(ctx).Errorf("Parsing Error: %w", err)
	}
	mi.SharePrice = int(sharePrice)
	mi.Creator = r.FormValue("creator")
	mi.Owner = r.FormValue("owner")
	fmt.Printf("Field Names:%v\n", r.Form["fieldNames[]"])
	fmt.Printf("Values:%v\n", r.Form["values[]"])
	mi.FieldNames = r.Form["fieldNames[]"]
	mi.Values = r.Form["values[]"]

	file, handler, err := r.FormFile("imgData")
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
	mi.ImgURL = "https://firebasestorage.googleapis.com/v0/b/tokenizer-dev-bae64.appspot.com/o/" + fileName + "?alt=media&token=" + wc.ObjectAttrs.Metadata["firebaseStorageDownloadTokens"]

	file, handler, err = r.FormFile("mapData")
	if (err != nil) {
		fmt.Printf("Uploaded File: %+v\n", handler.Filename)
		fmt.Printf("File Size: %+v\n", handler.Size)
		fmt.Printf("MIME Header: %+v\n", handler.Header)
		id = uuid.New()

		fileName = fmt.Sprintf("%v", time.Now().Unix())

		wc = globals.App.Bucket.Object(fileName).NewWriter(context.Background())
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
		mi.Map = "https://firebasestorage.googleapis.com/v0/b/tokenizer-dev-bae64.appspot.com/o/" + fileName + "?alt=media&token=" + wc.ObjectAttrs.Metadata["firebaseStorageDownloadTokens"]

		fmt.Printf("Img URL:%v\n", mi.Map)
	}
	v, err := firetils.Save(ctx, globals.App.Db, "assets", mi)
	if err != nil {
		return gotils.C(ctx).Errorf("fs error: %w", err)
	}
	gotils.WriteObject(w, http.StatusOK, map[string]interface{}{"asset": v})

	return nil
}

func getAssets(w http.ResponseWriter, r *http.Request) error {
	paths := strings.Split(r.URL.Path, "/")
	userID := paths[3]
	ctx := r.Context()

	//userID := firetils.UserID(ctx)
	fmt.Println("USERID:", userID)
	mi := &Assets{}

	vs, err := firetils.GetAllByQuery2(ctx, globals.App.Db.Collection("assets").Where("creator", "==", userID), mi)
	if err != nil {
		return gotils.C(ctx).Errorf("fs error: %w", err)
	}

	// TODO: store this in our own db as we build it up
	gotils.WriteObject(w, http.StatusOK, map[string]interface{}{"assets": vs})

	return nil
}

func tokenize(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	tokenizeReq := &TokenizeReq{}
	err := gotils.ParseJSONReader(r.Body, tokenizeReq)
	if err != nil {
		return gotils.C(ctx).Errorf("bad input: %w", err)
	}

	asset := &Assets{}
	firetils.GetByID(ctx, globals.App.Db, "assets", tokenizeReq.Id, asset)

	ipfsApiUrl := "https://api.pinata.cloud/pinning/pinFileToIPFS"
	pinataApiKey := "160672bbe8ae4df35b3a"
	pinataSecretKey := "6ef08e9c58c56efd156f56b4934a4c64bf8b702eb7989e562eca88f15222a78c"

	var b bytes.Buffer
	writer := multipart.NewWriter(&b)
	fw, err := writer.CreateFormFile("file", "newNft.png")
	if err != nil {
		return gotils.C(ctx).Errorf("Error creating form file: %w", err)
	}
	imgresp, err := http.Get(asset.ImgURL)
	if err != nil {
		return gotils.C(ctx).Errorf("Error getting Aset Image: %w", err)
	}
	defer imgresp.Body.Close()
	imgbody, err := ioutil.ReadAll(imgresp.Body)
	if err != nil {
		return gotils.C(ctx).Errorf("Err getting image response: %w", err)
	}
	_, err = fw.Write(imgbody)
	if err != nil {
		return gotils.C(ctx).Errorf("Err writing image: %w", err)
	}
	writer.Close()

	req, err := http.NewRequest("POST", ipfsApiUrl, &b)
	if err != nil {
		return gotils.C(ctx).Errorf("Error making ipfs request: %w", err)
	}
	// Don't forget to set the content type, this will contain the boundary.
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("pinata_api_key", pinataApiKey)
	req.Header.Set("pinata_secret_api_key", pinataSecretKey)

	// Submit the request
	hc := http.Client{}
	resp, err := hc.Do(req)
	if err != nil {
		return gotils.C(ctx).Errorf("Error calling API: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Failed calling API %v\n", resp)
		return gotils.C(ctx).Errorf("Failed calling API: %v", resp)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body) // response body is []byte
	if err != nil {
		return gotils.C(ctx).Errorf("Error reading API Response: %w", err)
	}
	var ipfsResp IPFSResp
	json.Unmarshal(body, &ipfsResp)
	fmt.Println("Image Response Hash:", ipfsResp.IpfsHash)

	metadata := &Metadata{}
	metadata.Name = asset.Name
	metadata.Description = asset.Description
	metadata.Edition = 1
	metadata.SellerFeeBasicPoints = 0
	metadata.Image = "https://ipfs.io/ipfs/" + ipfsResp.IpfsHash
	metadata.ExternalURL = "External URL"
	metadata.Attributes = []Attribute{
		{
			TraitType: "Name",
			Value:     asset.Name,
		},
		{
			TraitType: "Description",
			Value:     asset.Description,
		},
		{
			TraitType: "Equity",
			Value:     asset.Equity,
		},
		{
			TraitType: "Seeking",
			Value:     asset.Seeking,
		},
		{
			TraitType: "Location",
			Value:     asset.Location,
		},
		{
			TraitType: "Category",
			Value:     asset.Category,
		},
		{
			TraitType: "Valuation",
			Value:     asset.Valuation,
		},
		{
			TraitType: "SharePrice",
			Value:     asset.SharePrice,
		},
		{
			TraitType: "Creator",
			Value:     tokenizeReq.WalletAddress,
		},
		{
			TraitType: "ImgURL",
			Value:     asset.ImgURL,
		},
	}
	for i, fieldName := range asset.FieldNames {
		fmt.Printf("%v %v\n",asset.Values[i],fieldName)
    metadata.Attributes = append(metadata.Attributes, Attribute{
			TraitType: fieldName,
			Value: asset.Values[i],
		})
	}
	metadata.Collection.Name = "Tokenized NFT"
	metadata.Collection.Family = "Tokenized NFT"
	metadata.Date = time.Now().Unix()
	metadata.Properties.Files = []File{
		{
			URI:  "https://ipfs.io/ipfs/" + ipfsResp.IpfsHash,
			Type: "Image",
		},
	}
	metadata.Properties.Category = "Asset"
	metadata.Properties.Creators = []Creator{
		{
			Address: tokenizeReq.WalletAddress,
			Share:   100,
		},
	}
	metadata.Symbol = "Tokenized NFT"
	fmt.Printf("%v\n", metadata)

	//start
	var mb bytes.Buffer
	writer = multipart.NewWriter(&mb)
	fw, err = writer.CreateFormFile("file", "metadata.json")
	if err != nil {
		return gotils.C(ctx).Errorf("Error creating form file: %w", err)
	}
	metadt, err := json.Marshal(metadata)
	if err != nil {
		return gotils.C(ctx).Errorf("Error marshalling metadata: %w", err)
	}
	_, err = fw.Write(metadt)
	if err != nil {
		return gotils.C(ctx).Errorf("Err writing metadata: %w", err)
	}
	writer.Close()

	req, err = http.NewRequest("POST", ipfsApiUrl, &mb)
	if err != nil {
		return gotils.C(ctx).Errorf("Error making ipfs request: %w", err)
	}
	// Don't forget to set the content type, this will contain the boundary.
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("pinata_api_key", pinataApiKey)
	req.Header.Set("pinata_secret_api_key", pinataSecretKey)

	// Submit the request
	hc = http.Client{}
	resp, err = hc.Do(req)
	if err != nil {
		return gotils.C(ctx).Errorf("Error calling API: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Failed calling API %v\n", resp)
		return gotils.C(ctx).Errorf("Failed calling API: %v", resp)
	}
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body) // response body is []byte
	if err != nil {
		return gotils.C(ctx).Errorf("Error reading API Response: %w", err)
	}
	json.Unmarshal(body, &ipfsResp)
	fmt.Println("Metadata Response Hash:", ipfsResp.IpfsHash)

	gotils.WriteObject(w, http.StatusOK, map[string]interface{}{"metadataURL": "https://ipfs.io/ipfs/" + ipfsResp.IpfsHash})
	return nil
}

func completeTokenization(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	completeTokenizeReq := &CompleteTokenizeReq{}
	err := gotils.ParseJSONReader(r.Body, completeTokenizeReq)
	if err != nil {
		return gotils.C(ctx).Errorf("bad input: %w", err)
	}

	asset := &Assets{}
	firetils.GetByID(ctx, globals.App.Db, "assets", completeTokenizeReq.Id, asset)

	asset.TokenId = completeTokenizeReq.TokenId
	firetils.Save(ctx, globals.App.Db, "assets", asset)

	gotils.WriteObject(w, http.StatusOK, map[string]interface{}{"id": completeTokenizeReq.Id})

	return nil
}

func getOrganizationAssets(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	paths := strings.Split(r.URL.Path, "/")
	orgID := paths[3]

	fmt.Printf("orgID Id:%v\n", orgID)

	org := &Assets{}
	orgs, err := firetils.GetAllByQuery2(ctx, globals.App.Db.Collection("assets").Where("owner","==",orgID), org)

	if err != nil {
		return gotils.C(ctx).Errorf("fs error: %w", err)
	}

	// TODO: store this in our own db as we build it up
	gotils.WriteObject(w, http.StatusOK, map[string]interface{}{"assets": orgs})

	return nil
}

func getAsset(w http.ResponseWriter, r *http.Request) error {
  paths := strings.Split(r.URL.Path, "/")
  assetID := paths[4]
  ctx := r.Context()

  //userID := firetils.UserID(ctx)
  fmt.Println("AsseID:", assetID)
  org := &Assets{}

  firetils.GetByID(ctx, globals.App.Db, "assets", assetID, org)
  
  // if err != nil {
  //  return gotils.C(ctx).Errorf("fs error: %w", err)
  // }

  // TODO: store this in our own db as we build it up
  gotils.WriteObject(w, http.StatusOK, map[string]interface{}{"asset": org})

  return nil
}