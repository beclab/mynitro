#!/bin/bash
#

#cd nitro
#cd build
#cmake -DLLAMA_CUBLAS=ON ..
#make -j $(nproc)
#cd ..
#cd ..

#if [ ! -f "/usr/lib/x86_64-linux-gnu/libcuda.so.1" ]; then
#  cp /libcuda.so.1 /usr/lib/x86_64-linux-gnu/ && \
#  rm -f usr/lib/x86_64-linux-gnu/libcudart.so.12 && \
#  rm -f usr/lib/x86_64-linux-gnu/libcublas.so.12 && \
#  rm -f usr/lib/x86_64-linux-gnu/libcublasLt.so.12
#fi

# 检测 CPU 架构
ARCH=$(uname -m)

if [ "$ARCH" == "x86_64" ]; then
  LIBDIR="/usr/lib/x86_64-linux-gnu"
elif [ "$ARCH" == "aarch64" ]; then
  LIBDIR="/usr/lib/aarch64-linux-gnu"
else
  echo "Unsupported CPU architecture: $ARCH"
  exit 1
fi

# 复制 libcuda.so.1
if [ ! -f "$LIBDIR/libcuda.so.1" ]; then
  cp /libcuda.so.1 "$LIBDIR/" && \
  rm -f "$LIBDIR/libcudart.so.12" && \
  rm -f "$LIBDIR/libcublas.so.12" && \
  rm -f "$LIBDIR/libcublasLt.so.12"
fi

rm -f /libcuda.so.1

curl -sSf https://raw.githubusercontent.com/WasmEdge/WasmEdge/master/utils/install_v2.sh | bash -s
source $HOME/.bashrc

./mynitro
#tail -f /dev/null