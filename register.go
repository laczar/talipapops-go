package talipapops

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/matthewhartstonge/argon2"
	"google.golang.org/api/iterator"
)

const otpChars = "1234567890"

// SMSService for call function to send SMS
const sendSMSurl = "https://asia-east2-talipapops-wapak.cloudfunctions.net/SMSService"

// GenerateOTP to genrate random number
func GenerateOTP(length int) (string, error) {
	buffer := make([]byte, length)
	_, err := rand.Read(buffer)
	if err != nil {
		return "", err
	}

	otpCharsLength := len(otpChars)
	for i := 0; i < length; i++ {
		buffer[i] = otpChars[int(buffer[i])%otpCharsLength]
	}

	return string(buffer), nil
}

// Register User 1st
func Register(w http.ResponseWriter, r *http.Request) {

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

	// just say a return for function.
	argon := argon2.DefaultConfig()
	var d struct {
		Email    string `json:"email"`
		Mobile   string `json:"mobile"`
		Password string `json:"password"`
	}

	type Ret struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}

	//connect to firestore
	ctx := context.Background()
	projectID := "talipapops-wapak"
	client, _ := firestore.NewClient(ctx, projectID)

	returnThis := &Ret{Status: "fail", Message: "Everything is Required"}

	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {

		e, err := json.Marshal(returnThis)
		if err != nil {
			fmt.Println(err)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write(e)
		return
	}

	//check if email or mobile exists
	emailQuery := client.Collection("users").Where("email", "==", d.Email).Documents(ctx)
	for {
		doc, err := emailQuery.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			fmt.Println("EMQUERY=", emailQuery)
			returnThis = &Ret{Status: "fail", Message: "Email is already in used"}
			e, _ := json.Marshal(returnThis)
			w.Write(e)
			return
		}
		if doc.Data() != nil {
			fmt.Println("EMQUERY=", emailQuery)
			returnThis = &Ret{Status: "fail", Message: "Email is already in used"}
			e, _ := json.Marshal(returnThis)
			w.Write(e)
			return
		}
	}

	//check if email or mobile exists
	mobileQuery := client.Collection("users").Where("mobile", "==", d.Mobile).Documents(ctx)
	for {
		doc2, err := mobileQuery.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			fmt.Println("MOBQUERY=", mobileQuery)
			returnThis = &Ret{Status: "fail", Message: "Mobile Number is already in used"}
			e, _ := json.Marshal(returnThis)
			w.Write(e)
			return
		}
		if doc2.Data() != nil {
			fmt.Println("MOBILE=", emailQuery)
			returnThis = &Ret{Status: "fail", Message: "Mobile is already in used"}
			e, _ := json.Marshal(returnThis)
			w.Write(e)
			return
		}
	}

	if d.Password == "" || d.Mobile == "" || d.Email == "" {

		returnThis = &Ret{Status: "fail", Message: "Everything is Required"}
		e, err := json.Marshal(returnThis)
		if err != nil {
			fmt.Println(err)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write(e)
		return
	}

	encoded, err := argon.HashEncoded([]byte(d.Password))
	if err != nil {
		e, _ := json.Marshal(returnThis)
		w.Write(e)
		return
	}

	currentTime := time.Now().Unix()

	otp, _ := GenerateOTP(4)

	jsonValue := map[string]interface{}{
		"email":        d.Email,
		"mobile":       d.Mobile,
		"password":     string(encoded),
		"otp":          otp,
		"date_created": currentTime,
	}

	jsonStrforSMS := `{"mobilefrom" : "mobileNumberFrom", "mobileto" : "` + d.Mobile + `", "email":"` + d.Email + `","premessage":"Your OTP Registration number is ` + otp + `. This is strictly for mobile no. ` + d.Mobile + `", "timestamp":"` + strconv.FormatInt(currentTime, 10) + `" }`

	_, _, err = client.Collection("users").Add(ctx, jsonValue)

	if err != nil {
		// Handle any errors in an appropriate way, such as returning them.
		log.Printf("An error has occurred: %s", err)
		e, _ := json.Marshal(returnThis)
		w.Write(e)
		return
	} else {

		//send to SMS
		req, _ := http.NewRequest("POST", sendSMSurl, bytes.NewBuffer([]byte(jsonStrforSMS)))
		req.Header.Set("Content-Type", "application/json")

		SMSclient := http.Client{}
		resp, err := SMSclient.Do(req)
		if err != nil {
			fmt.Println("Unable to reach the server.")
			e, _ := json.Marshal(returnThis)
			w.Write(e)
			return
		} else {
			body, _ := ioutil.ReadAll(resp.Body)
			fmt.Println("body=", string(body))
		}

		returnThis := &Ret{Status: "success", Message: jsonStrforSMS}
		e, _ := json.Marshal(returnThis)
		w.Write(e)
		return
	}

}

// VerifyOTP verifies mobile registration number
func VerifyOTP(w http.ResponseWriter, r *http.Request) {

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

	var d struct {
		Email  string `json:"email"`
		Mobile string `json:"mobile"`
		Otp    string `json:"otp"`
	}

	type Ret struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		DocID   string `json:"docid"`
		Token   string `json:"token"`
	}

	//coonect to firestore
	ctx := context.Background()
	projectID := "talipapops-wapak"
	client, _ := firestore.NewClient(ctx, projectID)

	returnThis := &Ret{Status: "fail", Message: "Everything is Required"}

	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {

		e, err := json.Marshal(returnThis)
		if err != nil {
			fmt.Println(err)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write(e)
		return
	}
	//check if email or mobile exists
	userQuery := client.Collection("users").Where("email", "==", d.Email).Where("mobile", "==", d.Mobile).Where("otp", "==", d.Otp).Documents(ctx)
	for {

		fmt.Println("IM IN")
		doc, err := userQuery.Next()

		if err == iterator.Done {
			break
		}
		if err != nil {
			returnThis = &Ret{Status: "fail", Message: "Iterator Error"}
			e, _ := json.Marshal(returnThis)
			w.Write(e)
			return
		}
		if doc.Data() != nil {

			fmt.Println("IM IN TOO", doc.Data())
			//insert the verified status
			_, err = client.Collection("users").Doc(doc.Ref.ID).Set(ctx, map[string]interface{}{
				"verifiedAt": time.Now().Format(time.RFC3339),
			}, firestore.MergeAll)
			if err != nil {
				returnThis = &Ret{Status: "fail", Message: "Cannot Insert date"}
				e, _ := json.Marshal(returnThis)
				w.Write(e)
				return
			}

			pToken := createToken(d.Email, doc.Ref.ID)

			returnThis = &Ret{Status: "success", Message: "Thank you, you are now verified! " + d.Email, DocID: doc.Ref.ID, Token: pToken}
			e, _ := json.Marshal(returnThis)
			w.Write(e)
			return
		}
		if doc.Data() == nil {

			fmt.Println("DOC NLL", doc.Data())
			returnThis = &Ret{Status: "fail", Message: "User not found"}
			e, _ := json.Marshal(returnThis)
			w.Write(e)
			return
		}
	}

	returnThis = &Ret{Status: "fail", Message: "Invalid OTP"}
	e, _ := json.Marshal(returnThis)
	w.Write(e)
	return
}
