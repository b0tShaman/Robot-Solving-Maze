package main

import (
	"container/heap"
	"encoding/csv"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	Obstacles = make(map[string]bool)
	Target    = Coordinate{9, 9}
)

const gridSize = 10

type Coordinate struct {
	x int
	y int
}

type Edge struct {
	c        Coordinate
	distance float64
}

type Neighbours []Edge

func (nbs Neighbours) Len() int           { return len(nbs) }
func (nbs Neighbours) Less(i, j int) bool { return nbs[i].distance < nbs[j].distance }
func (nbs Neighbours) Swap(i, j int)      { nbs[i], nbs[j] = nbs[j], nbs[i] }
func (nbs *Neighbours) Push(x interface{}) {
	*nbs = append(*nbs, x.(Edge))
}

func (nbs *Neighbours) Pop() interface{} {
	n := len(*nbs)
	old := *nbs
	x := old[n-1]
	*nbs = old[:n-1]
	return x
}

func (c Coordinate) isValid() bool {
	return c.x >= 0 && c.y >= 0 && !Obstacles[fmt.Sprintf("%d_%d", c.x, c.y)]
}

func printGrid(robotX, robotY int) {
	for y := 0; y < gridSize; y++ {
		for x := 0; x < gridSize; x++ {
			if x == robotX && y == robotY {
				fmt.Print("🤖") 
			} else if Obstacles[fmt.Sprintf("%d_%d", x, y)] {
				fmt.Print("🚫")
			} else if x == Target.x && y == Target.y {
				fmt.Print("🚪")
			} else {
				fmt.Print("⬛")
			}
		}
		fmt.Println()
	}
	fmt.Println()
}

func main() {
	filepath := "obstacle.csv"

	// Store obstacle coordinates in obstacle hashmap
	file, err := os.Open(filepath)
	if err != nil {
		log.Fatalln("Error opening file", err)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	data, err := reader.ReadAll()
	if err != nil {
		log.Fatalln("Error reading all data from file")
		return
	}

	adjMatrix := make(map[Coordinate][]Edge, 0)

	for _, row := range data[1:] {
		x, y := row[0], row[1]
		Obstacles[fmt.Sprintf("%s_%s", x, y)] = true
	}

	for y := 0; y < gridSize; y++ {
		for x := 0; x < gridSize; x++ {
			// get all 8 nearby edges
			// add valid ones(not negative + not in obstacle hash) to []Edge
			// adjMatrix[coordinate(x,y)] = []Edge
			coordinate := Coordinate{x, y}
			edge1 := Edge{Coordinate{x - 1, y - 1}, math.Sqrt(2)}
			edge2 := Edge{Coordinate{x - 1, y}, 1}
			edge3 := Edge{Coordinate{x - 1, y + 1}, math.Sqrt(2)}
			edge4 := Edge{Coordinate{x, y + 1}, 1}
			edge5 := Edge{Coordinate{x + 1, y + 1}, math.Sqrt(2)}
			edge6 := Edge{Coordinate{x + 1, y}, 1}
			edge7 := Edge{Coordinate{x + 1, y - 1}, math.Sqrt(2)}
			edge8 := Edge{Coordinate{x, y - 1}, 1}

			edges := make([]Edge, 0)
			if edge1.c.isValid() {
				edges = append(edges, edge1)
			}
			if edge2.c.isValid() {
				edges = append(edges, edge2)
			}
			if edge3.c.isValid() {
				edges = append(edges, edge3)
			}
			if edge4.c.isValid() {
				edges = append(edges, edge4)
			}
			if edge5.c.isValid() {
				edges = append(edges, edge5)
			}
			if edge6.c.isValid() {
				edges = append(edges, edge6)
			}
			if edge7.c.isValid() {
				edges = append(edges, edge7)
			}
			if edge8.c.isValid() {
				edges = append(edges, edge8)
			}
			adjMatrix[coordinate] = edges
		}
	}

	distances := make(map[Coordinate]float64) // distance of coordinate from origin (x=0, y=0)
	routes := make(map[Coordinate][]string)

	for k := range adjMatrix {
		distances[k] = math.MaxFloat64
	}

	// shortest distance of Coordinate{0, 0} from origin = 0
	distances[Coordinate{0, 0}] = 0

	// Min Heap queue to keep track of neighbouring coordinate at the shortest distance from current coordinate
	var minPriorityQueue Neighbours
	heap.Init(&minPriorityQueue)
	heap.Push(&minPriorityQueue, Edge{Coordinate{0, 0}, 0})

	for len(minPriorityQueue) > 0 {
		edge := heap.Pop(&minPriorityQueue).(Edge)
		currCoordinate, currDistance := edge.c, edge.distance

		// Loop through neighbours.
		// Check if, (distance between current coordinate and origin) + (distance between current coordinate and neighbouring coordinate) < shortest distance of neighbouring coordinate from origin
		// If yes, then current coordinate is a part of shortest route from origin to neighbouring coordinate
		for _, neighbour := range adjMatrix[currCoordinate] {
			newDist := currDistance + neighbour.distance
			if newDist < distances[neighbour.c] {
				distances[neighbour.c] = newDist
				// New shortest route to neighbour.c is found and it is via currCoordinate.
				// Reset route to neighbour coordinate and update new route as route till current coordinate + current coordinate
				routes[neighbour.c] = append(append([]string{}, routes[currCoordinate]...), fmt.Sprintf("%d_%d", currCoordinate.x, currCoordinate.y))
				// routes[neighbour.c] = append(routes[currCoordinate], fmt.Sprintf("%d_%d", currCoordinate.x, currCoordinate.y))

				heap.Push(&minPriorityQueue, Edge{neighbour.c, newDist})
				if neighbour.c == Target {
					fmt.Println("Reached target")
					minPriorityQueue = Neighbours{}
					break
				}
			}
		}
	}

	route, ok := routes[Target]
	if !ok{
		printGrid(0, 0)
		fmt.Println("NO PATH FOUND TO TARGET")
		return
	}

	route = append(route, fmt.Sprintf("%d_%d", Target.x, Target.y))

	for _, c := range route {
		parts := strings.Split(c, "_")

		robotX, _ := strconv.Atoi(parts[0])
		robotY, _ := strconv.Atoi(parts[1])

		fmt.Print("\033[H\033[2J") // Clear terminal
		printGrid(robotX, robotY)
		time.Sleep(500 * time.Millisecond)
	}
}
