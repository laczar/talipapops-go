package talipapops

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strings"

	"cloud.google.com/go/firestore"
)

func createClient(ctx context.Context) *firestore.Client {
	// Sets your Google Cloud Platform project ID.
	projectID := "talipapops"

	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	// Close client when done with
	// defer client.Close()
	return client

}

// random generator
func rangeIn(low, hi int) int {
	return low + rand.Intn(hi-low)
}

// Hello the name of func
func Hello(w http.ResponseWriter, r *http.Request) {

	// **************
	name := r.URL.Query().Get("name")
	if name == "" {
		name = "someone"
	}
	// **************
	fmt.Fprintf(w, "Hello, %s!", name)

}

// HelloWorld inititae if function is ok
func HelloWorld(w http.ResponseWriter, r *http.Request) {
	var d struct {
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		fmt.Fprint(w, "Hello World Mankind!")
		return
	}
	if d.Message == "" {
		fmt.Fprint(w, "Hello World Humans!")
		return
	}
	fmt.Fprint(w, html.EscapeString(d.Message))
}

// PreRegister used for initial listing
func PreRegister(w http.ResponseWriter, r *http.Request) {

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

	var d struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		City  string `json:"city"`
	}

	type Ret struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}

	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {

		returnThis := &Ret{Status: "fail", Message: "Everything is kailangan"}
		e, err := json.Marshal(returnThis)

		log.Print("Logging in Go!", e)

		if err != nil {
			fmt.Println(err)
			return
		}

		ctx := context.Background()

		projectID := "talipapops-wapak"

		client, _ := firestore.NewClient(ctx, projectID)

		log.Print("client", client)

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		returnThis = &Ret{Status: "success", Message: d.Name + "Added"}
		e, _ = json.Marshal(returnThis)
		w.Write(e)
		return
	}

	if d.Name == "" || d.Email == "" {

		returnThis := &Ret{Status: "fail", Message: "Everything is Needed"}
		e, err := json.Marshal(returnThis)
		if err != nil {
			fmt.Println(err)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write(e)
		return
	} else {

		ctx := context.Background()

		projectID := "talipapops-"

		client, _ := firestore.NewClient(ctx, projectID)

		log.Print("client", client)

		fmt.Println("CLIENT", client)
		_, _, err := client.Collection("registrants").Add(ctx, map[string]interface{}{
			"Name":  d.Name,
			"Email": d.Email,
			"City":  d.City,
		})

		if err != nil {
			log.Fatalf("Failed adding aturing: %v", err)
		}
		returnThis := &Ret{Status: "success", Message: d.Name + " successfully Added!"}
		e, err := json.Marshal(returnThis)
		if err != nil {
			fmt.Println(err)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write(e)
	}

}

// SMSServiceOld connect to semaphore to send SMS
func SMSServiceOld(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		// w.Header().Set("Access-Control-Allow-Headers", "Authorization")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Max-Age", "3600")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	// Set CORS headers for the main request.
	w.Header().Set("Access-Control-Allow-Origin", "*")

	//semaphore
	semaphoreURL := "https://api.semaphore.co/api/v4/messages"
	apiKey := "this is semaphore key"

	var smsData struct {
		PreMessage string `json:"premessage"`
		MobileTo   string `json:"mobileto"`
		MobileFrom string `json:"mobilefrom"`
		Email      string `json:"email"`
		Otp        string `json:"otp"`
	}

	type Ret struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}

	returnThis := &Ret{Status: "fail", Message: "Everything is kailangan"}

	if err := json.NewDecoder(r.Body).Decode(&smsData); err != nil {

		e, err := json.Marshal(returnThis)

		if err != nil {
			fmt.Println(err)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		returnThis = &Ret{Status: "fail", Message: err.Error()}
		e, _ = json.Marshal(returnThis)
		w.Write(e)
		return
	} else {

		data := url.Values{}
		data.Set("apikey", apiKey)
		data.Add("number", smsData.MobileTo)
		data.Add("message", smsData.PreMessage)
		data.Add("sendername", "Talipapops")

		req, _ := http.NewRequest("POST", semaphoreURL, strings.NewReader(data.Encode()))
		req.PostForm = data
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		fmt.Println("ENCODE", data.Encode())
		resp, _ := http.DefaultClient.Do(req)

		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		returnThis = &Ret{Status: "success", Message: string(body)}
		e, _ := json.Marshal(returnThis)
		w.Write(e)
		return

	}
}

// SMSService connect to EasySMS to send SMS
func SMSService(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		// w.Header().Set("Access-Control-Allow-Headers", "Authorization")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Max-Age", "3600")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	// Set CORS headers for the main request.
	w.Header().Set("Access-Control-Allow-Origin", "*")

	//easySMS
	easySMSUrl := "https://easysms.com.ph/api/rest/send_sms"
	apiKey := "apikey"
	apiSecret := "apiSecret"

	var smsData struct {
		PreMessage string `json:"premessage"`
		MobileTo   string `json:"mobileto"`
		MobileFrom string `json:"mobilefrom"`
		Email      string `json:"email"`
		Otp        string `json:"otp"`
	}

	type Ret struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}

	returnThis := &Ret{Status: "fail", Message: "Everything is kailangan"}

	if err := json.NewDecoder(r.Body).Decode(&smsData); err != nil {

		e, err := json.Marshal(returnThis)

		if err != nil {
			fmt.Println(err)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		returnThis = &Ret{Status: "fail", Message: err.Error()}
		e, _ = json.Marshal(returnThis)
		w.Write(e)
		return
	} else {

		data := url.Values{}
		data.Set("key", apiKey)
		data.Set("secret", apiSecret)
		data.Add("mobile", smsData.MobileTo)
		data.Add("message", smsData.PreMessage)
		data.Add("sender", "Talipapops")

		req, _ := http.NewRequest("POST", easySMSUrl, strings.NewReader(data.Encode()))
		req.PostForm = data
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		fmt.Println("ENCODE", data.Encode())
		resp, _ := http.DefaultClient.Do(req)

		fmt.Println("RESPONSE", resp)

		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		returnThis = &Ret{Status: "success", Message: string(body)}
		e, _ := json.Marshal(returnThis)
		w.Write(e)
		return

	}
}
