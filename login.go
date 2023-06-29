package talipapops

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"cloud.google.com/go/firestore"
	"github.com/matthewhartstonge/argon2"
	"google.golang.org/api/iterator"
)

//Login executes to login users
func Login(w http.ResponseWriter, r *http.Request) {

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

	var d struct {
		Email    string `json:"email"`
		Password string `json:"password"`
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

	// encoded, err := argon.HashEncoded([]byte(d.Password))
	// if err != nil {
	// 	e, _ := json.Marshal(returnThis)
	// 	w.Write(e)
	// 	return
	// }

	//check if email or mobile exists
	userQuery := client.Collection("users").Where("email", "==", d.Email).Documents(ctx)
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

			fmt.Println("IM IN TOO", doc.Data()["password"])
			//insert the verified status

			encoded := fmt.Sprintf("%v", doc.Data()["password"])
			ok, err := argon2.VerifyEncoded([]byte(d.Password), []byte(encoded))
			if err != nil {
				panic(err)
			}

			if ok {

				pToken := createToken(d.Email, doc.Ref.ID)

				returnThis = &Ret{Status: "success", Message: "Welcome Back " + d.Email, DocID: doc.Ref.ID, Token: pToken}
				e, _ := json.Marshal(returnThis)
				w.Write(e)
				return
			} else {
				returnThis = &Ret{Status: "fail", Message: "ERROR" + d.Email}
				e, _ := json.Marshal(returnThis)
				w.Write(e)
				return
			}

		}
		if doc.Data() == nil {

			fmt.Println("DOC NLL", doc.Data())
			returnThis = &Ret{Status: "fail", Message: "Invalid Username or Password"}
			e, _ := json.Marshal(returnThis)
			w.Write(e)
			return
		}
	}

	returnThis = &Ret{Status: "fail", Message: "Invalid Username or Password"}
	e, _ := json.Marshal(returnThis)
	w.Write(e)
	return
}
