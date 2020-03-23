package main

import (
	"fmt"
	"flag"
	"os"
	"strings"
	"bufio"
)

func main() {

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

		//if file exists then the file is read using bufio library
		scanner := bufio.NewScanner(file)

		//prints out what was added to the pipe file
		for scanner.Scan(){
			fmt.Println(scanner.Text())
		}
	}
}
