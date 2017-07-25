REQ=262144
CLIENTS=64
ADDR=http://127.0.0.1:8500/v1/kv/bench
DATA="74a31e96-1d0f-4fa7-aa14-7212a326986e"
MAXPROCS=4

all: put get-default get-stale get-consistent

put:
	@echo "===== PUT test ====="
	GOMAXPROCS=${MAXPROCS} boom -m PUT -d ${DATA} -n ${REQ} -c ${CLIENTS} ${ADDR}

get-default:
	@echo "===== GET default test ====="
	GOMAXPROCS=${MAXPROCS} boom -n ${REQ} -c ${CLIENTS} ${ADDR}

get-stale:
	@echo "===== GET stale test ====="
	GOMAXPROCS=${MAXPROCS} boom -n ${REQ} -c ${CLIENTS} ${ADDR}?stale

get-consistent:
	@echo "===== GET consistent test ====="
	GOMAXPROCS=${MAXPROCS} boom -n ${REQ} -c ${CLIENTS} ${ADDR}?consistent

