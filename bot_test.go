package main

import "testing"
import "os"
import "bytes"
import "io"
import "strings"
import "io/ioutil"

func TestBot(t *testing.T) {
  test(t, "tests/Attacks.txt")
  test(t, "tests/AttackNeutralWhenItWouldWin.txt")
  test(t, "tests/PlaceNextToTheEnemyFirst.txt")
  test(t, "tests/DontCrashPickingStartingRegions.txt")
  test(t, "tests/TakeSuperRegionsWhenNoBorder.txt")
  test(t, "tests/ReinforceToBorderWIthMostEnemyNeighbours.txt")
  test(t, "tests/OnlyAttackIfWeCanKillTheBiggestEnemyArmy.txt")
  test(t, "tests/AttackNeutralWithTheRequiredForceToWin.txt")
  test(t, "tests/MoveUnitsAroundOnTheNeutralBorder.txt")
  test(t, "tests/WithMultipleTargetsLimitEnemyAttackToLargestForcePossiblyNeeded.txt")
  test(t, "tests/ReinforceInCorrectOrderForSafely.txt")
  test(t, "tests/AvoidContactWithTheEnemyWhenClaimingARegion.txt")
  test(t, "tests/PlaceArmiesForMaximumIncome.txt")
  test(t, "tests/AttackWithAllArmiesWhenItsTheLastTarget2.txt")
  test(t, "tests/PickSuperRegionsBasedOnShapeAlso.txt")
}

func test(t *testing.T, filename string) {
  actual := actualOutput(filename)
  expected := expectedOutput(filename)

  valid := false

  for _, expectation := range expected {
    valid = match(actual, expectation)
    if valid {
      break
    }
  }

  if !valid {
    t.Errorf("%s: Actual: %s. Expected: %s", filename, actual, strings.Join(expected, " OR "))
  }
}

func match(actual string, expectation string) bool {
  if strings.Index(expectation, "!") == 0 {
    return !match(actual, expectation[1:])
  } else if strings.Index(expectation, "[") == 0 {
    return strings.Index(actual, expectation[1:len(expectation)-1]) != -1
  } else {
    return actual == expectation
  }
}

func actualOutput(filename string) string {
  original_stdin, original_stdout := os.Stdin, os.Stdout

  os.Stdin, _ = os.Open(filename)

  r, w, _ := os.Pipe()
  os.Stdout = w
  outC := make(chan string)

  go func() {
      var buf bytes.Buffer
      io.Copy(&buf, r)
      outC <- buf.String()
  }()

  main()

  w.Close()
  output := <-outC

  os.Stdout = original_stdout
  os.Stdin = original_stdin

  lines := strings.Split(strings.TrimSpace(output), "\n")
  actual := lines[len(lines)-1]

  return actual
}

func expectedOutput(filename string) []string {
  raw, _ := ioutil.ReadFile(filename)
  content := strings.TrimSpace(string(raw))
  lines := strings.Split(content, "\n")

  expected := []string{}

  for _, line := range lines {
    if strings.Index(line, "# Valid: ") == 0 {
      expectation := line[len("# Valid: "):]
      expected = append(expected, expectation)
    }
  }

  return expected
}