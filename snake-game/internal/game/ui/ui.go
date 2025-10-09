package ui

import (
	"bufio"
	"fmt"
	"os"
	"snake-game/internal/game/logic"
	proto "snake-game/internal/proto/gen"
	"strconv"
	"strings"
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

func (ui *ConsoleUI) ShowGameList(games []*proto.GameAnnouncement) {
	if len(games) == 0 {
		fmt.Println("\nNo available games found.")
		return
	}
	fmt.Println("\n=== AVAILABLE GAMES ===")
	for _, game := range games {
		fmt.Printf("%s (Players: %d)\n", game.GameName, len(game.Players.Players))
		fmt.Printf("   Field: %dx%d, Food: %d\n",
			game.Config.Width, game.Config.Height, game.Config.FoodStatic)
	}
}

func (ui *ConsoleUI) ReadJoinInfo() (string, proto.NodeRole) {
	fmt.Print("Enter game name to join: ")
	gameName := ui.readStringInput()

	role := ui.readPlayerRole()

	return gameName, role
}

func (ui *ConsoleUI) readPlayerRole() proto.NodeRole {
	for {
		fmt.Print("Enter mode (possible values are NORMAL and VIEWER): ")
		mode := strings.ToUpper(ui.readStringInput())

		switch mode {
		case "NORMAL":
			return proto.NodeRole_NORMAL
		case "VIEWER":
			return proto.NodeRole_VIEWER
		default:
			fmt.Println("Invalid input. Please enter either NORMAL or VIEWER.")
		}
	}
}

func (ui *ConsoleUI) ReadGameName() string {
	fmt.Print("Enter game name: ")
	if ui.scanner.Scan() {
		return strings.TrimSpace(ui.scanner.Text())
	}
	return "Default Game"
}

func (ui *ConsoleUI) ReadPlayerName() string {
	fmt.Print("Enter your nickname: ")
	if ui.scanner.Scan() {
		name := strings.TrimSpace(ui.scanner.Text())
		if name != "" {
			return name
		}
	}
	return "Player"
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

func (ui *ConsoleUI) PrintGameInfo(game *proto.GameAnnouncement) {
	fmt.Println("========GAMES========")
	fmt.Printf("GameName: %s\n", game.GameName)
}

func (ui *ConsoleUI) readIntInput(min, max int) int {
	for {
		if ui.scanner.Scan() {
			input := strings.TrimSpace(ui.scanner.Text())
			value, err := strconv.Atoi(input)
			if err == nil && value >= min && value <= max {
				return value
			}
		}
		fmt.Printf("Please enter a number between %d and %d: ", min, max)
	}
}

func (ui *ConsoleUI) readStringInput() string {
	if ui.scanner.Scan() {
		text := strings.TrimSpace(ui.scanner.Text())
		if text != "" {
			return text
		}
	}
	return ""
}

func ShowScores(logic *logic.GameLogic) {
	players := logic.GetPlayers().GetPlayers()
	if len(players) == 0 {
		return
	}

	fmt.Print("\033[H\033[2J")

	fmt.Println("=== SNAKE GAME SCORES ===")
	fmt.Println("Player Name          | Score | Status")
	fmt.Println("---------------------|-------|--------")

	for _, player := range players {
		status := "ALIVE"
		if snake := logic.GetSnakeByPlayerID(player.Id); snake != nil {
			if snake.State == proto.GameState_Snake_ZOMBIE {
				status = "ZOMBIE"
			}
		}

		role := ""
		switch player.Role {
		case proto.NodeRole_MASTER:
			role = " [MASTER]"
		case proto.NodeRole_DEPUTY:
			role = " [DEPUTY]"
		case proto.NodeRole_VIEWER:
			role = " [VIEWER]"
		}

		fmt.Printf("%-20s | %5d | %s%s\n", player.Name, player.Score, status, role)
	}
	fmt.Println("---------------------|-------|--------")
}
