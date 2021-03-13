#!/bin/bash
# https://stackoverflow.com/questions/394230/how-to-detect-the-os-from-a-bash-script

# find exited container
docker run --rm -w "/builder" -v "${PWD}:/builder" heroiclabs/nakama-pluginbuilder:3.1.1 build -buildmode=plugin -trimpath -o ./nakama/nakama/data/modules/plugin_code.so
