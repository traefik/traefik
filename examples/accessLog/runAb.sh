#!/bin/bash
usage()
{
  echo 'runAb.sh - Run Apache Benchmark to test access log'
  echo '   Usage: runAb.sh [--conn nnn] [--log xxx] [--num nnn] [--time nnn] [--wait nn]'
  echo '     -c|--conn - number of simultaneous connections (default 100)'
  echo '     -l|--log  - name of logfile (default benchmark.log)'
  echo '     -n|--num  - number of requests (default 50000); ignored when -t specified'
  echo '     -t|--time - time in seconds for benchmark (default no limit)'
  echo '     -w|--wait - number of seconds to wait for Traefik to initialize (default 15)'
  echo '   '
  exit
}

# Parse options

conn=100
num=50000
wait=15
time=0
logfile=""
while [[ $1 =~ ^- ]]
do
  case $1 in
    -c|--conn)
      conn=$2
      shift
      ;;
    -h|--help)
      usage
      ;;
    -l|--log|--logfile)
      logfile=$2
      shift
      ;;
    -n|--num)
      num=$2
      shift
      ;;
    -t|--time)
      time=$2
      shift
      ;;
    -w|--wait)
      wait=$2
      shift
      ;;
    *)
      echo Unknown option "$1"
      usage
  esac
  shift
done
if [ -z "$logfile" ] ; then
  logfile="benchmark.log"
fi

# Change to accessLog examples directory

[ -d examples/accessLog ] && cd examples/accessLog
if [ ! -r exampleHandler.go ] ; then
  echo Please run this script either from the traefik repo root or from the examples/accessLog directory
  exit
fi

# Kill traefik and any running example processes

sudo pkill -f traefik
pkill -f exampleHandler
[ ! -d log ] && mkdir log

# Start new example processes

go build exampleHandler.go
[ $? -ne 0 ] && exit $?
./exampleHandler -n Handler1 -p 8081 &
[ $? -ne 0 ] && exit $?
./exampleHandler -n Handler2 -p 8082 &
[ $? -ne 0 ] && exit $?
./exampleHandler -n Handler3 -p 8083 &
[ $? -ne 0 ] && exit $?

# Wait a couple of seconds for handlers to initialize and start Traefik

cd ../..
sleep 2s
echo Starting Traefik...
sudo ./traefik -c examples/accessLog/traefik.ab.toml &
[ $? -ne 0 ] && exit $?

# Wait for Traefik to initialize and run ab

echo Waiting $wait seconds before starting ab benchmark
sleep ${wait}s
echo
stime=`date '+%s'`
if [ $time -eq 0 ] ; then
  echo Benchmark starting `date` with $conn connections until $num requests processed | tee $logfile
  echo | tee -a $logfile
  echo ab -k -c $conn -n $num http://127.0.0.1/test | tee -a $logfile
  echo | tee -a $logfile
  ab -k -c $conn -n $num http://127.0.0.1/test 2>&1 | tee -a $logfile
else
  if [ $num -ne 50000 ] ; then
    echo Request count ignored when --time specified
  fi
  echo Benchmark starting `date` with $conn connections for $time seconds | tee $logfile
  echo | tee -a $logfile
  echo ab -k -c $conn -t $time -n 100000000 http://127.0.0.1/test | tee -a $logfile
  echo | tee -a $logfile
  ab -k -c $conn -t $time -n 100000000 http://127.0.0.1/test 2>&1 | tee -a $logfile
fi

etime=`date '+%s'`
let "dt=$etime - $stime"
let "ds=$dt % 60"
let "dm=($dt / 60) % 60"
let "dh=$dt / 3600"
echo | tee -a $logfile
printf "Benchmark ended `date` after %d:%02d:%02d\n" $dh $dm $ds | tee -a $logfile
echo Results available in $logfile

