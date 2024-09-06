package t

import (
	"endtner.dev/nChess/internal/engine"
	"endtner.dev/nChess/internal/utils"
	"fmt"
	"strings"
	"testing"
	"time"
)

/*
	These tests can NOT run in parallel without interfering with the move generator.
*/

func doPerftTest(positionName string, positionFen string, expectedPerftResults []int64) bool {
	testingResult := true

	b := utils.FromFen(positionFen)

	resultString := strings.Builder{}

	for i := range len(expectedPerftResults) {

		start := time.Now()
		perftResult := engine.Perft(b, i, -1)

		resultString.WriteString(fmt.Sprintf("[%s] Perft(%d) %d, (Expected %d), match=%t, runtime=%s\n", positionName, i, perftResult, expectedPerftResults[i], perftResult == expectedPerftResults[i], time.Since(start)))
		if expectedPerftResults[i] != perftResult {
			testingResult = false
		}
	}

	fmt.Println(resultString.String())

	return testingResult
}

/*
	Running tests for perft position 1-6 in parallel. Also added the option to easily run more perft tests later on.
*/

func TestPerftInitialPosition(t *testing.T) {
	positionName := "Initial Position"
	positionFen := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	expectedPerftResults := []int64{1, 20, 400, 8902, 197281, 4865609, 119060324}

	if !doPerftTest(positionName, positionFen, expectedPerftResults) {
		t.Errorf("[%s] Testing Failed", positionName)
	}
}

func TestPerftPosition2(t *testing.T) {
	positionName := "Position 2"
	positionFen := "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1"
	expectedPerftResults := []int64{1, 48, 2039, 97862, 4085603, 193690690}

	if !doPerftTest(positionName, positionFen, expectedPerftResults) {
		t.Errorf("[%s] Testing Failed", positionName)
	}
}

func TestPerftPosition3(t *testing.T) {
	positionName := "Position 3"
	positionFen := "8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1"
	expectedPerftResults := []int64{1, 14, 191, 2812, 43238, 674624, 11030083}

	if !doPerftTest(positionName, positionFen, expectedPerftResults) {
		t.Errorf("[%s] Testing Failed", positionName)
	}
}

func TestPerftPosition4(t *testing.T) {
	positionName := "Position 4"
	positionFen := "r3k2r/Pppp1ppp/1b3nbN/nP6/BBP1P3/q4N2/Pp1P2PP/R2Q1RK1 w kq - 0 1"
	expectedPerftResults := []int64{1, 6, 264, 9467, 422333, 15833292}

	if !doPerftTest(positionName, positionFen, expectedPerftResults) {
		t.Errorf("[%s] Testing Failed", positionName)
	}
}

func TestPerftPosition5(t *testing.T) {
	positionName := "Position 5"
	positionFen := "rnbq1k1r/pp1Pbppp/2p5/8/2B5/8/PPP1NnPP/RNBQK2R w KQ - 1 8"
	expectedPerftResults := []int64{1, 44, 1486, 62379, 2103487, 89941194}

	if !doPerftTest(positionName, positionFen, expectedPerftResults) {
		t.Errorf("[%s] Testing Failed", positionName)
	}
}

func TestPerftPosition6(t *testing.T) {
	positionName := "Position 6"
	positionFen := "r4rk1/1pp1qppp/p1np1n2/2b1p1B1/2B1P1b1/P1NP1N2/1PP1QPPP/R4RK1 w - - 0 10"
	expectedPerftResults := []int64{1, 46, 2079, 89890, 3894594, 164075551} //, 6923051137}

	if !doPerftTest(positionName, positionFen, expectedPerftResults) {
		t.Errorf("[%s] Testing Failed", positionName)
	}
}
