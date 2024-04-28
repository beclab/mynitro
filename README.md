# MyNitro

## Description
**MyNitro** is a wrapper for the official **Nitro**, while also implementing its own model downloading and management system (including a separate webpage).

**MyNitro** is designed to provide LLM (Large Language Model) services to **dify**'s agents for their work.

About Nitro：https://github.com/janhq/nitro

## Getting Started

**MyNitro** can be run locally, provided that **Nitro** has been downloaded, compiled, and installed on the local machine in advance.

For specific steps, please refer to the following:

[Dockerfile_nitro](Dockerfile_nitro)

### Clone the repository
```shell
git clone --recursive https://github.com/beclab/mynitro.git
```
### Build
```shell
cd nitro
go mod tidy
go build -o mynitro
```
### Set OS Environments
```shell
export NGL_VALUE=33
export C_VALUE=4096
export OTHER_VALUES=
```
### Run
```shell
./mynitro
```
