package talipapops

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/o1egl/paseto"
	"golang.org/x/crypto/argon2"
)

type params struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	saltLength  uint32
	keyLength   uint32
}

const salt = "ASIN ASIN"

func createToken(email string, id string) string {

	//ceate JiT
	p := &params{
		memory:      64 * 1024,
		iterations:  3,
		parallelism: 2,
		saltLength:  16,
		keyLength:   32,
	}

	encodedHash, err := generateFromPassword(id, p)
	if err != nil {
		log.Fatal(err)
	}

	//Create PASETO tokens not JWT
	symmetricKey := []byte(salt) // Must be 32 bytes
	now := time.Now()
	exp := now.Add(24 * time.Hour * 30)

	jsonToken := paseto.JSONToken{
		Audience:   email + "||" + id,
		Issuer:     "Talipapops",
		Jti:        string(encodedHash),
		Subject:    "TalipapopsAuth",
		IssuedAt:   now,
		Expiration: exp,
	}

	// Add custom claim    to the token
	jsonToken.Set("data", "TalipapopsUser")
	footer := "Talipapops Inc"

	v2 := paseto.NewV2()
	pToken, err := v2.Encrypt(symmetricKey, jsonToken, footer)
	if err != nil {
		log.Fatal("ERRORRRR", err)
	}

	fmt.Println("PASETO", pToken)

	return pToken
}

func generateFromPassword(password string, p *params) (encodedHash string, err error) {
	salt, err := generateRandomBytes(p.saltLength)
	if err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, p.iterations, p.memory, p.parallelism, p.keyLength)

	// Base64 encode the salt and hashed password.
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	// Return a string using the standard encoded hash representation.
	encodedHash = fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, p.memory, p.iterations, p.parallelism, b64Salt, b64Hash)

	return encodedHash, nil
}

func generateRandomBytes(n uint32) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func validatePaseto(token string) (bool, paseto.JSONToken) {

	v2 := paseto.NewV2()
	var newJSONToken paseto.JSONToken
	sKey := []byte(salt) // Must be 32 bytes

	err := v2.Decrypt(token, sKey, &newJSONToken, nil)

	if err == nil {
		return true, newJSONToken
	} else {
		fmt.Println(err)
		return false, newJSONToken
	}
}

//ValidateToken make the cookie token valid
func ValidateToken(w http.ResponseWriter, r *http.Request) {

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
		Token string `json:"token"`
	}

	type Ret struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		DocID   string `json:"docid"`
		Email   string `json:"email"`
	}

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

	token := d.Token

	ret, jsonInfo := validatePaseto(token)

	if ret == true {
		sInfo := strings.Split(jsonInfo.Audience, "||")
		returnThis := &Ret{Status: "success", Message: "Valid Token", Email: sInfo[0], DocID: sInfo[1]}
		e, _ := json.Marshal(returnThis)
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write(e)
		return
	} else {
		returnThis = &Ret{Status: "fail", Message: "Invalid Token"}
		e, _ := json.Marshal(returnThis)
		w.Write(e)
		return
	}

}
