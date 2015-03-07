package main

import "os"
import "fmt"
import "bufio"
import "strings"
import "strconv"
import "math"

func main() {
    reader := bufio.NewReader(os.Stdin)

    state := &State{}
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
            } else if parts[1] == "opponent_starting_regions" {
                for i := 2; i < len(parts); i++ {
                    region_id, _ := strconv.ParseInt(parts[i], 10, 0)

                    region := state.regions[region_id]
                    region.owner = "them"
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

                neighbours_in_super_region := 0
                for _, neighbour := range region.neighbours {
                    if neighbour.super_region.id == region.super_region.id {
                        neighbours_in_super_region += 1
                    }
                }

                super_region_neighbour_ratio := float64(1+neighbours_in_super_region) / float64(len(region.super_region.regions))

                value := super_region_neighbour_ratio*float64(region.super_region.reward) / weighting

                // fmt.Fprintf(os.Stderr, "Starting region %d because score %g (reward %d armies %d weighting %g ratio %g score %d / %g)\n",
                //     region.id, value,
                //     region.super_region.reward, armies, weighting, super_region_neighbour_ratio,
                //     region.super_region.reward, weighting)
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
                region.visible = false
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
                region.visible = true
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
            // fmt.Fprintln(os.Stderr, "Place armies --")
            border_owner := "them"
            our_regions := ourBorderRegionsWithTheEnemy(state)

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

                    // fmt.Fprintf(os.Stderr, "Like border region %d (bordered_enemies %d)\n", region.id, bordered_enemies)
                }
            }

            if region == nil {
                best_score := float64(0)
                for _, border_region := range ourBorderRegionsWithNeutralOnly(state) {
                    // TODO - not DRY
                    enemy_armies_in_super_region := int64(0)
                    neutral_armies_in_super_region := int64(0)

                    for _, subregion := range border_region.super_region.regions {
                        if subregion.owner == "them" {
                            enemy_armies_in_super_region += subregion.armies
                        } else if subregion.owner == "neutral" {
                            neutral_armies_in_super_region += subregion.armies
                        }
                    }

                    if neutral_armies_in_super_region == 0 && enemy_armies_in_super_region == 0 {
                        continue
                    }

                    enemy_neighbours_in_super_region := 0
                    for _, neighbour := range border_region.neighbours {
                        // fmt.Fprintf(os.Stderr, "  neighbour.super_region.id (%d) == border_region.super_region.id (%d) && neighbour.owner (%s) != \"us\"\n",
                        //     neighbour.super_region.id, border_region.super_region.id,
                        //     neighbour.owner)
                        if neighbour.super_region.id == border_region.super_region.id &&
                            neighbour.owner != "us" {
                            enemy_neighbours_in_super_region = enemy_neighbours_in_super_region + 1
                        }
                    }

                    score := float64(border_region.super_region.reward)

                    cost := float64(neutral_armies_in_super_region + 3*enemy_armies_in_super_region)
                    cost = math.Pow(cost, 1.01) // inflate larger costs
                    score = score / cost

                    score += float64(enemy_neighbours_in_super_region)*0.001 // prefer neighbouring more to attack

                    // fmt.Fprintf(os.Stderr, "Place armies at %d has score %g (%d / %d+3*%d) - super region %d (enemy neighbours in super region %d)\n",
                    //     border_region.id, score,
                    //     border_region.super_region.reward,
                    //     neutral_armies_in_super_region, enemy_armies_in_super_region,
                    //     border_region.super_region.id,
                    //     enemy_neighbours_in_super_region)
                    if region == nil || score > best_score {
                        region = border_region
                        best_score = score
                    }
                }
            }

            if region == nil {
                region = ourBorderRegions(state)[0]
            }

            fmt.Printf("%s place_armies %d %d,\n", our_name, region.id, state.starting_armies)
            region.armies += state.starting_armies
            last_placements = []Placement{}
            last_placements = append(last_placements, Placement{region: region, armies: state.starting_armies})
            // fmt.Fprintf(os.Stderr, "Adding calculated placement at %d of %d armies\n",
            //     last_placements[0].region.id, last_placements[0].armies)

        } else if strings.Index(line, "go attack/transfer") == 0 {
            attacks := calculateAttacks(state)

            // consolidate attacks & transfers

            regionIdToIndex := make(map[int64]int64)
            i := int64(0)
            for {
                if (int(i) >= len(attacks)) {
                    break
                }

                key := attacks[i].region_from.id * 1000 + attacks[i].region_to.id
                index, exists := regionIdToIndex[key]
                if exists {
                    attack := attacks[i]

                    new_attacks := make([]Attack, len(attacks)-1)
                    copy(new_attacks[:i], attacks[:])
                    if len(attacks) > int(i+1) {
                        copy(new_attacks[i:], attacks[i+1:])
                    }
                    attacks = new_attacks

                    attacks[index] = Attack{
                                region_from: attacks[index].region_from,
                                region_to: attacks[index].region_to,
                                armies: attacks[index].armies + attack.armies }
                } else {
                    regionIdToIndex[key] = i
                    i++
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

func ourRegions(state *State) []*Region {
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

func ourBorderRegions(state *State) []*Region {
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

func ourBorderRegionsWithTheEnemy(state *State) []*Region {
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

func ourBorderRegionsWithNeutralOnly(state *State) []*Region {
    result := []*Region{}

    // fmt.Fprintf(os.Stderr, "ourRegions: %d\n", len(state.regions))
    for _, region := range state.regions {
        // fmt.Fprintf(os.Stderr, "%d %s\n", region.id, region.owner)
        if region.owner == "us" {
            borders_enemy := false
            borders_neutral := false
            for _, neighbour := range region.neighbours {
                if neighbour.owner == "them" {
                    borders_enemy = true
                } else if neighbour.owner == "neutral" {
                    borders_neutral = true
                }
            }

            if !borders_enemy && borders_neutral {
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
    visible bool
}

type SuperRegion struct {
    id int64
    regions []*Region
    reward int64
}

func isSuperRegionPossiblyOwnedByTheEnemy(super_region *SuperRegion) bool {
    possibly_owned_by_the_enemy := true

    for _, region := range super_region.regions {
        if !region.visible {
            continue
        }
        if region.owner != "them" {
            possibly_owned_by_the_enemy = false
            break
        }
    }

    return possibly_owned_by_the_enemy
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

func calculateAttacks(state *State) []Attack {
    attacks := []Attack{}

    attacks = reinforcement_moves(state, attacks)

    var newly_captured map[int64]bool
    attacks, newly_captured = attack_moves(state, attacks)

    attacks = border_reinforcement_moves(state, attacks, newly_captured)

    return attacks
}