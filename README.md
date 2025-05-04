
# C/C++ SDK

## Pico SDK

    git clone https://github.com/raspberrypi/pico-sdk.git
    cd pico-sdk/
    git submodule update --init
    export PICO_SDK_PATH=~/Repos/pico-sdk/

## Install `pioasm` tool

    cd ~/Repos/pico-sdk/tools/pioasm/
    mkdir build
    cd build/
    cmake ..
    make -j
    sudo make install

## Picotool

    git clone https://github.com/raspberrypi/picotool.git
    cd picotool/
    mkdir build
    cd build/
    cmake ..
    make -j
    sudo make install

    cd ..
    sudo cp udev/99-picotool.rules /etc/udev/rules.d/


## Pico examples

    git clone https://github.com/raspberrypi/pico-examples.git
    mkdir build
    cd build/
    cmake ..
    make -j
