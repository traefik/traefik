BIN := scheduler executor zkdetect
RACE := -race

.PHONY:
.PHONY: all $(BIN)

all: $(BIN)

$(BIN):
	mkdir -p _output
	godep go build $(RACE) -o _output/$@ ./$@
