package main

func border_reinforcement_moves(state *State, attacks []Attack, newly_captured map[int64]bool) []Attack {
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