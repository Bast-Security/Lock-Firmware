package main

import (
	"fmt"
	"flag"
	"os"
	"io"
	"strings"
	"bufio"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"net/http"
	"encoding/json"
	"crypto/x509"
	"encoding/pem"
	"log"
	"io/ioutil"
	"math/big"
	"bytes"
	"crypto/tls"
	"crypto/sha256"
)

type Door struct {
    Id     int64  `json:"id,omitempty"`
    System string  `json:"system,omitempty"`
    KeyX *big.Int     `json:"keyX,omitempty"`
    KeyY *big.Int     `json:"keyY,omitempty"`
    Challenge []byte  `json:"challenge,omitempty"`
    Response  []byte  `json:"response,omitempty"`
    Name   string `json:"name,omitempty"`
    Method int    `json:"method,omitemtpy"`
}

type RandS struct {
	R	*big.Int
	S	*big.Int
}

func main() {

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	//lock's unique id 
	var lockUniqueID string
	//lock's private key
	var privateKey *ecdsa.PrivateKey
	var err error
	var isRegistered = false;

	//check to see if a file exists with the privateKey if not then the file is created
	fileInfo, err := os.Stat("lockId.txt")
	if err != nil{
		//no file is found so then the lock registers with the server
		fmt.Println("\n\nUn-Registered\n")

		//generate an ecdsa key pair using the p384 curve once, when lock is used
		privateKey, err = ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
		//incase of an error
		if err != nil{
			panic(err)
		}

		///TESETING
		fmt.Println("PrivateKey: ", privateKey)
		fmt.Println("privateKeyX: ",privateKey.PublicKey.X)
		fmt.Println("privateKeyY: ",privateKey.PublicKey.Y)
		fmt.Println("unique Lock ID: ",lockUniqueID)
		fmt.Println("filename: ", fileInfo, "\n\n")

		//used to creating into byte form to send to request
		privateKeyBytes, err := x509.MarshalECPrivateKey(privateKey)
		if err != nil{
			panic(err)
		}

		//saving the privateKeybytes into a file named publicKey.pem
		pemFile, err := os.Create("publicKey.pem")
		if err != nil{
			os.Exit(1)
		}

		//encoding private key 
		var pemPrivateBlock = &pem.Block{
			Type: "ECDSA PRIVATE KEY",
			Bytes: privateKeyBytes,
		}

		err = pem.Encode(pemFile, pemPrivateBlock)
		if err != nil{
			os.Exit(1)
		}

		pemFile.Close()

	}else{

		isRegistered = true

		//file found so open file and save unique id number
		fmt.Println("\n\nRegistered\n")

		//opens the lockId.txt file and saves the lock unique Id to the corresponding variable
		lockIdFile, _ := os.Open("lockId.txt")
		//create a new scanner for the file
		scanner := bufio.NewScanner(lockIdFile)
		//loop over all the lines in the text file
		for scanner.Scan(){
			lockUniqueID = scanner.Text()
		}

		//for testing purposes
		fmt.Println("lockUniqueId: ", lockUniqueID)

		//opens the publicKey.pem file to get the private/public key of the lock
		publicKeyFile, err := os.Open("publicKey.pem")
		//checks for error
		if err != nil{
			fmt.Println(err)
			os.Exit(1)
		}
		
		privatePublicKeyInfo, _ := publicKeyFile.Stat()
		var size int64 = privatePublicKeyInfo.Size()
		pembytes := make([]byte, size)

		buffer := bufio.NewReader(publicKeyFile)
		stop , err := buffer.Read(pembytes)
		fmt.Println("Stop: ",stop)

		data, _ := pem.Decode([]byte(pembytes))

		publicKeyFile.Close()

		//decode the privatepublic key
		publicKeyImported, err := x509.ParseECPrivateKey(data.Bytes)
		if err != nil{
			fmt.Println(err)
			os.Exit(1)
		}

		//for testing purposes
		fmt.Println("Private Key: ", publicKeyImported)
		fmt.Println("\n")



		////LOGIN put login in a function, call function for both spots

	}

	//will be reading the argument that the user provides
	flag.Parse()

	//saves the path from the argument that the user provided
	var pathName = flag.Args()

	//converting pathName variable from []string to string
	file, err := os.OpenFile(strings.Join(pathName,""), os.O_RDONLY, os.ModeNamedPipe)
	
	//checks to see that file exists
	if err != nil{
		panic(err)
	}

	reader := bufio.NewReader(file)

	for {
		if line, err := reader.ReadString('\n'); err == nil {
			fmt.Println("Pin typed: ", line)

			if !isRegistered{
				fmt.Println("NOT NOT REGISTERED")
				//sending the public key X and Y to server
				requestBody, err := json.Marshal(Door{
					System: line,
					KeyX: privateKey.PublicKey.X,
					KeyY: privateKey.PublicKey.Y,
				})
				if err != nil{
					log.Fatalln(err)
				}

				//a post request that will send information to the server
				resp, err := http.Post("https://bast-security.xyz:8080/locks/register","",bytes.NewBuffer(requestBody))
				if err != nil{
					log.Fatalln(err)
				}

				defer resp.Body.Close()

				body, err := ioutil.ReadAll(resp.Body)
				if err != nil{
					log.Fatalln(err)
				}

				//prints out
				log.Println(string(body))

				//saves the unique lock id that the server sent as a response as the lock's unique id
				lockUniqueID = string(body)

				fmt.Println("xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxlockUniqueId: ", lockUniqueID)

				//CREATE A FILE NAMED lockId.txt and save the lockunique key
				f, err := os.Create("lockId.txt")
				if err != nil {
					fmt.Println(err)
				}

				_, err = io.WriteString(f, lockUniqueID)
				if err != nil{
					fmt.Println(err)
				}


				///CALL FUNCTION LOGIN HERE BBBB
				jwt, errr := loginFunction(lockUniqueID, privateKey)
				if jwt != "" && errr == nil {
					isRegistered = true
				}else{
					os.Exit(1)
				}

				//if true then is registered is true , else stays the same
				isRegistered = true
				//os.exit to exit program

			}
		} else if err != io.EOF {
			panic(err)
		}
	}
}

//function to login in lock-firmware to system
func loginFunction(lockUniqueID string, privateKey *ecdsa.PrivateKey) (string, error) {

	//get reuest to server to login in using the unique lock id
	resp, err := http.Get("https://bast-security.xyz:8080/locks/lockUniqueID")
	if err != nil{
		log.Fatalln(err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	//prints out
	fmt.Println(string(body))

	//saving the respond of the controller to the lock-firmware
	var challengeString = string(body)

	//hash the challenge string to return to controller
	var hashedChallengeString = sha256.Sum256([]byte(challengeString))
	fmt.Println("hashed challenge: ", hashedChallengeString)

	r,s, err := ecdsa.Sign(rand.Reader, privateKey, hashedChallengeString[:])
	if err != nil{
		panic(err)
	}

	//sending the public key X and Y to server
	requestBody, err := json.Marshal(RandS{
		R: r,
		S: s,
	})
	if err != nil{
		log.Fatalln(err)
	}

	//sending the challengestring
	var postRequest string = "https://bast-security.xyz:8080/locks/" + lockUniqueID + "/login"
	respP, err := http.Post(postRequest,"",bytes.NewBuffer(requestBody))
	if err != nil{
		log.Fatalln(err)
	}
	
	defer respP.Body.Close()

	if respP.StatusCode == 200 {
		bodyP, err := ioutil.ReadAll(resp.Body)
		if err != nil{
			log.Fatalln(err)
		}

		var jwt = string(bodyP)
		return jwt, nil
	}else{
		return "", fmt.Errorf("Failed to login")
	}

	
	//lock scramble it
	//use hash.sha256
	//create a signature using private key- using function sign? pash private key and hash, give you two numbers, r and s are signmature,
	// send those numbers to the server
	//crypto/rand
	//ecdsa.sign.random.reader
}