go build -o go_gs_gate main.go
nohup ./go_gs_gate -c ../../configs/config.toml > data.log 2>&1 &
