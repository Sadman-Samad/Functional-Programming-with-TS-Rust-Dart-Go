const value: number[] = [2, 3, 4];

function main() {
  let value1 = [...value, 3];
  value1.push(4);

  return value1;
}

console.log(main());

export {};
