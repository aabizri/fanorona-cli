// Main file

// fanorona-cli
//
// This is a small projet to test & grow with the fanorona/basic library, however it should work well enough
//
// How to play:
//
//  Start a game by doing
//       fanorona-cli print
//
//  Then move
//       fanorona-cli move 4,3 East y
//
//  The coordinates are the "4,3", the direction is "East", while "y" indicates that the elimination ray should be in the same direction than the move direction
//
// Currently chaining the moves is still impossible
//
// Licensed under UNLICENSE
//
// by @nodvos <alexandre@bizri.fr>
package main

import (
	"github.com/nodvos/go-fanorona"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
)

const filename string = "fanorona.save"

var (
	Board *basic.Board
	Turns uint
)

// Checks whether the input is a direction
func isDir(input string) bool {
	rg, err := regexp.Compile("(((N|n)orth|(S|s)outh))?((W|w)est|(E|e)ast)?")
	if err != nil {
		panic(err)
	}

	return rg.MatchString(input)
}

// Check whether the input is a coordinates
func isCoordinates(input string) bool {
	rg, err := regexp.Compile("[1-9],[1-5]")
	if err != nil {
		panic(err)
	}

	return rg.MatchString(input)
}

// Store the info
func save_file(str string) error {
	err := ioutil.WriteFile(filename, []byte(str), 0666)
	return err
}

// Wrapper for serialize + save_file
func save() error {
	str, err := serialise(Board)
	if err != nil {
		return err
	}
	err = save_file(str)
	return err
}

// Wrapper for load_file + parse
func load() (*basic.Board, uint, error) {
	str, err := load_file()
	if err != nil {
		return &basic.Board{}, 0, err
	}
	b, t, err := parse(str)
	return b, t, err
}

// Loads the save files
func load_file() (string, error) {
	raw_data, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}
	str_data := string(raw_data[:])
	return str_data, err
}

/*
	Parsing the save data
	Exemple: Turn 1283, lines delimited by _
		1283_00111_00011_00111_00011_00-11_00111_00011_00111_00011


	Regex string with lookahead/lookbehind: (^[1-9][0-9]*(?=_))|((?!_)(-|0|1){5}(?=(_|$)))
	without: (^[1-9][0-9]*)|((-|0|1){5})
*/
func parse(str string) (*basic.Board, uint, error) {

	// Parse each part
	rg, err := regexp.Compile("(^[1-9][0-9]*)|((-|0|1){5})") // THIS IS THE PART THATS FUCKED UP
	if err != nil {
		return &basic.Board{}, 0, err
	}

	// Small sanity check
	indexes_parts := rg.FindAllStringIndex(str, 10)
	if len(indexes_parts) != 10 {
		return &basic.Board{}, 0, errors.New("No 10 parts of save file")
	}

	// Cut it up
	parts := rg.FindAllString(str, 10)
	turns_raw, err := strconv.Atoi(parts[0])
	turns := uint(turns_raw)
	h_lines := parts[1:10]

	// Manage the lines
	b := basic.NewBoard() // A new board
	for i, k := range h_lines {
		b[i] = parse_vline(k, uint(i), b)
	}

	// Return the result
	return b, turns, err
}

// Parse a vertical line
func parse_vline(str string, v uint, b *basic.Board) [5]*basic.Slot {
	var v_line [basic.Vertical]*basic.Slot
	var slot *basic.Slot
	for i, k := range str {
		switch k {
		case '-': // Empty slot
			slot = basic.SetupSlot(v, uint(i), b)
		case '0': // White slot
			slot = basic.SetupSlot(v, uint(i), b)
			piece := &basic.Piece{false, slot}
			slot.Populate(piece)
		case '1': // Black slot
			slot = basic.SetupSlot(v, uint(i), b)
			piece := &basic.Piece{true, slot}
			slot.Populate(piece)
		}
		v_line[i] = slot
	}
	return v_line
}

// Serialise the data
func serialise(b *basic.Board) (string, error) {
	str := strconv.Itoa(int(Turns))
	for _, k := range b {
		str += "_"
		hline_str, err := serialise_vline(k)
		if err != nil {
			return "", err
		}
		str += hline_str
	}
	return str, nil
}

func serialise_vline(vline [5]*basic.Slot) (string, error) {
	str := ""
	for _, k := range vline {
		if k.Piece == nil {
			str += "-"
		} else {
			if k.Piece.Black {
				str += "1"
			} else {
				str += "0"
			}
		}
	}
	return str, nil
}

/*
func new_game() error {
  board := basic.SetupBoard()
}*/

func print_cell(t int) string {
	color := ""
	switch t {
	case -1:
		color = "\033[8m " // Hidden
	case 0:
		color = "\033[34m■" // Blue
	case 1:
		color = "\033[31m■" // Red
	}

	return "[" + color + "\033[0m]"
}

const (
	//horizontal_label_alpha string= "   A  B  C  D  E  F  G  H  I \n"
	horizontal_label_num string = "   1  2  3  4  5  6  7  8  9 \n"
)

func print() error {

	player := ""
	if isBlackTurn() {
		player = "black"
	} else {
		player = "white"
	}
	fmt.Printf("This is turn %d: %s's turn's \n", Turns, player)

	result := ""
	for v := basic.Vertical - 1; v >= 0; v-- {
		h_line := fmt.Sprintf("%d ", v+1)
		for h := uint(0); h < basic.Horizontal; h++ {
			if Board[h][v].Piece != nil {
				if !Board[h][v].Piece.Black {
					h_line += print_cell(0)
				} else {
					h_line += print_cell(1)
				}
			} else {
				h_line += print_cell(-1)
			}
		}
		result += h_line + "\n"
		if v == 0 {
			break
		}
	}
	fmt.Print(result + horizontal_label_num)
	return nil
}

func parseCoordinates(str string) (uint, uint, error) {

	if !isCoordinates(str) {
		return 0, 0, errors.New("Coordinates provided aren't in the good format")
	}

	rg, err := regexp.Compile("[1-9]")
	if err != nil {
		return 0, 0, err
	}
	both := rg.FindAllString(str, 3) // 3 to allow detection of corrupted format
	if len(both) != 2 {
		return 0, 0, errors.New("Coordinates provided aren't in the good format")
	}
	h, err := strconv.Atoi(both[0])
	v, err := strconv.Atoi(both[1])

	return uint(h - 1), uint(v - 1), err
}

func parseDirection(str string) (basic.Offset, error) {

	if !isDir(str) {
		return basic.Offset{}, errors.New("Coordinates provided aren't in the good format")
	}

	for i, k := range basic.Directions {
		if strings.ToLower(str) == strings.ToLower(i) {
			return k, nil
		}
	}
	return basic.Offset{}, errors.New("Couldn't find a corresponding direction event though filter sais input was good")
}

func parseSameDirection(str string) (bool, error) {
	m_y, err := regexp.MatchString("(((y|Y)(es)?)|((t|T)rue))", str)

	if err != nil {
		return false, err
	}
	if !m_y {
		m_n, err := regexp.MatchString("((((n|N)(o)?)|((f|F)(alse)?)))", str)
		if err != nil {
			return false, err
		}
		if m_n {
			return false, nil
		} else {
			return false, errors.New("Neither, wtf was input?")
		}
	} else {
		return true, nil
	}

	return false, errors.New("parseSameDirection: Ended without result")
}

func isBlackTurn() bool {
	return Turns%2 == 0
}

func move(args []string) error {
	if len(args) != 3 {
		return errors.New("Not enough arguments for move")
	}

	h, v, err := parseCoordinates(args[0])

	if Board[h][v].Piece == nil {
		return errors.New("No piece there")
	}

	if Board[h][v].Piece.Black != isBlackTurn() {
		return errors.New("Can't play this piece")
	}

	direction, err := parseDirection(args[1])
	same, err := parseSameDirection(args[2])

	if !Board[h][v].Piece.CanMove(direction) {
		return errors.New("Can't move")
	}

	err = Board[h][v].Piece.MovEval(direction, same)
	if err != nil {
		return err
	}
	print()
	Turns += 1
	return err
}

func main() {

	// Load
	var (
		b *basic.Board
		t uint
	)
	b, t, err := load()
	if err != nil && !os.IsNotExist(err) {
		fmt.Printf("Error while loading file:\n\t %s", err)
		panic(err)
	}
	if os.IsNotExist(err) {
		b = basic.SetupBoard()
		t = 1
	}
	Board = b
	Turns = t

  // Check for win
  win,black := b.Win()
  if win {
    s := ""
    switch black{
      case true:
        s = "black"
      case false:
        s = "white"
    }
    fmt.Printf("The %s player has won\n",s)
  }
  
	// Now read
	if len(os.Args) == 1 {
		fmt.Println("Please give arguments")
		return
	}
  
	switch os.Args[1] {
	case "print":
		err = print()
	case "move":
		err = move(os.Args[2:])
		if err != nil {
			fmt.Println(err.Error())
		}
	}

	// Save
	err = save()
	if err != nil {
		fmt.Printf("Received error while saving: %s", err.Error())
	}
}
