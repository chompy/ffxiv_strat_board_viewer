import {spriteParameters, StrategyBoardObject} from './ffxiv-strategy-board-viewer/objects.ts';
console.log(JSON.stringify(
Object.entries(spriteParameters).map(([key, value]) => ({
    id: parseInt(key),
    name: StrategyBoardObject[key as keyof typeof StrategyBoardObject].toString(),
    ...value
}))))