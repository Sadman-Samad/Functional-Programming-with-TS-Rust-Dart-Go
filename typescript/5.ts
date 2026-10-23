const value: number[] = [4, 5, 6, 7];


function main() {
  let v = value.map((v) => v * 3);
  console.log(v);
  let m = value.filter((v) => v == 4);
  console.log(m);
  let fnd = value.find((v) => v == 7);
  console.log(fnd);
  let reduce = value.reduce((v, n) => v * n + 2, 0);
  console.log("kd", reduce);
  return 0

}

main();
