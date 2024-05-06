#!/bin/bash
#

#cd nitro
#cd build
#cmake -DLLAMA_CUBLAS=ON ..
#make -j $(nproc)
#cd ..
#cd ..

curl -sSf https://raw.githubusercontent.com/WasmEdge/WasmEdge/master/utils/install_v2.sh | bash -s
source $HOME/.bashrc

./mynitro
#tail -f /dev/null