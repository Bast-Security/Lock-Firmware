package main

import (
	"github.com/stianeikeland/go-rpio"
	"fmt"
)

func main() {

	//////////For Pin Pad//////////
	//		col 1	col 2	col 3
	// row 1	1	2	3
	// row 2	4	5	6
	// row 3	7	8	9
	// row 4	*	0	#

	//////////OUTPUTS//////////
	//GPIO10 -- row 1
	row1 :=  rpio.Pin(10)
	//GPIO3 -- row 2
	row2 := rpio.Pin(3)
	//GPIO4 -- row 3
	row3 := rpio.Pin(4)
	//GPIO27 -- row 4
	row4 := rpio.Pin(27)

	row1.Output()
	row2.Output()
	row3.Output()
	row4.Output()

	//////////INPUTS//////////
	//GPIO22 -- column 1
	column1 := rpio.Pin(22)
	//GPIO9 -- column 2
	column2 := rpio.Pin(9)
	//GPIO17 -- column 3
	column3 := rpio.Pin(17)

	column1.Input()
	column1.PullUp()
	column2.Input()
	column2.PullUp()
	column3.Input()
	column3.PullUp()

	//creating arrays of objects pins
	outputs := [4]rpio.Pin{row1, row2, row3, row4}
	inputs := [3]rpio.Pin{column1, column2, column3}

	//keyboard that will be used to output what number was pressed
	keyboardKey := [4][3]string{
		{"1", "2", "3"},
		{"4", "5", "6"},
		{"7", "8", "9"},
		{"*", "0", "#"},
	}

	wasPressed := [4][3]bool{
		{ false, false, false },
		{ false, false, false },
		{ false, false, false },
		{ false, false, false },
	}

	//variable will be saving the input of the user
	userPin := ""

	for {
		//for loop will loop through the outputs
		for row := 0; row < 4; row++ {

			//sets the current output to be 0
			outputs[row].Low()

			//for loop will loop through the inputs
			for column := 0; column < 3; column++ {

				//checks whether an input is 0
				if inputs[column].Read() == 0 && !wasPressed[row][column] {
					wasPressed[row][column] = true

					//if statment will recognize if # was pressed by user
					if keyboardKey[row][column] == "#" {
						//breaks out of loop
						fmt.Println(userPin)

						userPin = ""
						break
					}

					//saves the keypressed to userPin variable
					userPin = userPin + keyboardKey[row][column]
				} else if inputs[column].Read() != 0 {
					wasPressed[row][column] = false
				}
			}

			//closes the connection
			outputs[row].High()
		}
	}
}

