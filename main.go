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

func main(){
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	/**----Variables----*/
	//lock unique id
	var lockUniqueID int64
	//lock private key
	var lockPrivateKey *ecdsa.PrivateKey
	//determines whether the lock is registerd to a system
	var isRegistered = false
	var err error

	/**checks to see if there is a file containing the the unique lock id named lockID.txt; if not then 
	lock is not registered; if exists then lock is registered*/
	lockIDfileInfo, err := os.Stat("lockID.txt")
	//lockID file does not exist
	if err!= nil{
		fmt.Println("\n---Lock is Un-Registered---\n")

		//generates an ecdsa key pair using the p384 curve for the lock
		lockPrivateKey, err = ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
		//incase of an error and ecdsa key pair is not generated
		if err != nil{
			panic(err)
		}

		//creates lockPrivateKey into data type byte in order to send a request to controller
		lockPrivateKeyBytes, err := x509.MarshalECPrivateKey(lockPrivateKey)
		//incase of an error that lockPrivateKey is not turned into data type bytes
		if err != nil{
			panic(err)
		}

		//Creates file named lockPrivateKey.pem to store lockPrivateKey
		pemFile, err := os.Create("lockPrivateKey.pem")
		//incase of an error and the file lockPrivateKey is not created, program exits
		if err != nil{
			fmt.Println("Pem File failed to create")
			os.Exit(1)
		}

		//encoding the lock private key
		var pemPrivateKeyBlock = &pem.Block{
			Type: "ECDSA Private Key",
			Bytes: lockPrivateKeyBytes,
		}

		//writes the encoded private key into lockPrivateKey.pem file
		err = pem.Encode(pemFile, pemPrivateKeyBlock)
		//incase the encoded private key does not write into the lockPrivateKey.pem file; program exits
		if err != nil{
			fmt.Println("Did not write into Pem File")
			os.Exit(1)
		}

		//closes the lockPrivateKey.pem file
		pemFile.Close()

	//lockID file exists	
	}else{
		fmt.Println("\n---Lock is Registered---\n")

		//lock is registered so isRegistered is set to true
		isRegistered = true

		//reads the lock id from the lockID.txt file
		lockIDfile, err := os.Open("lockID.txt")
		//incase the file does not open, program exits
		if err != nil{
			fmt.Println("Could not open lockID.txt")
			os.Exit(1)
		}

		//scanner is created to read from lockID.txt file
		scanner := bufio.NewScanner(lockIDfile)

		//variable holds data from the lockID.txt file when it is read
		var line string

		//for loop will read through each line from the lockID.txt
		for scanner.Scan(){
			line = scanner.Text()
		}

		//trims the string
		line = strings.TrimSpace(line)

		//converts string int an int64
		lockUniqueID, err = strconv.ParseInt(line,10,64)
		//incase string was not able to convert to int
		if err != nil{
			fmt.Println("Not able to convert to int")
			os.Exit(1)
		}

		//opens lockPrivateKey.pem file to get the private key of the lock
		pemFile, err := os.Open("lockPrivateKey.pem")
		//incase lockPrivateKey.pem file does not open/error
		if err != nil{
			fmt.Println("lockPrivateKey.pem could not open")
			os.Exit(1)
		}

		privatePublicKeyInfo, _ := pemFile.Stat()
		var size int64 = privatePublicKeyInfo.Size()
		pemBytes := make([]byte, size)

		buffer := bufio.NewReader(pemFile)
		stop, err := buffer.Read(pemBytes)
		fmt.Println("Stop: ", stop)

		data, _ := pem.Decode([]byte(pemBytes))

		pemFile.Close()

		//decode the data from the lockPrivateKey.pem file
		privatePublicKeyImported, err := x509.ParseECPrivateKey(data.Bytes)
		//incase data from the file is not decoded
		if err != nil{
			fmt.Println("Could not decode")
			os.Exit(1)
		}

		//saves the privatePublicKeyImported to lockPrivateKey
		lockPrivateKey = privatePublicKeyImported

		/*****************************************************/
		/************************LOGIN************************/
		/*****************************************************/
		jwt, err := login(lockUniqueID, lockPrivateKey)
		if jwt != "" && err == nil{
			isRegistered = true
			fmt.Println("HEYY")
		}else{
			os.Exit(1)
		}
	}

	//reads the argument that the user provides
	flag.Parse()

	//saves the path from the argument that the user provides
	var pathName = flag.Args()

	//converting pathName variable from []string to string
	file, err := os.OpenFile(strings.Join(pathName,""), os.O_RDONLY, os.ModeNamedPipe)
	//incase file does not open
	if err != nil{
		panic(err)
	}

	//reader will read the data inside the file
	reader := bufio.NewReader(file)
	
	/**for loop will continously loop and read a pin when entered in a terminal*/
	for{
		/*---------------------------------------------------------------------------------------------------*/
		/*----------------------ASK FABIO ABOUT THIS; reads system everytime first ran-----------------------*/
		/*---------------------------------------------------------------------------------------------------*/
		//if loop will read the data from the terminal
		if line, err := reader.ReadString('\n'); err == nil{
			fmt.Println("Pin typed: ", line)
			//removes the '\n\' from line
			line = strings.TrimSpace(line)

			//converting line from the terminal into an int64 data type
			systemID, err := strconv.ParseInt(line, 10, 64)
			//incase line does not successfully convert into an int64
			if err != nil{
				panic(err)
			}

			/**if look checks to see if lock is registered or not*/
			if !isRegistered{
				fmt.Println("\n---Lock is Un-Registered---\n")

				//creating a json string containing systemID and public x and y key
				requestBody, err := json.Marshal(Door{
					System: systemID,
					KeyX: lockPrivateKey.PublicKey.X,
					KeyY: lockPrivateKey.PublicKey.Y,
				})
				//incase requestBody could not be generated
				if err != nil{
					fmt.Println("requestBody could not be created")
					panic(err)
				}

				/*****************************************************/
				/**********************TESTING************************/
				fmt.Println("requestBody: ", string(requestBody))
				/*****************************************************/

				//post request will send information to the server
				registerResponse, err := http.Post("https://bast-security.xyz:8080/locks/register","",bytes.NewBuffer(requestBody))
				//incase registerResponse could not post
				if err != nil{
					panic(err)
				}

				defer registerResponse.Body.Close()

				//reads the response from the server
				registerResponseBody, err := ioutil.ReadAll(registerResponse.Body)
				//incase not able to read the response from the server
				if err != nil{
					panic(err)
				}

				/*****************************************************/
				/**********************TESTING************************/
				fmt.Println("registerResponseBody: ", string(registerResponseBody))
				/*****************************************************/

				//registerResponseBody is converted to a UniqueLockNumber in order to get the id of the lock the server sent
				registerResponseJSON := UniqueLockNumber{}
				json.Unmarshal([]byte(string(registerResponseBody)), &registerResponseJSON)
				//saves the id response to the lock lockUniqueID
				lockUniqueID = registerResponseJSON.Id

				//create a file named lockID.txt to store the unique lock id that the server sent to lock
				lockIDfile, err := os.Create("lockID.txt")
				//incase lockID.txt was not created
				if err != nil{
					fmt.Println("lockID.txt unable to be created")
					panic(err)
				}

				//writing into the lockID.txt file
				_, err = io.WriteString(lockIDfile, strconv.FormatInt(lockUniqueID, 10))
				//incase lockUniqueID was not written into the lockIDfile
				if err != nil{
					fmt.Println("Could not write to lockID.txt")
					panic(err)
				}

				/*****************************************************/
				/************************LOGIN************************/
				/*****************************************************/
				jwt, err := login(lockUniqueID, lockPrivateKey)
				if jwt != "" && err == nil{
					isRegistered = true
				}else{
					os.Exit(1)
				}

			}
		}else if err != io.EOF{
			panic(err)
		}
	}
	/*****************************************************/
	/**********************TESTING************************/
	fmt.Println("lockUniqueID: ", lockUniqueID)
	fmt.Println("lockPrivateKey: ", lockPrivateKey)
	fmt.Println("isRegistered: ", isRegistered)
	fmt.Println("lockIDfile: ", lockIDfileInfo)
	/*****************************************************/
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


	/*-------------------------------------------------------------------------------------*/
	fmt.Println("https://bast-security.xyz:8080/locks/" + strconv.FormatInt(lockUniqueID,10))


	defer loginResponse.Body.Close()

	//reads the response from the server
	loginResponseBody, err := ioutil.ReadAll(loginResponse.Body)
	//incase not able to read the response from the server
	if err != nil{
		panic(err)
	}

	/*****************************************************/
	/**********************TESTING************************/
	fmt.Println("loginResponseBody: ", string(loginResponseBody))
	/*****************************************************/

	//saving the login response body into a variable
	var challengeString = string(loginResponseBody)

	//hashing the challenge string
	var hashedChallengeString = sha256.Sum256([]byte(challengeString))
	
	/*****************************************************/
	/**********************TESTING************************/
	fmt.Println("hashedChallengeString: ", hashedChallengeString)
	/*****************************************************/

	//creats an R and S to send to server to confirm lock is who they say they are
	r, s, err := ecdsa.Sign(rand.Reader, lockPrivateKey, hashedChallengeString[:])
	//incase of an error
	if err != nil{
		panic(err)
	}

	//creating a json to send to server
	randSrequestBody, err := json.Marshal(RandS{
		R: r,
		S: s,
	})
	//incase of an error and json body is not created
	if err != nil{
		panic(err)
	}

	//sending the hashed challenge string to the server
	var challengeStringRequest string = "https://bast-security.xyz:8080/locks/" + strconv.FormatInt(lockUniqueID,10) + "/login"
	challengeStringResponse, err := http.Post(challengeStringRequest, "", bytes.NewBuffer(randSrequestBody))
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