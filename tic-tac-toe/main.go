package main

import (
	"fmt"
)

func override_placeholders(current_pawn string, placeholder *map[int]string, position int) int {
	if !(position >= 1 && position <= 9){
		fmt.Println("Invalid Position Input, try again !!!")
		return 0
	}
	if (*placeholder)[position] != "#" && (*placeholder)[position] != "@" {
		(*placeholder)[position] = current_pawn
		return 1
	} else {
		fmt.Println("Position already taken, try again !!!")
		return 0
	}
}

func check_status(placeholder *map[int]string) int {
	x := *placeholder
	if (x[1] == x[2] && x[2] == x[3]) || (x[4] == x[5] && x[5] == x[6]) || 
		(x[7] == x[8] && x[8] == x[9]) || 
		(x[1] == x[4] && x[4] == x[7]) || (x[2] == x[5] && x[5] == x[8]) || 
		(x[3] == x[6] && x[6] == x[9]) || 
		(x[1] == x[5] && x[5] == x[9]) || (x[3] == x[5] && x[5] == x[7]) {
		return 1
	}
	return 0

}

func layout(placeholders map[int]string) {
	double_dotted_line := "=============\n"
	single_dotted_line := "-------------\n"
	fmt.Printf("%s| %s | %s | %s |\n%s", double_dotted_line, placeholders[1], placeholders[2], placeholders[3], single_dotted_line)
	fmt.Printf("| %s | %s | %s |\n%s", placeholders[4], placeholders[5], placeholders[6], single_dotted_line)
	fmt.Printf("| %s | %s | %s |\n%s", placeholders[7], placeholders[8], placeholders[9], double_dotted_line)
}

func main() {
	placeholders := map[int]string{1: "1", 2: "2", 3: "3", 4: "4", 5: "5", 6: "6", 7: "7", 8: "8", 9: "9"}
	player := [2]string{}
	fmt.Print("Enter Player 1 Name: ")
	fmt.Scanf("%s\n", &player[0])
	fmt.Print("Enter Player 2 Name: ")
	fmt.Scanf("%s", &player[1])
	layout(placeholders)
	var position int
	var discard string
	var status int
	for status != 1 {
		for current_player := range player {
			current_pawn := "#"
			if player[current_player] == player[1] {
				current_pawn = "@"
			}
			flag := 0
			for flag == 0 {
				fmt.Printf("\n%s -> #, %s -> @\n", player[0], player[1])
				fmt.Scanln(&discard)
				fmt.Printf("%s_ Enter a number 1-9: ", player[current_player])
				fmt.Scanf("%d", &position)
				fmt.Print("\033[H\033[2J") // scrolls up previous content in the terminal.
				flag = override_placeholders(current_pawn, &placeholders, position)
				layout(placeholders)
				status = check_status(&placeholders)
				if status == 1 {
					fmt.Printf("\n%s won the game.", player[current_player])
					return
				}
			}
		}
	}
}
