package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"time"
	"github.com/eiannone/keyboard"
)

const (
	NOTHING = 0
	WALL    = 1
	OBJECT = 2
	FINISH = 3
	PLAYER  = 69
	MAX_SAMPLES = 100
)

type position struct {
	x, y int
}

// INPUT
type input struct {
	pressedKey rune
}

func (i *input) update() {
	key, _, err := keyboard.GetKey()
	if err != nil {
		fmt.Println("Error reading key:", err)
		return
	}
	i.pressedKey = key
}
// END INPUT

// PLAYER
type player struct {
	pos position
	level *level
	input *input
}

func (p *player) update() {
	switch p.input.pressedKey {
	case 'w', 'W':
		if p.pos.y > 1 {
			p.pos.y--
		}
	case 's', 'S':
		if p.pos.y < p.level.height-2 {
			p.pos.y++
		}
	case 'a', 'A':
		if p.pos.x > 1 {
			p.pos.x--
		}
	case 'd', 'D':
		if p.pos.x < p.level.width-2 {
			p.pos.x++
		}
	}
}
// END PLAYER

// OBJECT
type object struct {
	pos position
	level *level
	reverse bool
}

func (o *object) updateUpDown() {
	if o.reverse {
		o.pos.y -= 1
		if o.pos.y == 1 {
			o.pos.y += 1
			o.reverse = false
		}
		return
	}
	o.pos.y += 1
	if o.pos.y == o.level.height-1 {
		o.pos.y -= 1
		o.reverse = true
	}
}

func (o *object) updateWaterFall() {
	o.pos.y += 1
	if o.pos.y == o.level.height-1 {
		o.pos.y = 1
	}
}
// END OBJECT

// FINISH
type finish struct {
	pos position
	level *level
}
// END FINISH

// STATS
type stats struct {
	start time.Time
	gameTime time.Duration
}

func newStats() *stats {
	return &stats{
		start: time.Now(),
	}
}

func (s *stats) update() {
	s.gameTime = time.Since(s.start)
}

func (g *game) renderStats() {
	g.drawBuf.WriteString("--Stats\n")
	g.drawBuf.WriteString(fmt.Sprintf("Time: %.2f\n", g.stats.gameTime.Seconds()))
}
// END STATS

// LEVEL
type level struct {
	width, height int
	data [][]int
}

func newLevel(width, height int) *level {
	data := make([][]int, height)
	for h := 0; h < height; h++ {
		for w := 0; w < width; w++{
			data[h] = make([]int, width)
		}
	}
	
	for h := 0; h < height; h++ {
		for w := 0; w < width; w++ {
			if h == 0 || w == 0 || h == height-1 || w == width-1 {
				data[h][w] = WALL
			}
		}
	}
	return &level{
		width: width,
		height: height,
		data: data,
	}
}

func (l *level) set(pos position, v int) {
	l.data[pos.y][pos.x] = v
}
// END LEVEL

// GAME
type game struct {
	isRunning bool
	gameEnd   int
	level     *level
	stats     *stats
	player    *player
	input     *input
	drawBuf   *bytes.Buffer
	object	  *object
	object2	  *object
	finish    *finish
}

func newGame(width, height int) *game {
	cmd := exec.Command("cmd", "/c", "cls")
	cmd.Stdout = os.Stdout
	cmd.Run()
	var (
		lvl = newLevel(width, height)
		inpu = &input{}
	)
	return &game{
		level: lvl,
		drawBuf: new(bytes.Buffer),
		stats: newStats(),
		input: inpu,
		player: &player{
			input: inpu,
			level: lvl,
			pos: position{x: 2, y: 5},
		},
		object: &object{
			level: lvl,
			pos: position{x: 16, y: 5},
		},
		object2: &object{
			level: lvl,
			pos: position{x: 10, y: 5},
		},
		finish: &finish{
			level: lvl,
			pos: position{x: 75, y: 10},
		},
	}
}

func (g *game) start() {
	g.isRunning = true
	g.loop()
}

func (g *game) loop() {
	for g.isRunning {
		g.input.update()
		g.update()
		g.render()
		g.stats.update()
	}
}

func (g *game) update() {
	g.level.set(g.object2.pos, NOTHING)
	g.object2.updateWaterFall()
	g.level.set(g.object2.pos, OBJECT)

	g.level.set(g.object.pos, NOTHING)
	g.object.updateUpDown()
	g.level.set(g.object.pos, OBJECT)

	g.level.set(g.finish.pos, FINISH)

	g.level.set(g.player.pos, NOTHING)
	g.player.update()
	g.level.set(g.player.pos, PLAYER)
		
	if g.object.pos == g.player.pos {
		g.gameEnd = 0
		g.isRunning = false
	}
	if g.finish.pos == g.player.pos {
		g.gameEnd = 1
		g.isRunning = false
	}
}

// Draws game
func (g *game) renderLevel() {
	for h := 0; h < g.level.height; h++ {
		for w := 0; w < g.level.width; w++ {
			if g.level.data[h][w] == NOTHING {
				g.drawBuf.WriteString(" ")
			} else if g.level.data[h][w] == WALL {
				g.drawBuf.WriteString("H")
			} else if g.level.data[h][w] == PLAYER {
				g.drawBuf.WriteString("P")
			} else if g.level.data[h][w] == OBJECT {
				g.drawBuf.WriteString("X")
			} else if g.level.data[h][w] == FINISH {
				g.drawBuf.WriteString("O")
			}
		}
		g.drawBuf.WriteString("\n")
	}
}

func (g *game) render() {
	g.drawBuf.Reset()
	fmt.Fprint(os.Stdout, "\033[2j\033[1;1H")
	g.renderLevel()
	g.renderInfo()
	g.renderStats()
	fmt.Fprint(os.Stdout, g.drawBuf.String())
}
// END GAME

func (g *game) renderInfo(){
	g.drawBuf.WriteString("Use WASD to move your character.\n")
}

func (g *game) renderFinalStats() {
	fmt.Printf("You completed this level in: %.2f seconds\n", g.stats.gameTime.Seconds())
}

func main() {
	err := keyboard.Open()
	if err != nil {
		fmt.Println("Error opening keyboard:", err)
		return
	}
	defer keyboard.Close()

	fmt.Println("Welcome to kamelKase!")
	fmt.Println("Would you like to start? (y/n): ")
	playAgain := true
	for playAgain{

		key, _, err := keyboard.GetKey()
		if err != nil {
			fmt.Println("Error reading key:", err)
			return
		}

		if key == 'y' || key == 'Y'{
				width := 80
				height := 18
				g := newGame(width, height)
				g.start()
				if g.gameEnd == 0 {
					fmt.Println("You colided with an 'X'! GAME OVER!")
				} else if g.gameEnd == 1{
					fmt.Println("You reached the Finish! LEVEL COMPLETE!")
					g.renderFinalStats()
				}
				fmt.Println("Would you like to Play Again? (y/n): ")
			}
		
		if key == 'n' || key == 'N' {
			playAgain = false
			fmt.Println("Exiting the game. Goodbye!")
			return
		} else if key != 'n' && key != 'N' && key != 'y' && key != 'Y' {
			fmt.Println("Invalid input. Please press 'y' to start or 'n' to exit.")
		}
	}	
}