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
)

func main() {

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	//lock's unique id 
	var lockUniqueID string
	//lock's private key
	var privateKey *ecdsa.PrivateKey
	var err error

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

		//sending the public key X and Y to server
		requestBody, err := json.Marshal(map[string]*big.Int{
			"X": privateKey.PublicKey.X,
			"Y": privateKey.PublicKey.Y,
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

	}else{
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

	}

	//LOGIN
	//////////////////////////////////////////////////////////////////////////////////////////
	//////////////////////////////////////////////////////////////////////////////////////////

	//get request to server to login
	/*
	respG, err := http.Get("https://bast-security.xyz:8080/locks/lockUniqueID")
	if err != nil{
		log.Fatalln(err)
	}

	defer respG.Body.Close()

	bodyG, err := ioutil.ReadAll(respG.Body)
	if err != nil{
		log.Fatalln(err)
	}

	//prints out
	log.Println(string(bodyG))
	*/


	//loop will loop constantly until forever and ever man
	for true{

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
				fmt.Println("key inserted: ", line)
			} else if err != io.EOF {
				panic(err)
			}
		}


		//SEND THIS INFORMATION 
	}
}
