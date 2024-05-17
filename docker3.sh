#!/bin/bash
#

#cd nitro
#cd build
#cmake -DLLAMA_CUBLAS=ON ..
#make -j $(nproc)
#cd ..
#cd ..

if [ ! -f "/usr/lib/x86_64-linux-gnu/libcuda.so.1" ]; then
  cp /libcuda.so.1 /usr/lib/x86_64-linux-gnu/ && \
  rm -f usr/lib/x86_64-linux-gnu/libcudart.so.12 && \
  rm -f usr/lib/x86_64-linux-gnu/libcublas.so.12 && \
  rm -f usr/lib/x86_64-linux-gnu/libcublasLt.so.12
fi

rm -f /libcuda.so.1

curl -sSf https://raw.githubusercontent.com/WasmEdge/WasmEdge/master/utils/install_v2.sh | bash -s
source $HOME/.bashrc

./mynitro
#tail -f /dev/null