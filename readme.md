
## build nbastats
```
export GOPATH=$(pwd)
go build nbastats
```

## download data for a specific season
```
./download_season.sh 2015-16
./download_season.sh 2016-17
```

## run a report
```
./nbastats -o output.csv -s 2015-16,2016-17
```

