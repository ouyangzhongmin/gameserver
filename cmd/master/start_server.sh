go build -o go_gs_master main.go
nohup ./go_gs_master -c ../../configs/config.toml > data.log 2>&1 &
