package main

import "os"
import "fmt"
import "bufio"
import "strings"
import "strconv"

func main() {
    reader := bufio.NewReader(os.Stdin)

    state := State{}
    our_name := ""
    // their_name := ""

    for {
        line, err := reader.ReadString('\n')

        if err != nil {
            break
        }

        line = strings.TrimSpace(line)

        parts := strings.Split(line, " ")

        if len(line) == 0 {

        } else if strings.Index(line, "#") == 0 {

        } else if parts[0] == "settings" {
            if parts[1] == "your_bot" {
                our_name = parts[2]
                // fmt.Println(our_name)
            } else if parts[1] == "opponent_bot" {
                // their_name = parts[2]
            } else if parts[1] == "starting_armies" {
                state.starting_armies, _ = strconv.ParseInt(parts[2], 10, 0)
            }
        } else if parts[0] == "setup_map" {
            if parts[1] == "super_regions" {
                state.super_regions = make(map[int64]*SuperRegion)
                for i := 2; i < len(parts); i += 2 {
                    id, _ := strconv.ParseInt(parts[i], 10, 0)
                    reward, _ := strconv.ParseInt(parts[i+1], 10, 0)

                    state.super_regions[id] = &SuperRegion{id: id, reward: reward}
                }
            } else if parts[1] == "regions" {
                state.regions = make(map[int64]*Region)
                for i := 2; i < len(parts); i += 2 {
                    region_id, _ := strconv.ParseInt(parts[i], 10, 0)
                    super_region_id, _ := strconv.ParseInt(parts[i+1], 10, 0)

                    super_region := state.super_regions[super_region_id]
                    region := &Region{id: region_id, owner: "neutral", super_region: super_region, armies: 2}
                    state.regions[region_id] = region
                    super_region.regions = append(super_region.regions, region)
                }
            } else if parts[1] == "neighbors" {
                for i := 2; i < len(parts); i += 2 {
                    region_id, _ := strconv.ParseInt(parts[i], 10, 0)
                    neighbour_ids := strings.Split(parts[i+1], ",")

                    region := state.regions[region_id]
                    region.neighbours = []*Region{}

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
            for _, region := range state.regions {
                if region.owner == "us" {
                    region.owner = "them"
                }
            }

            for i := 1; i < len(parts); i += 3 {
                region_id, _ := strconv.ParseInt(parts[i], 10, 0)
                who := ""
                if parts[i+1] == our_name {
                    who = "us"
                } else {
                    who = "them"
                }
                // fmt.Fprintf(os.Stderr, "%d %s %s %s\n", region_id, who, parts[i+1], our_name)
                armies, _ := strconv.ParseInt(parts[i+2], 10, 0)

                region := state.regions[region_id]
                region.owner = who
                // fmt.Fprintf(os.Stderr, "%d %s vs %s\n", region_id, state.regions[region_id].owner, who)
                region.armies = armies
            }
        } else if parts[0] == "opponent_moves" {
            // TODO
        } else if parts[0] == "Round" {
            // ignore this
        } else if strings.Index(line, "Output from your bot: ") == 0 {

        } else if strings.Index(line, "go place_armies") == 0 {
            // fmt.Println("No moves")
            our_regions := ourRegions(state)
            region := our_regions[0]
            fmt.Printf("%s place_armies %d %d,\n", our_name, region.id, state.starting_armies)
            region.armies += state.starting_armies
        } else if strings.Index(line, "go attack/transfer") == 0 {
            var attack_from *Region
            var attack_to *Region
            attack_from = nil
            attack_to = nil
            for _, region := range ourRegions(state) {
                for _, neighbour := range region.neighbours {
                    // fmt.Fprintf(os.Stderr, "%d (%d - %s) -> %d (%d -%s)\n",
                        // region.id, region.armies, region.owner,
                        // neighbour.id, neighbour.armies, neighbour.owner)
                    if neighbour.owner != "us" {
                        if region.armies >= 5 + neighbour.armies {
                            if attack_to == nil || neighbour.armies > attack_to.armies {
                                attack_to = neighbour
                                attack_from = region
                            }
                        }
                    }
                }
            }

            if attack_from == nil {
                fmt.Println("No moves")
            } else {
                fmt.Printf("%s attack/transfer %d %d %d\n",
                    our_name, attack_from.id, attack_to.id, attack_from.armies - 1)
            }
        } else {
            fmt.Fprintf(os.Stderr, "Don't recognire: "+line+"\n")
        }

        // fmt.Printf("#regions: %d\n", len(state.regions))
    }
}

type State struct {
    regions map[int64]*Region
    super_regions map[int64]*SuperRegion
    starting_armies int64
}

func ourRegions(state State) []*Region {
    result := []*Region{}

    // fmt.Fprintf(os.Stderr, "ourRegions: %d\n", len(state.regions))
    for _, region := range state.regions {
        // fmt.Fprintf(os.Stderr, "%d %s\n", region.id, region.owner)
        if region.owner == "us" {
            result = append(result, region)
        }
    }

    return result
}

type Region struct {
    id int64
    super_region *SuperRegion
    neighbours []*Region
    owner string
    armies int64
}

type SuperRegion struct {
    id int64
    regions []*Region
    reward int64
}