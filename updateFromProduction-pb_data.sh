ssh fiber "cd dev/huuper && ./deploy.sh down"
rsync -avz --delete fiber:dev/huuper/pb_data pb_data.fiber
ssh fiber "cd dev/huuper && ./deploy.sh up"
