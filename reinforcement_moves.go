package main

func reinforcement_moves(state *State, attacks []Attack) []Attack {
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
