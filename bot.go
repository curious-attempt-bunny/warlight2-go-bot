package main

import "os"
import "fmt"
import "bufio"
import "strings"
import "strconv"
import "math"

func main() {
    reader := bufio.NewReader(os.Stdin)

    state := State{}
    our_name := ""
    last_placements := []Placement{}
    // their_name := ""

    for {
        line, err := reader.ReadString('\n')

        if err != nil {
            break
        }

        line = strings.TrimSpace(line)

        // fmt.Fprintf(os.Stderr, ">> "+line+"\n")

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
                    region := &Region{
                        id: region_id,
                        owner: "neutral",
                        super_region: super_region,
                        armies: 2,
                        neighbours: []*Region{} }
                    state.regions[region_id] = region
                    super_region.regions = append(super_region.regions, region)
                }
            } else if parts[1] == "neighbors" {
                for i := 2; i < len(parts); i += 2 {
                    region_id, _ := strconv.ParseInt(parts[i], 10, 0)
                    neighbour_ids := strings.Split(parts[i+1], ",")

                    region := state.regions[region_id]

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
            var selected *Region
            selected = nil
            var selected_value float64
            selected_value = 0

            for i := 2; i < len(parts); i++ {
                region_id, _ := strconv.ParseInt(parts[i], 10, 0)
                region := state.regions[region_id]
                var armies int64
                armies = -2 // account for the one that is claimed
                for _, r := range region.super_region.regions {
                    armies += r.armies
                }
                // 2 armies -> 1
                // 4 armies -> 2
                // 6 armies -> 4
                // 8 armies -> 5
                weighting := math.Floor(float64(armies)/2.0)
                if armies >= 6 {
                    weighting += 1
                }
                value := float64(region.super_region.reward) / weighting

                if selected == nil || value > selected_value {
                    selected = region
                    selected_value = value
                }
            }

            fmt.Printf("%d\n", selected.id)
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
                } else if parts[i+1] == "neutral" {
                    who = "neutral"
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
            command := line[len("Output from your bot: ")+1:len(line)-1]
            // fmt.Fprintf(os.Stderr, "Output: "+command+"\n")
            parts = strings.Split(command, " ")
            if len(parts) > 2 && parts[1] == "place_armies" {
                for _, placement := range last_placements {
                    // fmt.Fprintf(os.Stderr, "Removing calculated placement at %d of %d armies\n",
                        // placement.region.id, placement.armies)
                    state.regions[placement.region.id].armies -= placement.armies
                }

                for i := 0; i < len(parts); i += 4 {
                    if parts[i] != our_name || parts[i+1] != "place_armies" {
                        break
                    }
                    region_id, _ := strconv.ParseInt(parts[i+2], 10, 0)
                    armies, _ := strconv.ParseInt(parts[i+3][0:len(parts[i+3])-1], 10, 0)

                    state.regions[region_id].armies += armies

                    // fmt.Fprintf(os.Stderr, "Adding recorded placement at %d of %d armies\n",
                    //     region_id, armies)
                }
            }
        } else if strings.Index(line, "go place_armies") == 0 {
            // fmt.Println("No moves")
            border_owner := "them"
            our_regions := ourBorderRegionsWithTheEnemy(state)
            if len(our_regions) == 0 {
                our_regions = ourBorderRegions(state)
                border_owner = "neutral"
            }

            var region *Region
            region = nil
            bordered_enemies := 0

            for _, border_region := range our_regions {
                bordered := 0
                for _, neighbour := range border_region.neighbours {
                    if neighbour.owner == border_owner {
                        bordered += 1
                    }
                }

                if region == nil || bordered > bordered_enemies ||
                    (bordered == bordered_enemies && border_region.id > region.id) { // repeatability
                    region = border_region
                    bordered_enemies = bordered
                }
            }

            fmt.Printf("%s place_armies %d %d,\n", our_name, region.id, state.starting_armies)
            region.armies += state.starting_armies
            last_placements = []Placement{}
            last_placements = append(last_placements, Placement{region: region, armies: state.starting_armies})
            // fmt.Fprintf(os.Stderr, "Adding calculated placement at %d of %d armies\n",
            //     last_placements[0].region.id, last_placements[0].armies)

        } else if strings.Index(line, "go attack/transfer") == 0 {
            // fmt.Fprintf(os.Stderr, "Attacks:\n")
            attacks := []Attack{}

            // --------- reinforcements

            move_to := ourBorderRegions(state)
            considered := make(map[int64]bool)
            for _, region := range move_to {
                considered[region.id] = true
            }

            for {
                if len(move_to) == 0 {
                    break
                }

                region := move_to[0]
                move_to = move_to[1:]

                for _, neighbour := range region.neighbours {
                    if neighbour.owner != "us" {
                        continue
                    }

                    if considered[neighbour.id] {
                        continue
                    }

                    if neighbour.armies > 1 {
                        attacks = append(attacks, Attack{
                            region_from: neighbour,
                            region_to: region,
                            armies: neighbour.armies - 1})
                    }

                    considered[neighbour.id] = true
                    move_to = append(move_to, neighbour)
                }
            }

            // --------- attacks

            used := make(map[int64]bool)
            for {
                var attack_from *Region
                var attack_to *Region
                attack_from = nil
                attack_to = nil
                value := float64(0)
                for _, region := range ourRegions(state) {
                    if used[region.id] {
                        continue;
                    }
                    bordering_enemy := false
                    largest_enemy := int64(0)
                    for _, neighbour := range region.neighbours {
                        if neighbour.owner == "them" {
                            bordering_enemy = true
                            if neighbour.armies > largest_enemy {
                                largest_enemy = neighbour.armies
                            }
                        }
                    }

                    for _, neighbour := range region.neighbours {
                        if used[neighbour.id] {
                            continue;
                        }
                        if neighbour.owner != "us" {
                            candidate_value := float64(neighbour.armies)
                            if neighbour.owner == "neutral" {
                                enemy_armies_in_super_region := int64(0)
                                neutral_armies_in_super_region := int64(0)

                                for _, subregion := range neighbour.super_region.regions {
                                    if subregion.owner == "them" {
                                        enemy_armies_in_super_region += subregion.armies
                                    } else if subregion.owner == "neutral" {
                                        neutral_armies_in_super_region += subregion.armies
                                    }
                                }

                                candidate_value = float64(neighbour.super_region.reward)

                                cost := float64(neutral_armies_in_super_region + 3*enemy_armies_in_super_region)
                                if cost != 0 { // can cost ever really by zero !?
                                    candidate_value = candidate_value / cost
                                }
                            }

                            attackable := region.armies >= 5 + neighbour.armies
                            if neighbour.owner == "neutral" {
                                attackable = (region.armies >= 3 && neighbour.armies == 1) ||
                                    (region.armies > 2 && region.armies >= 2*neighbour.armies)
                                if bordering_enemy {
                                    attackable = false // never attack neutral when enemy is present
                                }
                            } else if neighbour.owner == "them" {
                                if neighbour.armies < largest_enemy {
                                    attackable = false // only attack the largest neighbouring army group
                                }
                            }

                            // fmt.Fprintf(os.Stderr, "%d (%d - %s) -> %d (%d - %s) attackable %t value %d best_value %d\n",
                            //     region.id, region.armies, region.owner,
                            //     neighbour.id, neighbour.armies, neighbour.owner,
                            //     attackable, candidate_value, value)
                            if attackable {
                                if attack_to == nil ||
                                    (candidate_value >= value) ||
                                    (neighbour.armies == attack_to.armies && region.armies > attack_from.armies) {
                                    attack_to = neighbour
                                    attack_from = region
                                    value = candidate_value
                                }
                            }
                        }
                    }
                }

                if attack_from == nil {
                    break
                } else {
                    attacks = append(attacks, Attack{
                        region_from: attack_from,
                        region_to: attack_to,
                        armies: attack_from.armies-1})
                    used[attack_from.id] = true
                    used[attack_to.id] = true
                }
            }

            if len(attacks) == 0 {
                fmt.Println("No moves")
            } else {
                for _, attack := range attacks {
                    fmt.Printf("%s attack/transfer %d %d %d,",
                        our_name, attack.region_from.id, attack.region_to.id, attack.armies)
                }
                fmt.Println()
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

func ourBorderRegions(state State) []*Region {
    result := []*Region{}

    // fmt.Fprintf(os.Stderr, "ourRegions: %d\n", len(state.regions))
    for _, region := range state.regions {
        // fmt.Fprintf(os.Stderr, "%d %s\n", region.id, region.owner)
        if region.owner == "us" {
            borders := false
            for _, neighbour := range region.neighbours {
                if neighbour.owner != "us" {
                    borders = true
                    break
                }
            }

            if borders {
                result = append(result, region)
            }
        }
    }

    return result
}

func ourBorderRegionsWithTheEnemy(state State) []*Region {
    result := []*Region{}

    // fmt.Fprintf(os.Stderr, "ourRegions: %d\n", len(state.regions))
    for _, region := range state.regions {
        // fmt.Fprintf(os.Stderr, "%d %s\n", region.id, region.owner)
        if region.owner == "us" {
            borders := false
            for _, neighbour := range region.neighbours {
                if neighbour.owner == "them" {
                    borders = true
                    break
                }
            }

            if borders {
                result = append(result, region)
            }
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

type Placement struct {
    region *Region
    armies int64
}

type Attack struct {
    region_from *Region
    region_to *Region
    armies int64
}