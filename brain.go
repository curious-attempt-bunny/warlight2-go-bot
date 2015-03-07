package main

func calculateAttacks(state *State) []Attack {
    attacks := []Attack{}

    // --------- reinforcements

    attacks = reinforcements(state, attacks)

    // --------- attacks

    superRegionOwnedByEnemy := make(map[int64]bool)
    for _, super_region := range state.super_regions {
        superRegionOwnedByEnemy[super_region.id] = isSuperRegionPossiblyOwnedByTheEnemy(super_region)
        // fmt.Fprintf(os.Stderr, "Super region %d possibly owned by the enemy? %t\n",
        //     super_region.id, superRegionOwnedByEnemy[super_region.id])
    }

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
                            if attack_to == nil ||
                                    (superRegionOwnedByEnemy[attack_to.super_region.id] == superRegionOwnedByEnemy[region.super_region.id] &&
                                        !superRegionOwnedByEnemy[attack_to.super_region.id] ||
                                        attack_to.super_region.reward <= region.super_region.reward) ||
                                    superRegionOwnedByEnemy[region.super_region.id] {
                                attack_to = neighbour
                                attack_from = region
                                value = candidate_value
                            }
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
            // fmt.Fprintf(os.Stderr, "Region %d now has %d armies\n", attack_from.id, attack_from.armies)
            attack_to.owner = "us" // let's assume
            newly_captured[attack_to.id] = true
            used[attack_to.id] = true
        }
    }

    // --------- reinforcements along the neutral border

    ideal_destinations := ourBorderRegionsWithTheEnemy(state)
    regionToIdeal := make(map[int64]bool)
    for _, region := range ideal_destinations {
        regionToIdeal[region.id] = true
    }

    move_from := ourBorderRegionsWithNeutralOnly(state)
    armiesPlaced := make(map[int64]int64)
    for _, region := range move_from {
        armiesPlaced[region.id] = region.armies
    }

    for _, region := range move_from {
        if region.armies == 1 {
            continue
        }

        if newly_captured[region.id] {
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

            if best_destination == nil ||
                (armiesPlaced[neighbour.id] >= region.armies &&
                    regionToIdeal[neighbour.id] == regionToIdeal[region.id]) ||
                (regionToIdeal[neighbour.id] && !regionToIdeal[region.id]) {
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

    return attacks
}

func reinforcements(state *State, attacks []Attack) []Attack {
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

    return attacks
}