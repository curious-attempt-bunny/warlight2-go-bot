package main

import "os"
import "fmt"
import "bufio"
import "strings"

func main() {
    reader := bufio.NewReader(os.Stdin)

    for {
        line, err := reader.ReadString('\n')
        // our_name := ""
        // their_name := ""

        if err != nil {
            break
        }

        line = strings.TrimSpace(line)

        parts := strings.Split(line, " ")

        if len(line) == 0 {

        } else if strings.Index(line, "#") == 0 {

        } else if parts[0] == "settings" {
            if parts[1] == "your_bot" {
                // our_name = parts[2]
            } else if parts[1] == "opponent_bot" {
                // their_name = parts[2]
            }
        } else if parts[0] == "setup_map" {

        } else if parts[0] == "pick_starting_region" {
            // pick the first one
            fmt.Println(parts[2])
        } else if parts[0] == "update_map" {

        } else if parts[0] == "opponent_moves" {

        } else if parts[0] == "Round" {

        } else if strings.Index(line, "Output from your bot: ") == 0 {

        } else if strings.Index(line, "go place_armies") == 0 {
            fmt.Println("No moves")
        } else if strings.Index(line, "go attack/transfer") == 0 {
            fmt.Println("No moves")
        } else {
            fmt.Fprintf(os.Stderr, "Don't recognire: "+line+"\n")
        }
    }
}