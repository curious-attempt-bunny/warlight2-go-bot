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
                value := float64(region.super_region.reward) / weighting

                // fmt.Fprintf(os.Stderr, "Starting region %d because score %g (reward %d armies %d weighting %d score %d / %d)\n",
                //     region.id, value, region.super_region.reward, armies, weighting, region.super_region.reward, weighting)
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

                    score := float64(border_region.super_region.reward)

                    cost := float64(neutral_armies_in_super_region + 3*enemy_armies_in_super_region)
                    cost = math.Pow(cost, 1.01) // inflate larger costs
                    score = score / cost

                    if region == nil || score > best_score {
                        region = border_region
                        best_score = score
                        // fmt.Fprintf(os.Stderr, "Place armies at %d has score %g (%d / %d+3*%d)\n",
                        //     border_region.id, score,
                        //     border_region.super_region.reward,
                        //     neutral_armies_in_super_region, enemy_armies_in_super_region)
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
            newly_captured := make(map[int64]bool)
            for {
                // fmt.Fprintf(os.Stderr, "Attacks:\n")

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
                                    (region.armies > 3 && region.armies >= 2*neighbour.armies)
                                if bordering_enemy {
                                    attackable = false // never attack neutral when enemy is present
                                }
                            } else if neighbour.owner == "them" {
                                if neighbour.armies < largest_enemy {
                                    attackable = false // only attack the largest neighbouring army group
                                }
                            }

                            neighbour_bordering_enemy := false
                            for _, n := range neighbour.neighbours {
                                if n.owner == "them" {
                                    neighbour_bordering_enemy = true
                                    break
                                }
                            }

                            if neighbour_bordering_enemy {
                                candidate_value -= 0.001 // tie breaker
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
                    armies_to_attack_with := attack_from.armies-1

                    if attack_to.owner == "neutral" {
                        armies_to_attack_with = (attack_to.armies * 2) - 1
                        if armies_to_attack_with < 2 {
                            armies_to_attack_with = 2
                        }
                    } else {
                        max_enemy_placement := int64(5) // incorrect!
                        armies_to_needed := ((attack_to.armies + max_enemy_placement) * 2) - 1 + 2
                        if armies_to_needed + 3 < armies_to_attack_with {
                            // if we would end up with enough armies to attack a neutral with
                            armies_to_attack_with = armies_to_needed
                        }
                    }

                    neighbours_not_us := 0

                    for _, neighbour := range attack_from.neighbours {
                        if neighbour.owner != "us" {
                            neighbours_not_us += 1
                        }
                    }

                    // fmt.Fprintf(os.Stderr, "Attacking %d from %d (has %d armies) we have %d non-us neighbours.\n",
                    //     attack_to.id, attack_from.id, attack_from.armies, neighbours_not_us)
                    if neighbours_not_us == 1 {
                        armies_to_attack_with = attack_from.armies - 1
                    }

                    attacks = append(attacks, Attack{
                        region_from: attack_from,
                        region_to: attack_to,
                        armies: armies_to_attack_with})
                    attack_from.armies -= armies_to_attack_with
                    attack_to.owner = "us" // let's assume
                    newly_captured[attack_to.id] = true
                    used[attack_to.id] = true
                }
            }

            // --------- reinforcements along the neutral border

            move_from := ourBorderRegionsWithNeutralOnly(state)
            armiesPlaced := make(map[int64]int64)
            for _, region := range move_from {
                armiesPlaced[region.id] = region.armies
            }

            for _, region := range move_from {
                if region.armies == 1 {
                    continue
                }

                if armiesPlaced[region.id] > region.armies {
                    continue // don't move armies if we've been reinforced
                }

                var best_destination *Region
                best_destination = nil

                for _, neighbour := range region.neighbours {
                    if neighbour.owner != "us" {
                        continue
                    }

                    if best_destination == nil || armiesPlaced[neighbour.id] >= region.armies {
                        borders_something := false
                        for _, n := range neighbour.neighbours {
                            if n.owner != "us" {
                                borders_something = true
                                break
                            }
                        }

                        if borders_something {
                            best_destination = neighbour
                        }
                    }
                }

                if best_destination != nil {
                    armies_to_move := region.armies - 1
                    safe_to_prepend := (newly_captured[best_destination.id] != true)

                    attack := Attack{
                                region_from: region,
                                region_to: best_destination,
                                armies: armies_to_move }
                    if safe_to_prepend {
                        new_attacks := make([]Attack, len(attacks)+1)
                        copy(new_attacks[1:], attacks[:])
                        new_attacks[0] = attack
                        attacks = new_attacks
                    } else {
                        attacks = append(attacks, attack)
                    }

                    region.armies -= armies_to_move
                    armiesPlaced[best_destination.id] += armies_to_move

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

func ourBorderRegionsWithNeutralOnly(state State) []*Region {
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