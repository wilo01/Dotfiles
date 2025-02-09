#!/bin/bash

docker pull kalor62/recatest
# Need to run xhost + to allow to connect to DISPLAY
xhost +local:

create_directory_if_not_exist() {
    if [ ! -d "$1" ]; then
        mkdir -p "$1"
    fi
}

# echo -n "Enter Company Name: "
# read -r COMPANY

# echo -n "Enter Username: "
# read -r USERNAME

# echo -n "Enter Password: "
# read -rs PASSWORD

create_directory_if_not_exist "$HOME/recatest"
create_directory_if_not_exist "$HOME/recatest/tests"
create_directory_if_not_exist "$HOME/recatest/tests/assertion_images"
create_directory_if_not_exist "$HOME/recatest/tests/db_config"

docker run -it -v ~/recatest/tests:/usr/src/app/tests --net=host \
    -e DISPLAY \
    -e COMPANY="" \
    -e USERNAME="" \
    -e PASSWORD="" \
    kalor62/recatest

xhost -local:
