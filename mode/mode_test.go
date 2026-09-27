package mode

import "testing"

const (
	NAME_UNKNOWN = "unknown"
	ID_UNKNOWN   = 0
)

func TestLookup(test *testing.T) {
	for _, params := range MODES {
		byName, isName := ByName(params.Name)
		byId, isId := ById(params.Id)
		isSame := isName && isId &&
			byName == params && byId == params
		if !isSame {
			test.Fatalf("режим %s", params.Name)
		}
	}
}

func TestLookupUnknown(test *testing.T) {
	_, isName := ByName(NAME_UNKNOWN)
	_, isId := ById(ID_UNKNOWN)
	if isName || isId {
		test.Fatalf("найден несуществующий режим")
	}
}
