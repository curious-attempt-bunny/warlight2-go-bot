package main

import "os"
import "fmt"
import "bufio"
import "strings"

func main() {
    reader := bufio.NewReader(os.Stdin)

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

        } else if parts[0] == "setup_map" {

        } else if parts[0] == "pick_starting_region" {

        } else if parts[0] == "update_map" {

        } else if parts[0] == "opponent_moves" {

        } else if parts[0] == "Round" {

        } else if strings.Index(line, "Output from your bot: ") == 0 {

        } else if strings.Index(line, "go place_armies") == 0 {

        } else if strings.Index(line, "go attack/transfer") == 0 {

        } else {
            fmt.Fprintf(os.Stderr, "Don't recognire: "+line+"\n")
        }
    }
}