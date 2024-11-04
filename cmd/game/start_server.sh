go build -o go_gs_game main.go
export GODEBUG=gctrace=1
nohup ./go_gs_game -c ../../configs/config.toml > data.log 2>&1 &
