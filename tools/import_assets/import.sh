#!/usr/bin/env bash

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
OBJECTS_TS_URL="https://raw.githubusercontent.com/Ennea/ffxiv-strategy-board-viewer/refs/heads/master/objects.ts"

echo """import {spriteParameters, StrategyBoardObject} from './object.ts';
console.log(JSON.stringify(
Object.entries(spriteParameters).map(([key, value]) => ({
    id: parseInt(key),
    name: StrategyBoardObject[key as keyof typeof StrategyBoardObject].toString(),
    ...value
}))))""" > $SCRIPT_DIR/extract.ts

curl "$OBJECTS_TS_URL" -o "$SCRIPT_DIR/object.ts"
tsx $SCRIPT_DIR/extract.ts > $SCRIPT_DIR/../../assets/assets.json

rm extract.ts
rm object.ts