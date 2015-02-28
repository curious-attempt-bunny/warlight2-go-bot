package main

import "os"
import "fmt"
import "bufio"
import "strings"
import "strconv"

func main() {
    reader := bufio.NewReader(os.Stdin)

    for {
        line, err := reader.ReadString('\n')
        state := State{}

        // our_name := ""
        // their_name := ""

        if err != nil {
            break
        }

        line = strings.TrimSpace(line)

        parts := strings.Split(line, " ")

        if len(line) == 0 {

        } else if strings.Index(line, "#") == 0 {

        } else if parts[0] == "settings" {
            if parts[1] == "your_bot" {
                // our_name = parts[2]
            } else if parts[1] == "opponent_bot" {
                // their_name = parts[2]
            }
        } else if parts[0] == "setup_map" {
            if parts[1] == "super_regions" {
                state.super_regions = make(map[int64]SuperRegion)
                for i := 2; i < len(parts); i += 2 {
                    id, _ := strconv.ParseInt(parts[i], 10, 0)
                    reward, _ := strconv.ParseInt(parts[i+1], 10, 0)

                    state.super_regions[id] = SuperRegion{id: id, reward: reward}
                }
            } else if parts[1] == "regions" {
                state.regions = make(map[int64]Region)
                for i := 2; i < len(parts); i += 2 {
                    region_id, _ := strconv.ParseInt(parts[i], 10, 0)
                    super_region_id, _ := strconv.ParseInt(parts[i+1], 10, 0)

                    super_region := state.super_regions[super_region_id]
                    region := Region{id: region_id, owner: "neutral", super_region: super_region, armies: 2}
                    state.regions[region_id] = region
                    super_region.regions = append(super_region.regions, region)
                }
            } else if parts[1] == "neighbors" {
                for i := 2; i < len(parts); i += 2 {
                    region_id, _ := strconv.ParseInt(parts[i], 10, 0)
                    neighbour_ids := strings.Split(parts[i+1], ",")

                    region := state.regions[region_id]
                    region.neighbours = []Region{}

                    for _, neighbour_id := range neighbour_ids {
                        id, _ := strconv.ParseInt(neighbour_id, 10, 0)
                        neighbour := state.regions[id]
                        region.neighbours = append(region.neighbours, neighbour)
                        neighbour.neighbours = append(neighbour.neighbours, region)
                    }
                }
            } else if parts[1] == "wastelands" {
                for i := 2; i < len(parts); i++ {
                    region_id, _ := strconv.ParseInt(parts[i], 10, 0)

                    region := state.regions[region_id]
                    region.armies = 6
                }
            }
        } else if parts[0] == "pick_starting_region" {
            // pick the first one
            fmt.Println(parts[2])
        } else if parts[0] == "update_map" {

        } else if parts[0] == "opponent_moves" {

        } else if parts[0] == "Round" {

        } else if strings.Index(line, "Output from your bot: ") == 0 {

        } else if strings.Index(line, "go place_armies") == 0 {
            fmt.Println("No moves")
        } else if strings.Index(line, "go attack/transfer") == 0 {
            fmt.Println("No moves")
        } else {
            fmt.Fprintf(os.Stderr, "Don't recognire: "+line+"\n")
        }
    }
}

type State struct {
    regions map[int64]Region
    super_regions map[int64]SuperRegion
}

type Region struct {
    id int64
    super_region SuperRegion
    neighbours []Region
    owner string
    armies int64
}

type SuperRegion struct {
    id int64
    regions []Region
    reward int64
}