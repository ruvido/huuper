ssh fiber "cd dev/huuper/deploy && docker compose -f docker-compose.yml down"
rsync -avz --delete fiber:dev/huuper/shared/data/ data.fiber
ssh fiber "cd dev/huuper/deploy && docker compose -f docker-compose.yml up -d"
