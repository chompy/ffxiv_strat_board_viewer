#!/usr/bin/env bash

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
GIT_URL="https://github.com/Ennea/ffxiv-strategy-board-viewer.git"

# clone source repo to fetch assets
mkdir -p $SCRIPT_DIR/data
git clone $GIT_URL $SCRIPT_DIR/data/ffxiv-strategy-board-viewer

# make assets.json
echo """import {spriteParameters, StrategyBoardObject} from './ffxiv-strategy-board-viewer/objects.ts';
console.log(JSON.stringify(
Object.entries(spriteParameters).map(([key, value]) => ({
    id: parseInt(key),
    name: StrategyBoardObject[key as keyof typeof StrategyBoardObject].toString(),
    ...value
}))))""" > $SCRIPT_DIR/data/extract.ts
tsx $SCRIPT_DIR/data/extract.ts > $SCRIPT_DIR/data/assets.json
rm $SCRIPT_DIR/data/extract.ts

go run $SCRIPT_DIR