package main

/**libraries that the lock-firmware will use throught the code*/
import (
	"fmt"
	"flag"
	"os"
	"io"
	"strings"
	"bufio"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/asn1"
	"crypto/rand"
	"net/http"
	"encoding/json"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"math/big"
	"bytes"
	"crypto/tls"
	"crypto/sha256"
	"strconv"
)

/**object Door; used to read/send json response from controller*/
type Door struct {
    Id     int64  `json:"id,omitempty"`
    System int64  `json:"system,omitempty"`
    KeyX *big.Int     `json:"keyX,omitempty"`
    KeyY *big.Int     `json:"keyY,omitempty"`
    Challenge []byte  `json:"challenge,omitempty"`
    Response  []byte  `json:"response,omitempty"`
    Name   string `json:"name,omitempty"`
    Method int    `json:"method,omitemtpy"`
    Totp string `json:"totp,omitempty"`
}

/**object RandS; used to send json response to controller with hashed challenge string*/
type RandS struct {
	R	*big.Int
	S	*big.Int
}

/**object UniqueLockNumber; used to read json response from controller of unique lock id*/
type UniqueLockNumber struct {
	Id	int64	`json:"id"`
}

/**object AccessDoor; used to send pin code to the server*/
type AccessDoor struct {
	Pin	string
}

/**object AccessDoorCard; used to send card number to the server*/
type AccessDoorCard struct {
	Card string
}

func main(){
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	var (
		lockUniqueID int64
		lockPrivateKey *ecdsa.PrivateKey
		isRegistered bool = false
		err error
	)

	if _, err = os.Stat("lockID.txt"); os.IsNotExist(err) {
		fmt.Println("\n---Lock is Un-Registered---\n")

		lockPrivateKey, err = ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
		if err != nil {
			panic(err)
		}

		lockPrivateKeyBytes, err := x509.MarshalECPrivateKey(lockPrivateKey)
		if err != nil {
			panic(err)
		}

		pemFile, err := os.Create("lockPrivateKey.pem")
		if err != nil {
			panic(err)
		}

		var pemPrivateKeyBlock = &pem.Block{
			Type: "ECDSA Private Key",
			Bytes: lockPrivateKeyBytes,
		}

		err = pem.Encode(pemFile, pemPrivateKeyBlock)
		if err != nil {
			panic(err)
		}

		pemFile.Close()
	}else{
		fmt.Println("\n---Lock is Registered---")

		isRegistered = true

		lockIDfile, err := os.Open("lockID.txt")
		if err != nil {
			panic(err)
		}

		scanner := bufio.NewScanner(lockIDfile)

		line := strings.TrimSpace(scanner.Text())

		lockUniqueID, err = strconv.ParseInt(line,10,64)
		if err != nil {
			panic(err)
		}

		pemFile, err := os.Open("lockPrivateKey.pem")
		if err != nil {
			panic(err)
		}

		privatePublicKeyInfo, _ := pemFile.Stat()
		var size int64 = privatePublicKeyInfo.Size()
		pemBytes := make([]byte, size)

		buffer := bufio.NewReader(pemFile)
		_, err = buffer.Read(pemBytes)

		data, _ := pem.Decode([]byte(pemBytes))

		pemFile.Close()

		privatePublicKeyImported, err := x509.ParseECPrivateKey(data.Bytes)
		if err != nil{
			panic(err)
		}

		lockPrivateKey = privatePublicKeyImported

		jwt, err := login(lockUniqueID, lockPrivateKey)
		if jwt != "" && err == nil{
			fmt.Println("---Login Successful---")
			isRegistered = true
		}else{
			os.Exit(1)
		}
	}

	var (
		pinPipe string
		cardPipe string
	)

	flag.StringVar(&pinPipe, "pin-pipe", "pin-pipe", "Location of a named pipe to read pin-data from")
	flag.StringVar(&cardPipe, "card-pipe", "card-pipe", "Location of a named pipe to read card-data from")
	flag.Parse()

	fmt.Println("pin pipe ", pinPipe)
	fmt.Println("card pipe ", cardPipe)

	filePipe0, err := os.OpenFile(pinPipe, os.O_RDONLY, os.ModeNamedPipe)
	if err != nil{
		panic(err)
	}

	fmt.Println("Opened pin-pipe")

	reader := bufio.NewReader(filePipe0)

	filePipe1, err := os.OpenFile(cardPipe, os.O_RDONLY, os.ModeNamedPipe)
	if err != nil{
		panic(err)
	}

	fmt.Println("Opened card-pipe")

	reader1 := bufio.NewReader(filePipe1)

	for{
		//if loop will read the data from filepipe1
		if line, err := reader1.ReadString('\n'); err == nil{
			fmt.Println("-----------------------------------------------")
			fmt.Println("\n--Card: ", line)
			line = strings.TrimSpace(line)

			fmt.Println("---Accessing Door---")

			accessRequestBody, err := json.Marshal(AccessDoorCard{
				Card: line,
			})
			if err != nil {
				continue
			}

			var accessDoorURL string = "https://bast-security.xyz:8080/locks/" + strconv.FormatInt(lockUniqueID,10) + "/access"
			accessDoorResponse, err := http.Post(accessDoorURL,"",bytes.NewBuffer(accessRequestBody))
			if err != nil{
				continue
			}

			defer accessDoorResponse.Body.Close()
			if accessDoorResponse.StatusCode == 200{
				fmt.Println("---Access Granted---")
			}else{
				fmt.Println("---Access Denied---")
			}
			fmt.Println("-----------------------------------------------")

		}else if err != io.EOF{
			panic(err)
		}

		//if loop will read the data from filepipe0
		if line, err := reader.ReadString('\n'); err == nil{
			fmt.Println("-----------------------------------------------")
			fmt.Println("\n--Pin: ", line)
			//removes the '\n\' from line
			line = strings.TrimSpace(line)

			/**if look checks to see if lock is registered or not*/
			if !isRegistered{
				fmt.Println("\n---Lock is Un-Registered---\n")

				tokens := strings.Split(line, "*")

				if len(tokens) < 2 {
					continue
				}

				systemID, err := strconv.ParseInt(tokens[0], 10, 64)
				if err != nil{
					continue
				}

				//creating a json string containing systemID and public x and y key
				requestBody, err := json.Marshal(Door{
					Name: fmt.Sprintf("Lock %s", tokens[1]),
					System: systemID,
					KeyX: lockPrivateKey.PublicKey.X,
					KeyY: lockPrivateKey.PublicKey.Y,
					Totp: tokens[1],
				})
				//incase requestBody could not be generated
				if err != nil{
					fmt.Println("requestBody could not be created")
					continue
				}

				//post request will send information to the server
				registerResponse, err := http.Post("https://bast-security.xyz:8080/locks/register","",bytes.NewBuffer(requestBody))
				//incase registerResponse could not post
				if err != nil{
					fmt.Println("Registration Failed")
					continue
				}

				defer registerResponse.Body.Close()

				registerResponseBody, err := ioutil.ReadAll(registerResponse.Body)
				if err != nil{
					fmt.Println("Failed to read response")
					continue
				}

				registerResponseJSON := UniqueLockNumber{}
				json.Unmarshal([]byte(string(registerResponseBody)), &registerResponseJSON)
				lockUniqueID = registerResponseJSON.Id

				lockIDfile, err := os.Create("lockID.txt")
				if err != nil{
					fmt.Println("lockID.txt unable to be created")
					continue
				}

				//writing into the lockID.txt file
				_, err = io.WriteString(lockIDfile, strconv.FormatInt(lockUniqueID, 10))
				//incase lockUniqueID was not written into the lockIDfile
				if err != nil{
					fmt.Println("Could not write to lockID.txt")
					continue
				}

				/*****************************************************/
				/************************LOGIN************************/
				/*****************************************************/
				jwt, err := login(lockUniqueID, lockPrivateKey)
				if jwt != "" && err == nil{
					fmt.Println("\n---Login Successful---\n")
					isRegistered = true
				}else{
					continue
				}

			/***************************************************************/
			}else if isRegistered{
				fmt.Println("---Accessing Door---")

				//creating a json string containing pin number to send to server
				accessRequestBody, err := json.Marshal(AccessDoor{
					Pin: line,
				})
				//incase accessRequestBody could not be generated
				if err != nil{
					fmt.Println("Failed to marshal json")
					continue
				}

				//post request will send accessDoor information to the server
				var accessDoorURL string = "https://bast-security.xyz:8080/locks/" + strconv.FormatInt(lockUniqueID,10) + "/access"
				accessDoorResponse, err := http.Post(accessDoorURL,"",bytes.NewBuffer(accessRequestBody))
				//incase accessDoorResponse could not post
				if err != nil{
					fmt.Println("Request failed")
					continue
				}

				defer accessDoorResponse.Body.Close()

				//if loop checks to see if the pin entered has accesses to the door
				if accessDoorResponse.StatusCode == 200{
					fmt.Println("---Accessing Granted---")
				}else{
					fmt.Println("---Accessing Denied---")
				}
				fmt.Println("-----------------------------------------------")
			}
		}else if err != io.EOF{
			panic(err)
		}
	}
}

/**Login function for the lock to connect to server*/
func login(lockUniqueID int64, lockPrivateKey *ecdsa.PrivateKey)(string, error){
	//get request will receive a challenge string from the server
	loginResponse, err := http.Get("https://bast-security.xyz:8080/locks/" + strconv.FormatInt(lockUniqueID,10))
	//incase the server dosent respond
	if err != nil{
		fmt.Println("Challenge String was not recieved")
		panic(err)
	}

	defer loginResponse.Body.Close()

	//reads the response from the server
	loginResponseBody, err := ioutil.ReadAll(loginResponse.Body)
	//incase not able to read the response from the server
	if err != nil{
		panic(err)
	}

	var response map[string][]byte
	if err := json.Unmarshal(loginResponseBody, &response); err != nil {
		panic(err)
	}

	challengeString := response["challenge"]
	hashedChallengeString := sha256.Sum256([]byte(challengeString))

	r, s, err := ecdsa.Sign(rand.Reader, lockPrivateKey, hashedChallengeString[:])
	if err != nil{
		panic(err)
	}

	var payload []byte
	if payload, err = asn1.Marshal(struct{ R, S *big.Int }{ r, s }); err != nil {
		panic(err)
	}

	//sending the hashed challenge string to the server
	var challengeStringRequest string = "https://bast-security.xyz:8080/locks/" + strconv.FormatInt(lockUniqueID,10) + "/login"
	challengeStringResponse, err := http.Post(challengeStringRequest, "", bytes.NewBuffer(payload))
	//incase post request was not sent
	if err!= nil{
		panic(err)
	}

	defer challengeStringResponse.Body.Close()

	//successful, the server responded
	if challengeStringResponse.StatusCode == 200{
		//reads the response from the server
		challengeStringBody, err := ioutil.ReadAll(challengeStringResponse.Body)
		//incase not able to read the response from the server
		if err != nil{
			panic(err)
		}

		//saves the response from the server to jwt
		var jwt = string(challengeStringBody)

		return jwt, nil
	}else{
		return "", fmt.Errorf("Failed to login")
	}
}
