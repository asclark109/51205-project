# note this script's directory
SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" &>/dev/null && pwd -P)

# set env vars
. $SCRIPT_DIR/set-env-vars.sh

# turn down application
cd $SCRIPT_DIR/../
docker-compose down

# rebuild images
. $SCRIPT_DIR/build-all-images.sh

# turn on application (first set env vars!)
docker-compose up -d

# turn on application
docker exec -it auctions-service /bin/bash
./main