
SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" &>/dev/null && pwd -P)

# turn down application
cd $SCRIPT_DIR/../
docker-compose down

# rebuild images
$SCRIPT_DIR/build-all-images.sh

# turn on application
docker-compose up -d

# turn on application
docker exec -it auctions-service /bin/bash
./main