package talipapops

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/storage"
	"github.com/nfnt/resize"
	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"
	"google.golang.org/api/iterator"
)

// Ret is the return expected for API
// type Ret struct {
// 	Status  string          `json:"status"`
// 	Message string          `json:"message"`
// 	Data    json.RawMessage `json:"data"`
// }

// var (
// 	storageClient *storage.Client
// )

// const projectID = "talipapops-wapak"
// const bucketName = "files.talipapops.ml"

// Item is the list al documents
type Item struct {
	ItemID     string          `json:"item_id"`
	ItemName   string          `json:"item_name"`
	ItemImages json.RawMessage `json:"images"`
	Condition  string          `json:"condition"`
	Owner      string          `json:"owner"`
	TradeWith  json.RawMessage `json:"trade_with"`
}

// AddItem provides updating of avatar and details
func AddItem(w http.ResponseWriter, r *http.Request) {

	ctx := context.Background()
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization")
		w.Header().Set("Access-Control-Max-Age", "3600")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	// Set CORS headers for the main request.
	w.Header().Set("Access-Control-Allow-Origin", "*")

	authorization := r.Header.Get("Authorization")

	_, UserInfo := validatePaseto(authorization)

	splitInfo := strings.Split(UserInfo.Audience, "||")

	returnThis := &Ret{Status: "fail", Message: "Everything is Required"}

	r.ParseMultipartForm(0)

	iLength, err := strconv.Atoi(r.FormValue("imgCount"))
	if err != nil {
		fmt.Println(err)
	}

	var imageFiles [5]string

	for i := 0; i < iLength; i++ {

		fmt.Println("IM SIMILA")

		imgconcatenated := fmt.Sprint("imgFiles", i)

		imgProps := strings.Split(r.FormValue(imgconcatenated), ",")
		imgTypes := strings.Split(imgProps[0], ";")
		imgTypes2 := strings.Split(imgTypes[0], ":")

		imgData := imgProps[1]
		imgMimeType := imgTypes2[1]
		imgTypes3 := strings.Split(imgMimeType, "/")

		imgExtension := imgTypes3[1]
		fmt.Println("IM EXTENSION", imgExtension)

		reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(imgData))

		image, _, err := image.Decode(reader)
		if err != nil {
			log.Fatal(err)
		}

		newImage := resize.Resize(1024, 0, image, resize.Lanczos3)

		client, err := storage.NewClient(ctx)
		if err != nil {
			fmt.Println(err)
		}

		bucket := client.Bucket(bucketName)

		name := splitInfo[1] + "/items/" + strconv.FormatInt(time.Now().Unix(), 10) + "." + imgExtension

		w := bucket.Object(name).NewWriter(ctx)

		w.ACL = []storage.ACLRule{{Entity: storage.AllUsers, Role: storage.RoleReader}}
		w.CacheControl = "public, max-age=86400"

		// if _, err := io.Copy(w, reader); err != nil {
		// 	fmt.Println(err)
		// }

		switch imgExtension {
		case "png":
			err = png.Encode(w, newImage)
		case "gif":
			err = gif.Encode(w, newImage, nil)
		case "jpeg", "jpg":
			err = jpeg.Encode(w, newImage, nil)
		case "bmp":
			err = bmp.Encode(w, newImage)
		case "tiff":
			err = tiff.Encode(w, newImage, nil)
		default:
		}

		if err := w.Close(); err != nil {
			fmt.Println(err)
		}

		fmt.Println("IM FILE", "https://storage.googleapis.com/"+bucketName+"/"+name)

		imageFiles[i] = "https://storage.googleapis.com/" + bucketName + "/" + name
	}

	tradeWith := strings.Split(r.FormValue("tradeWith"), ",")

	if r.FormValue("itemName") != "" && r.FormValue("condition") != "" && r.FormValue("imgCount") != "" && r.FormValue("tradeWith") != "" {

		client, _ := firestore.NewClient(ctx, projectID)

		_, _, err := client.Collection("items").Add(ctx, map[string]interface{}{
			"name":       r.FormValue("itemName"),
			"condition":  r.FormValue("condition"),
			"tradewith":  tradeWith,
			"images":     imageFiles,
			"created_at": time.Now(),
			"owner":      splitInfo[1],
			"status":     1,
		})

		if err != nil {

			returnThis = &Ret{Status: "fail", Message: "Cannot Add Item"}

			e, _ := json.Marshal(returnThis)
			w.Write(e)
			return
		} else {
			returnData := map[string]interface{}{
				"id":    splitInfo[1],
				"email": splitInfo[0],
			}
			jsonString, _ := json.Marshal(returnData)
			returnThis = &Ret{Status: "success", Message: "Item Added Successfully", Data: jsonString}
			e, _ := json.Marshal(returnThis)
			w.Write(e)
			return
		}

	} else {
		returnThis = &Ret{Status: "fail", Message: "Cannot Update Item"}
		e, _ := json.Marshal(returnThis)
		w.Write(e)
		return
	}

}

// CountItems I owned
func CountItems(w http.ResponseWriter, r *http.Request) {

	ctx := context.Background()
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization")
		w.Header().Set("Access-Control-Max-Age", "3600")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	// Set CORS headers for the main request.
	w.Header().Set("Access-Control-Allow-Origin", "*")

	authorization := r.Header.Get("Authorization")

	_, UserInfo := validatePaseto(authorization)

	splitInfo := strings.Split(UserInfo.Audience, "||")

	returnThis := &Ret{Status: "fail", Message: "Everything is Required"}

	client, _ := firestore.NewClient(ctx, projectID)

	iter := client.Collection("items").Where("owner", "==", splitInfo[1]).Documents(ctx)

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			returnThis = &Ret{Status: "fail", Message: "Cannot Add Item"}

			e, _ := json.Marshal(returnThis)
			w.Write(e)
			return
		}
		if doc.Data() == nil {

			fmt.Println("DOC NLL", doc.Data())
			returnThis = &Ret{Status: "fail", Message: "No items"}
			e, _ := json.Marshal(returnThis)
			w.Write(e)
			return
		}

		if doc.Data() != nil {

			fmt.Println(doc.Data())

			jsonString, _ := json.Marshal(doc.Data())
			returnThis = &Ret{Status: "success", Message: "Welcome Back ", Data: jsonString}
			e, _ := json.Marshal(returnThis)
			w.Write(e)
			return

		}
	}

}

// NewItems list all new items
func NewItems(w http.ResponseWriter, r *http.Request) {

	items := []Item{}

	ctx := context.Background()
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization")
		w.Header().Set("Access-Control-Max-Age", "3600")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	// Set CORS headers for the main request.
	w.Header().Set("Access-Control-Allow-Origin", "*")

	authorization := r.Header.Get("Authorization")

	_, UserInfo := validatePaseto(authorization)

	splitInfo := strings.Split(UserInfo.Audience, "||")

	fmt.Print(splitInfo)

	returnThis := &Ret{Status: "fail", Message: "Everything is Required"}

	client, _ := firestore.NewClient(ctx, projectID)

	firstPage := client.Collection("items").OrderBy("created_at", firestore.Desc).Documents(ctx)

	docs, err := firstPage.GetAll()

	fmt.Println("DOCSSS", docs)
	if err != nil {
		returnThis = &Ret{Status: "fail", Message: err.Error()}
		e, _ := json.Marshal(returnThis)
		w.Write(e)
		return
	}

	// Get the last document.
	lastDoc := docs[len(docs)-1]
	s := fmt.Sprintf("%v", lastDoc.Data()["created_at"])

	itemCollections := client.CollectionGroup("items").OrderBy("created_at", firestore.Desc).Documents(ctx)
	for {
		doc, err := itemCollections.Next()

		if err == iterator.Done {
			break
		}

		if err != nil {
			returnThis = &Ret{Status: "fail", Message: err.Error()}
			e, _ := json.Marshal(returnThis)
			w.Write(e)
			return
		}

		if err == nil {
			fmt.Println(doc.Data())

			name := fmt.Sprintf("%v", doc.Data()["name"])
			images, _ := json.Marshal(doc.Data()["images"])
			condition := fmt.Sprintf("%v", doc.Data()["condition"])
			owner := fmt.Sprintf("%v", doc.Data()["owner"])
			tradeWith, _ := json.Marshal(doc.Data()["tradewith"])

			its := []Item{
				{
					ItemID:     doc.Ref.ID,
					ItemName:   name,
					ItemImages: images,
					Condition:  condition,
					Owner:      owner,
					TradeWith:  tradeWith,
				},
			}

			items = append(items, its...)
			fmt.Println(items)
		}

	}

	jsonString, _ := json.Marshal(items)

	returnThis = &Ret{Status: "success", Message: s, Data: jsonString}
	e, _ := json.Marshal(returnThis)
	w.Write(e)
	return

}

// SearchItems using querystring
func SearchItems(w http.ResponseWriter, r *http.Request) {

	querystring, _ := r.URL.Query()["q"]

	fmt.Println("QQQ", querystring[0])
	items := []Item{}

	ctx := context.Background()
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization")
		w.Header().Set("Access-Control-Max-Age", "3600")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	// Set CORS headers for the main request.
	w.Header().Set("Access-Control-Allow-Origin", "*")

	authorization := r.Header.Get("Authorization")

	_, UserInfo := validatePaseto(authorization)

	splitInfo := strings.Split(UserInfo.Audience, "||")

	fmt.Print(splitInfo)

	returnThis := &Ret{Status: "fail", Message: "Everything is Required"}

	client, _ := firestore.NewClient(ctx, projectID)

	// itemCollections := client.CollectionGroup("items").Where("name", ">=", querystring[0]).Where("name", "<=", querystring[0]+"~").Documents(ctx)
	itemCollections := client.CollectionGroup("items").OrderBy("name", firestore.Asc).StartAt(querystring[0]).EndAt(querystring[0] + "\uf8ff").Documents(ctx)
	for {
		doc, err := itemCollections.Next()

		if err == iterator.Done {
			break
		}

		if err != nil {
			returnThis = &Ret{Status: "fail", Message: err.Error()}
			e, _ := json.Marshal(returnThis)
			w.Write(e)
			return
		}

		if err == nil {
			fmt.Println(doc.Data())

			name := fmt.Sprintf("%v", doc.Data()["name"])
			images, _ := json.Marshal(doc.Data()["images"])
			condition := fmt.Sprintf("%v", doc.Data()["condition"])
			owner := fmt.Sprintf("%v", doc.Data()["owner"])
			tradeWith, _ := json.Marshal(doc.Data()["tradewith"])

			its := []Item{
				{
					ItemID:     doc.Ref.ID,
					ItemName:   name,
					ItemImages: images,
					Condition:  condition,
					Owner:      owner,
					TradeWith:  tradeWith,
				},
			}

			items = append(items, its...)
			fmt.Println(items)
		}

	}

	jsonString, _ := json.Marshal(items)

	returnThis = &Ret{Status: "success", Message: "Search Result", Data: jsonString}
	e, _ := json.Marshal(returnThis)
	w.Write(e)
	return

}

// MyItems I owned
func MyItems(w http.ResponseWriter, r *http.Request) {

	items := []Item{}

	ctx := context.Background()
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization")
		w.Header().Set("Access-Control-Max-Age", "3600")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	// Set CORS headers for the main request.
	w.Header().Set("Access-Control-Allow-Origin", "*")

	authorization := r.Header.Get("Authorization")

	_, UserInfo := validatePaseto(authorization)

	splitInfo := strings.Split(UserInfo.Audience, "||")

	fmt.Print(splitInfo)

	returnThis := &Ret{Status: "fail", Message: "Everything is Required"}

	client, _ := firestore.NewClient(ctx, projectID)

	firstPage := client.Collection("items").OrderBy("created_at", firestore.Desc).Documents(ctx)

	docs, err := firstPage.GetAll()

	fmt.Println("DOCSSS", docs)
	if err != nil {
		returnThis = &Ret{Status: "fail", Message: err.Error()}
		e, _ := json.Marshal(returnThis)
		w.Write(e)
		return
	}

	// Get the last document.
	lastDoc := docs[len(docs)-1]
	s := fmt.Sprintf("%v", lastDoc.Data()["created_at"])

	itemCollections := client.Collection("items").Where("owner", "==", splitInfo[1]).Documents(ctx)
	for {
		doc, err := itemCollections.Next()

		if err == iterator.Done {
			break
		}

		if err != nil {
			returnThis = &Ret{Status: "fail", Message: err.Error()}
			e, _ := json.Marshal(returnThis)
			w.Write(e)
			return
		}

		if err == nil {
			fmt.Println(doc.Data())

			name := fmt.Sprintf("%v", doc.Data()["name"])
			images, _ := json.Marshal(doc.Data()["images"])
			condition := fmt.Sprintf("%v", doc.Data()["condition"])
			owner := fmt.Sprintf("%v", doc.Data()["owner"])
			tradeWith, _ := json.Marshal(doc.Data()["tradewith"])

			its := []Item{
				{
					ItemID:     doc.Ref.ID,
					ItemName:   name,
					ItemImages: images,
					Condition:  condition,
					Owner:      owner,
					TradeWith:  tradeWith,
				},
			}

			items = append(items, its...)
			fmt.Println(items)
		}

	}

	jsonString, _ := json.Marshal(items)

	returnThis = &Ret{Status: "success", Message: s, Data: jsonString}
	e, _ := json.Marshal(returnThis)
	w.Write(e)
	return

}

// ViewItems show specific item
func ViewItem(w http.ResponseWriter, r *http.Request) {

	ctx := context.Background()
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization")
		w.Header().Set("Access-Control-Max-Age", "3600")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	// Set CORS headers for the main request.
	w.Header().Set("Access-Control-Allow-Origin", "*")
	itemId := r.URL.Query().Get("id")

	authorization := r.Header.Get("Authorization")

	_, UserInfo := validatePaseto(authorization)

	splitInfo := strings.Split(UserInfo.Audience, "||")

	fmt.Print(splitInfo)

	returnThis := &Ret{Status: "fail", Message: "Everything is Required"}

	client, _ := firestore.NewClient(ctx, projectID)

	dsnap, err := client.Collection("items").Doc(itemId).Get(ctx)
	if err != nil {
		returnThis = &Ret{Status: "fail", Message: err.Error()}
		e, _ := json.Marshal(returnThis)
		w.Write(e)
		return
	}
	m := dsnap.Data()

	fmt.Printf("Document data: %#v\n", m)

	jsonString, _ := json.Marshal(m)

	returnThis = &Ret{Status: "success", Message: "View Item", Data: jsonString}
	e, _ := json.Marshal(returnThis)
	w.Write(e)
	return

}

// TradeOffer offers items to other items
func TradeOffer(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Max-Age", "3600")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	// Set CORS headers for the main request.
	w.Header().Set("Access-Control-Allow-Origin", "*")
	ctx := context.Background()
	authorization := r.Header.Get("Authorization")

	_, UserInfo := validatePaseto(authorization)

	splitInfo := strings.Split(UserInfo.Audience, "||")

	fmt.Print("SPLIT", splitInfo)

	returnThis := &Ret{Status: "fail", Message: "Everything is Required"}

	//get the content of body

	var tradeStuffs struct {
		ItemID      string   `json:"itemId"`
		PayDelivery bool     `json:"payDelivery"`
		TradeDate   string   `json:"tradeDate"`
		MyItems     []string `json:"myItems"`
	}

	if err := json.NewDecoder(r.Body).Decode(&tradeStuffs); err != nil {
		returnThis = &Ret{Status: "fail", Message: err.Error()}
		e, _ := json.Marshal(returnThis)
		w.Write(e)
		return
	}

	fmt.Println("DDD", tradeStuffs)

	client, _ := firestore.NewClient(ctx, projectID)

	dsnap, err := client.Collection("items").Doc(tradeStuffs.ItemID).Get(ctx)
	if err != nil {
		returnThis = &Ret{Status: "fail", Message: err.Error()}
		e, _ := json.Marshal(returnThis)
		w.Write(e)
		return
	}
	m := dsnap.Data()

	fmt.Println("DFGHJK", m["owner"])

	ownerInfo, err := client.Collection("users").Doc(m["owner"].(string)).Get(ctx)
	if err != nil {
		returnThis = &Ret{Status: "fail", Message: err.Error()}
		e, _ := json.Marshal(returnThis)
		w.Write(e)
		return
	}
	owner := ownerInfo.Data()

	//get the information for the things being traded
	fmt.Println("my Items", tradeStuffs.MyItems)

	itemsTraded := make(map[string]interface{}, len(tradeStuffs.MyItems))

	traderId := ""

	for i := 0; i < len(tradeStuffs.MyItems); i++ {

		itemsInside, err := client.Collection("items").Doc(tradeStuffs.MyItems[i]).Get(ctx)
		if err != nil {
			returnThis = &Ret{Status: "fail", Message: err.Error()}
			e, _ := json.Marshal(returnThis)
			w.Write(e)
			return
		}
		m := itemsInside.Data()
		traderId = itemsInside.Data()["owner"].(string)
		m["item_id"] = tradeStuffs.MyItems[i]
		itemsTraded["item"+strconv.Itoa(i)] = m
	}

	traderInfoz, err := client.Collection("users").Doc(traderId).Get(ctx)
	if err != nil {
		returnThis = &Ret{Status: "fail", Message: err.Error()}
		e, _ := json.Marshal(returnThis)
		w.Write(e)
		return
	}
	trader := traderInfoz.Data()

	fmt.Println("ITEMSTRADED", itemsTraded)

	insertThis := map[string]interface{}{
		"item":         m,
		"owner_info":   owner,
		"trader_info":  trader,
		"items_traded": itemsTraded,
		"trade_date":   tradeStuffs.TradeDate,
		"pay_delivery": tradeStuffs.PayDelivery,
		"status":       "pending",
	}

	jsonString, _ := json.Marshal(m)
	fmt.Println(jsonString)
	//send email
	//	ctx2 := appengine.NewContext(r)
	//	msg := &mail.Message{
	//		Sender:  "Trade Offer Notifications <admin@talipapops.com>",
	//		To:      []string{"support@talipapops.ml"},
	//		Subject: "Transaction Offer Started",
	//		Body:    fmt.Sprintf(string(jsonString)),
	//	}
	//	if err := mail.Send(ctx2, msg); err != nil {
	//		logmail.Errorf(ctx2, "Couldn't send email: %v", err)
	//	}
	//

	//end send email
	currentTime := time.Now().Unix()

	docRef, _, err := client.Collection("offers").Add(ctx, insertThis)
	jsonStrforSMS := `{"mobilefrom" : "mobileNumberFrom", "mobileto" : "mobileNumberTo", "email":"Transaction Started","premessage":"Transaction Started", "timestamp":"` + strconv.FormatInt(currentTime, 10) + `" }`

	fmt.Println(jsonStrforSMS)

	fmt.Println("DOCREF", docRef)
	if err != nil {
		// Handle any errors in an appropriate way, such as returning them.
		log.Printf("An error has occurred: %s", err)
		returnThis = &Ret{Status: "fail", Message: err.Error()}
		e, _ := json.Marshal(returnThis)
		w.Write(e)
		return
	} else {
		//req, _ := http.NewRequest("POST", sendSMSurl, bytes.NewBuffer([]byte(jsonStrforSMS)))
		//req.Header.Set("Content-Type", "application/json")

		//SMSclient := http.Client{}
		//resp, err := SMSclient.Do(req)
		//if err != nil {
		//	fmt.Println("Unable to reach the server.")
		//	e, _ := json.Marshal(returnThis)
		//	w.Write(e)
		//	return
		//} else {
		//	body, _ := ioutil.ReadAll(resp.Body)
		//	fmt.Println("body=", string(body))
		//}

		docRefID, _ := json.Marshal(docRef.ID)

		returnThis = &Ret{Status: "success", Message: "View Item", Data: docRefID}
		e, _ := json.Marshal(returnThis)
		w.Write(e)
		return
	}
}

// ViewOffer show specific offer
func ViewOffer(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization")
		w.Header().Set("Access-Control-Max-Age", "3600")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	// Set CORS headers for the main request.
	w.Header().Set("Access-Control-Allow-Origin", "*")
	offerId := r.URL.Query().Get("id")

	ctx := context.Background()
	authorization := r.Header.Get("Authorization")

	_, UserInfo := validatePaseto(authorization)

	splitInfo := strings.Split(UserInfo.Audience, "||")

	fmt.Print(splitInfo)

	returnThis := &Ret{Status: "fail", Message: "Everything is Required"}

	client, _ := firestore.NewClient(ctx, projectID)

	dsnap, err := client.Collection("offers").Doc(offerId).Get(ctx)
	if err != nil {
		returnThis = &Ret{Status: "fail", Message: err.Error()}
		e, _ := json.Marshal(returnThis)
		w.Write(e)
		return
	}
	m := dsnap.Data()

	fmt.Printf("Document data: %#v\n", m)

	jsonString, _ := json.Marshal(m)

	returnThis = &Ret{Status: "success", Message: "View Item", Data: jsonString}
	e, _ := json.Marshal(returnThis)
	w.Write(e)
	return

}
