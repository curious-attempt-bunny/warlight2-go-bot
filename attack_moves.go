package main

func attack_moves(state *State, attacks []Attack) ([]Attack, map[int64]bool) {
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

    return attacks, newly_captured
}