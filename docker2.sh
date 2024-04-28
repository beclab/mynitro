#!/bin/bash
#

ls
mkdir build
cd build
cmake .. -DLLAMA_CUBLAS=ON -DLLAMA_AVX=OFF -DLLAMA_AVX2=OFF -DLLAMA_F16C=OFF -DLLAMA_FMA=OFF -DLLAMA_AVX512=OFF -DLLAMA_SSE=OFF -DLLAMA_SSSE=OFF -DLLAMA_VSX=OFF
cmake --build . --config Release
cd ..

build/bin/server -m models/3b_v1.1_gguf/3b_ggml-model-q4_0.gguf --port 8080 --host 0.0.0.0 -ngl $NGL_VALUE -c $C_VALUE $OTHER_VALUES
#build/bin/server -m models/llama-2-7b-chat.Q4_K_M.gguf --port 8080 --host 0.0.0.0 -ngl $NGL_VALUE -c $C_VALUE $OTHER_VALUES
#build/bin/server -m models/llama-2-13b-chat.Q4_K_M.gguf --port 8080 --host 0.0.0.0 -ngl $NGL_VALUE -c $C_VALUE $OTHER_VALUES