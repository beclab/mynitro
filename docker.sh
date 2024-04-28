#!/bin/bash
#

cmake -DLLAMA_CUBLAS=ON ..
make -j $(nproc)
nohup ./nitro 1 0.0.0.0 3928 > nitro.log 2>&1 &
sleep 5
curl http://localhost:3928/inferences/llamacpp/loadmodel \
  -H 'Content-Type: application/json' \
  -d '{
    "llama_model_path": "model/llama-2-7b-chat.Q4_K_M.gguf",
    "ctx_len": '${C_VALUE}',
    "ngl": '${NGL_VALUE}'
  }'
tail -f /dev/null

#./nitro 1 0.0.0.0 3928 &
#sleep 5
#
#ls
#curl http://localhost:3928/inferences/llamacpp/loadmodel \
#  -H 'Content-Type: application/json' \
#  -d '{
#    "llama_model_path": "model/3b_ggml-model-q4_0.gguf",
#    "ctx_len": 512,
#    "ngl": 100,
#  }'
