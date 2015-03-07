package main

func calculateAttacks(state *State) []Attack {
    attacks := []Attack{}

    attacks = reinforcement_moves(state, attacks)

    var newly_captured map[int64]bool
    attacks, newly_captured = attack_moves(state, attacks)

    attacks = border_reinforcement_moves(state, attacks, newly_captured)

    return attacks
}

