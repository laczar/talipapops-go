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
)

//Ret is the return expected for API
type Ret struct {
	Status  string          `json:"status"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

var (
	storageClient *storage.Client
)

const projectID = "talipapops-wapak"
const bucketName = "files.talipapops.ml"

//UpdateProfile provides updating of avatar and details
func UpdateProfile(w http.ResponseWriter, r *http.Request) {

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

	//let get the Document ID
	_, UserInfo := validatePaseto(authorization)

	splitInfo := strings.Split(UserInfo.Audience, "||")

	returnThis := &Ret{Status: "fail", Message: "Everything is Required"}

	r.ParseMultipartForm(0)

	fmt.Println("FirstName", r.FormValue("firstName"))
	fmt.Println("FirstName", r.PostFormValue("firstName"))

	// process image
	profilepicture := "/assets/images/avatar.svg"
	imgData := ""
	imgMimeType := ""

	if r.FormValue("profilePicture") != "/assets/images/avatar.svg" {

		if strings.HasPrefix(r.FormValue("profilePicture"), "https://storage.googleapis.com/files.talipapops.ml") {
			profilepicture = r.FormValue("profilePicture")
		} else {
			imgProps := strings.Split(r.FormValue("profilePicture"), ",")
			imgTypes := strings.Split(imgProps[0], ";")
			imgTypes2 := strings.Split(imgTypes[0], ":")

			imgData = imgProps[1]
			imgMimeType = imgTypes2[1]
			imgTypes3 := strings.Split(imgMimeType, "/")

			imgExtension := imgTypes3[1]

			reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(imgData))

			image, _, err := image.Decode(reader)
			if err != nil {
				log.Fatal(err)
			}

			newImage := resize.Resize(388, 0, image, resize.Lanczos3)

			client, err := storage.NewClient(ctx)
			if err != nil {
				fmt.Println(err)
			}

			bucket := client.Bucket(bucketName)

			name := splitInfo[1] + "/profile/" + strconv.FormatInt(time.Now().Unix(), 10) + "." + imgExtension

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

			profilepicture = "https://storage.googleapis.com/" + bucketName + "/" + name
		}

	}

	if r.FormValue("firstName") != "" && r.FormValue("lastName") != "" && r.FormValue("city") != "" && r.FormValue("barangay") != "" {

		client, _ := firestore.NewClient(ctx, projectID)

		_, err := client.Collection("users").Doc(splitInfo[1]).Set(ctx, map[string]interface{}{
			"firstname":      r.FormValue("firstName"),
			"lastname":       r.FormValue("lastName"),
			"city":           r.FormValue("city"),
			"region":         r.FormValue("region"),
			"barangay":       r.FormValue("barangay"),
			"address":        r.FormValue("address"),
			"profilepicture": profilepicture,
			"status":         1,
		}, firestore.MergeAll)

		if err != nil {
			returnThis = &Ret{Status: "fail", Message: "Cannot Update Profile"}
			e, _ := json.Marshal(returnThis)
			w.Write(e)
			return
		} else {
			returnData := map[string]interface{}{
				"id":    splitInfo[1],
				"email": splitInfo[0],
			}
			jsonString, _ := json.Marshal(returnData)

			returnThis = &Ret{Status: "success", Message: "Profile Updated Successfully", Data: jsonString}
			e, _ := json.Marshal(returnThis)
			w.Write(e)
			return
		}

	} else {
		returnThis = &Ret{Status: "fail", Message: "Cannot Update Profile"}
		e, _ := json.Marshal(returnThis)
		w.Write(e)
		return
	}

}

//GetProfile of the user
func GetProfile(w http.ResponseWriter, r *http.Request) {

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

	//let get the Document ID
	_, UserInfo := validatePaseto(authorization)
	splitInfo := strings.Split(UserInfo.Audience, "||")
	docID := splitInfo[1]

	returnThis := &Ret{Status: "fail", Message: "Everything is Required"}

	client, _ := firestore.NewClient(ctx, projectID)

	userInfo, err := client.Collection("users").Doc(docID).Get(ctx)

	if err != nil {
		returnThis = &Ret{Status: "fail", Message: "User had Issues"}
		e, _ := json.Marshal(returnThis)
		w.Write(e)
		return
	} else {
		m := userInfo.Data()
		fmt.Printf("Document data: %#v\n", m)
		jsonString, _ := json.Marshal(m)
		returnThis = &Ret{Status: "success", Message: "Info fetched", Data: jsonString}
		e, _ := json.Marshal(returnThis)
		w.Write(e)
		return
	}

}
