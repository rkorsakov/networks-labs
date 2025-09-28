package ui

import (
	"bufio"
	"fmt"
	"os"
	proto "snake-game/internal/proto/gen"
	"strconv"
)

type ConsoleUI struct {
	scanner *bufio.Scanner
}

func NewConsoleUI() *ConsoleUI {
	return &ConsoleUI{
		scanner: bufio.NewScanner(os.Stdin),
	}
}

type MenuOption int

const (
	StartNewGame MenuOption = iota + 1
	JoinGame
	ShowGames
	Exit
)

func (ui *ConsoleUI) ShowMainMenu() MenuOption {
	fmt.Println("\n=== SNAKE MULTIPLAYER ===")
	fmt.Println("1. Start new game")
	fmt.Println("2. Join existing game")
	fmt.Println("3. Show available games")
	fmt.Println("4. Exit")
	fmt.Print("Choose option: ")

	return MenuOption(ui.readIntInput(1, 4))
}

func (ui *ConsoleUI) ShowGameList(games []*proto.GameAnnouncement) (int, bool) {
	if len(games) == 0 {
		fmt.Println("\nNo available games found.")
		fmt.Println("1. Back to main menu")
		fmt.Print("Choose option: ")
		choice := ui.readIntInput(1, 1)
		return 0, choice == 1
	}
	fmt.Println("\n=== AVAILABLE GAMES ===")
	for i, game := range games {
		fmt.Printf("%d. Game (Players: %d)\n", i+1, len(game.Players.Players))
		fmt.Printf("   Field: %dx%d, Food: %d\n",
			game.Config.Width, game.Config.Height, game.Config.FoodStatic)
	}
	fmt.Printf("%d. Back to main menu\n", len(games)+1)
	fmt.Print("Choose game to join: ")

	choice := ui.readIntInput(1, len(games)+1)
	if choice == len(games)+1 {
		return 0, true
	}
	return choice - 1, false
}

func (ui *ConsoleUI) ShowConnectionInfo(role proto.NodeRole) {
	var roleStr string
	switch role {
	case proto.NodeRole_MASTER:
		roleStr = "MASTER"
	case proto.NodeRole_DEPUTY:
		roleStr = "DEPUTY"
	case proto.NodeRole_NORMAL:
		roleStr = "PLAYER"
	case proto.NodeRole_VIEWER:
		roleStr = "VIEWER"
	}
	fmt.Printf("Connected as %s. Starting game...\n", roleStr)
}

func (ui *ConsoleUI) ShowErrorMessage(message string) {
	fmt.Printf("Error: %s\n", message)
}

func (ui *ConsoleUI) ShowInfoMessage(message string) {
	fmt.Println(message)
}

func (ui *ConsoleUI) readIntInput(min, max int) int {
	for {
		if ui.scanner.Scan() {
			input := ui.scanner.Text()
			value, err := strconv.Atoi(input)
			if err == nil && value >= min && value <= max {
				return value
			}
		}
		fmt.Printf("Please enter a number between %d and %d: ", min, max)
	}
}
