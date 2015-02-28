package main

import "os"
import "fmt"
import "bufio"

func main() {
    reader := bufio.NewReader(os.Stdin)

    for {
        line, err := reader.ReadString('\n')

        if err != nil {
            break
        }

        fmt.Printf("<< "+line)
    }
}